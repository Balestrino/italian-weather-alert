#!/usr/bin/env python3
"""Regression tests use synthetic files/containers and never alter live projects."""
from __future__ import annotations

import copy
import importlib.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parent))
from environment_config import (APPLICATIONS, CORE, Environment, EnvironmentError,
                                inspect_boundaries, verify_staging_release)
from deployment_checks import check_runtime, templates

ROOT = Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location('smoke_compose', ROOT / 'scripts/smoke-compose.py')
smoke_module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(smoke_module)


class EnvironmentTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name) / 'download'
        (self.root / 'scripts').mkdir(parents=True)
        (self.root / 'deploy').mkdir()
        (self.root / '.local').mkdir()
        for name in ('compose-env.sh', 'environment_config.py', 'init-secrets.py', 'release-image.sh'):
            shutil.copy2(ROOT / 'scripts' / name, self.root / 'scripts' / name)
        shutil.copy2(ROOT / 'compose.yaml', self.root / 'compose.yaml')
        for path in (ROOT / 'deploy').glob('compose.*.yaml'):
            shutil.copy2(path, self.root / 'deploy' / path.name)
        for name in ('development', 'staging', 'production'):
            shutil.copy2(ROOT / f'deploy/{name}.env.example', self.root / f'.local/{name}.env')
        self.trace = Path(self.temp.name) / 'trace.jsonl'
        self.state = Path(self.temp.name) / 'state.json'
        self.bin = Path(self.temp.name) / 'bin'
        self.bin.mkdir()
        docker = self.bin / 'docker'
        docker.write_text('''#!/usr/bin/env python3
import json,os,sys
args=sys.argv[1:]
with open(os.environ['TEST_DOCKER_TRACE'],'a') as f:
    f.write(json.dumps({'args':args,'env':{k:v for k,v in os.environ.items() if k.startswith(('IWA_','COMPOSE_'))}})+'\\n')
state=json.load(open(os.environ['TEST_DOCKER_STATE']))
if 'config' in args:
    print(json.dumps(state['config']))
elif args[:2]==['image','inspect']:
    if '--format' in args:
        print(state['revision'])
    else:
        print(json.dumps([state['image']]))
elif args[:1]==['inspect']:
    print(json.dumps(state.get('containers',[])))
elif 'ps' in args:
    print('fixture-id' if state.get('containers') else '')
''')
        docker.chmod(0o755)
        self.config = {'name': 'iwa-production', 'services': {role: {'image': 'ghcr.io/example/iwa@sha256:' + 'a' * 64} for role in APPLICATIONS}}
        self.write_state()
        self.env = dict(os.environ, PATH=str(self.bin) + os.pathsep + os.environ['PATH'],
                        TEST_DOCKER_TRACE=str(self.trace), TEST_DOCKER_STATE=str(self.state),
                        IWA_PUBLIC_PORT='wrong', COMPOSE_PROJECT_NAME='wrong', COMPOSE_FILE='/missing')

    def write_state(self, **extra):
        self.state.write_text(json.dumps({'config': self.config, **extra}))

    def wrapper(self, *args):
        return subprocess.run([str(self.root / 'scripts/compose-env.sh'), *args],
                              cwd=self.temp.name, env=self.env, capture_output=True, text=True)

    def calls(self):
        return [json.loads(line) for line in self.trace.read_text().splitlines()] if self.trace.exists() else []

    def test_archive_root_fixed_files_and_inherited_overrides(self):
        for environment, project in [('development', 'iwa'), ('staging', 'iwa-staging'), ('production', 'iwa-production')]:
            with self.subTest(environment=environment):
                result = self.wrapper(environment, '--dry-run', '--profile=test', 'config', '--format', 'json')
                self.assertEqual(result.returncode, 0, result.stderr)
                call = self.calls()[-1]
                self.assertEqual(call['env'], {})
                args = call['args']
                self.assertEqual(args[args.index('-p') + 1], project)
                self.assertEqual(args[args.index('--project-directory') + 1], str(self.root))
                self.assertIn(str(self.root / 'compose.yaml'), args)
                if environment != 'development':
                    self.assertIn(str(self.root / f'deploy/compose.{environment}.yaml'), args)

    def test_identity_overrides_rejected_before_docker(self):
        for option in ('-p', '-piwa', '--project-name=iwa', '--project-name', '-f', '-fcompose.yaml', '--file=compose.yaml', '--env-file=x', '--project-directory=/tmp', '--unknown'):
            for position in ('before', 'after'):
                with self.subTest(option=option, position=position):
                    if self.trace.exists():
                        self.trace.unlink()
                    args = [option, 'iwa', 'config'] if position == 'before' else ['config', option, 'iwa']
                    if option == '--unknown' and position == 'after':
                        continue  # Docker subcommand validates its own ordinary flags.
                    result = self.wrapper('staging', *args)
                    self.assertNotEqual(result.returncode, 0)
                    self.assertEqual(self.calls(), [])

    def test_missing_option_value_fails_without_docker(self):
        result = self.wrapper('staging', '--profile')
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(self.calls(), [])

    def test_release_builds_rejected_and_development_can_build(self):
        for environment in ('staging', 'production'):
            for args in (['build', 'admin'], ['up', '--build', 'admin'], ['up', '--build=true', 'admin'], ['watch']):
                self.assertNotEqual(self.wrapper(environment, *args).returncode, 0)
        self.assertEqual(self.calls(), [])
        self.assertEqual(self.wrapper('development', 'build', 'admin').returncode, 0)

    def test_production_invalid_images_rejected_before_image_operation(self):
        for image in ('iwa-app:local', 'ghcr.io/example/iwa:latest', 'ghcr.io/example/iwa@sha256:REPLACE',
                      'ghcr.io/example/iwa@sha256:' + 'a' * 63, 'ghcr.io/example/iwa@sha256:' + 'g' * 64,
                      'other.example/iwa@sha256:' + 'a' * 64):
            for command in ('pull', 'up', 'run', 'create', 'start', 'restart'):
                with self.subTest(image=image, command=command):
                    self.trace.unlink(missing_ok=True)
                    for role in APPLICATIONS:
                        self.config['services'][role]['image'] = image
                    self.write_state()
                    self.assertNotEqual(self.wrapper('production', command, 'admin').returncode, 0)
                    self.assertTrue(all('config' in call['args'] for call in self.calls()))
        for command in ('config', 'ps', 'logs', 'stop', 'down'):
            self.assertEqual(self.wrapper('production', command).returncode, 0)

    def test_valid_production_digest_reaches_operation(self):
        self.assertEqual(self.wrapper('production', '--profile', 'production', 'up', '--no-build', 'admin').returncode, 0)
        self.assertIn('up', self.calls()[-1]['args'])

    def test_strict_umask_repeatability_and_custom_permissions(self):
        directory = self.root / '.local/development/secrets'
        for mask in (0o022, 0o077):
            old = os.umask(mask)
            try:
                subprocess.run([sys.executable, str(self.root / 'scripts/init-secrets.py'), '--directory', str(directory)], check=True, capture_output=True)
            finally:
                os.umask(old)
            self.assertEqual(directory.stat().st_mode & 0o777, 0o700)
            self.assertTrue(all(p.stat().st_mode & 0o777 == 0o644 for p in directory.iterdir()))
            values = {p.name: p.read_bytes() for p in directory.iterdir()}
            special = directory / 'postgres_password'
            special.chmod(0o640)
            subprocess.run([sys.executable, str(self.root / 'scripts/init-secrets.py'), '--directory', str(directory)], check=True, capture_output=True)
            self.assertEqual(values, {p.name: p.read_bytes() for p in directory.iterdir()})
            self.assertEqual(special.stat().st_mode & 0o777, 0o640)
            shutil.rmtree(directory)

    def test_release_from_archive_fails_with_helpful_prerequisite(self):
        result = subprocess.run([str(self.root / 'scripts/release-image.sh'), 'build'], cwd=self.temp.name, env=self.env, capture_output=True, text=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('Git checkout', result.stderr)
        self.assertEqual(self.calls(), [])

    def test_real_compose_from_archive_keeps_fixed_namespace(self):
        if shutil.which('docker') is None:
            self.skipTest('Docker Compose is needed for template parsing')
        for name, project in [('development', 'iwa'), ('staging', 'iwa-staging'), ('production', 'iwa-production')]:
            result = subprocess.run([str(self.root / 'scripts/compose-env.sh'), name, '--profile', '*', 'config', '--format', 'json'],
                                    cwd=self.temp.name, env=dict(os.environ, IWA_PUBLIC_PORT='wrong', COMPOSE_PROJECT_NAME='wrong', COMPOSE_FILE='/missing'), capture_output=True, text=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            config = json.loads(result.stdout)
            self.assertEqual(config['name'], project)
            self.assertEqual(config['volumes']['postgres_data']['name'], project + '_postgres_data')
            self.assertTrue(config['secrets']['postgres_password']['file'].startswith(str(self.root / f'.local/{name}/secrets')))
            if name != 'development':
                self.assertNotIn('build', config['services']['admin'])

    def test_publication_checks_staging_before_push(self):
        subprocess.run(['git', 'init', '-q'], cwd=self.root, check=True)
        (self.root / '.gitignore').write_text('.local/\n')
        subprocess.run(['git', 'add', '.'], cwd=self.root, check=True)
        subprocess.run(['git', '-c', 'user.name=Fixture', '-c', 'user.email=fixture@example.test', '-c', 'core.hooksPath=/dev/null', 'commit', '-qm', 'fixture'], cwd=self.root, check=True)
        revision = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=self.root).decode().strip()
        self.config = {'name': 'iwa-staging', 'services': {role: {'image': 'ghcr.io/balestrino/italian-weather-alert:' + revision, 'secrets': []} for role in CORE}}
        containers = []
        for role in CORE:
            ports = {}
            if role in ('public', 'admin'):
                port = '28080' if role == 'public' else '28081'
                self.config['services'][role]['ports'] = [{'host_ip': '127.0.0.1', 'published': port}]
                ports = {'8080/tcp': [{'HostIp': '127.0.0.1', 'HostPort': port}]}
            containers.append({'Config': {'Labels': {'com.docker.compose.project': 'iwa-staging', 'com.docker.compose.service': role}}, 'Image': 'sha256:tested', 'State': {'Running': True}, 'HostConfig': {'PortBindings': ports}, 'NetworkSettings': {'Ports': ports}, 'Mounts': []})
        candidate = {'Id': 'sha256:tested', 'Config': {'Labels': {'org.opencontainers.image.revision': revision}}}
        for mismatch in (True, False):
            self.trace.unlink(missing_ok=True)
            containers[0]['Image'] = 'sha256:other' if mismatch else 'sha256:tested'
            self.write_state(revision=revision, image=candidate, containers=containers)
            result = subprocess.run([str(self.root / 'scripts/release-image.sh'), 'publish'], cwd=self.temp.name, env=self.env, capture_output=True, text=True)
            pushed = any(call['args'][0] == 'push' for call in self.calls())
            self.assertEqual(pushed, not mismatch)
            if mismatch:
                self.assertNotEqual(result.returncode, 0)


class RuntimeTests(unittest.TestCase):
    def setUp(self):
        self.environment = type('SelectedEnvironment', (), {'name': 'staging', 'project': 'iwa-staging'})()
        self.config = {'services': {role: {'image': 'ghcr.io/example/iwa:revision', 'restart': 'unless-stopped', 'secrets': []} for role in CORE}, 'secrets': {}}
        self.containers = []
        for role in CORE:
            service = self.config['services'][role]
            bindings = {}
            if role in ('public', 'admin'):
                port = '28080' if role == 'public' else '28081'
                service['ports'] = [{'host_ip': '127.0.0.1', 'published': port}]
                bindings = {'8080/tcp': [{'HostIp': '127.0.0.1', 'HostPort': port}]}
            self.containers.append({'Config': {'Labels': {'com.docker.compose.project': 'iwa-staging', 'com.docker.compose.service': role}}, 'State': {'Running': True}, 'Image': 'sha256:candidate', 'HostConfig': {'PortBindings': bindings, 'RestartPolicy': {'Name': 'unless-stopped'}}, 'NetworkSettings': {'Ports': bindings}, 'Mounts': []})

    def test_runtime_wrong_project_or_actual_port_fails(self):
        for change in ('project', 'port', 'mount'):
            containers = copy.deepcopy(self.containers)
            config = copy.deepcopy(self.config)
            if change == 'project':
                containers[0]['Config']['Labels']['com.docker.compose.project'] = 'iwa'
            elif change == 'port':
                containers[0]['NetworkSettings']['Ports']['8080/tcp'][0]['HostIp'] = '0.0.0.0'
            else:
                config['services']['public']['secrets'] = [{'source': 'key', 'target': '/run/secrets/key'}]
                config['secrets'] = {'key': {'file': '/synthetic/staging/key'}}
                containers[0]['Mounts'] = [{'Source': '/synthetic/development/key', 'Destination': '/run/secrets/key', 'RW': False}]
            with self.assertRaises(EnvironmentError):
                inspect_boundaries(self.environment, config, containers)

    def test_smoke_default_is_read_only_and_unused_provider_key_is_optional(self):
        self.environment.config = lambda **kwargs: self.config
        self.environment.containers = lambda: self.containers
        calls = []
        self.environment.capture = lambda *args: calls.append(args) or b'fixture logs'
        with patch.object(smoke_module, 'validate_config'), patch.object(smoke_module, 'check_http', return_value=[b'{}']):
            smoke_module.smoke(self.environment)
        self.assertTrue(calls)
        self.assertTrue(all(call[0] == 'logs' for call in calls))

    def test_fault_assertion_failure_recovers_same_environment(self):
        calls = []
        self.environment.capture = lambda *args: calls.append(args) or b''
        with patch.object(smoke_module, 'request', return_value=(200, b'{}')), patch.object(smoke_module, 'ready'):
            with self.assertRaises(EnvironmentError):
                smoke_module.fault_check(self.environment, self.config)
        self.assertEqual(calls, [('stop', 'rustfs'), ('start', '--wait', '--wait-timeout', '180', 'rustfs')])

    def test_leaked_secret_fails_without_disclosing_value(self):
        self.environment.config = lambda **kwargs: self.config
        self.environment.containers = lambda: self.containers
        value = b'synthetic-secret-for-redaction-check'
        self.environment.capture = lambda *args: b'log contains ' + value
        with tempfile.TemporaryDirectory() as directory:
            secret = Path(directory) / 'secret'
            secret.write_bytes(value)
            with patch.object(smoke_module, 'validate_config'), patch.object(smoke_module, 'check_http', return_value=[b'{}']), patch.object(smoke_module, 'used_secret_paths', return_value={secret}):
                with self.assertRaises(EnvironmentError) as caught:
                    smoke_module.smoke(self.environment)
        self.assertNotIn(value.decode(), str(caught.exception))
        self.assertIn('value withheld', str(caught.exception))

    def test_production_interruption_rejected_before_resolving_config(self):
        self.environment.name = 'production'
        with self.assertRaises(EnvironmentError):
            smoke_module.smoke(self.environment, True)

    def test_release_checks_runtime_image_identity(self):
        self.environment.config = lambda **kwargs: self.config
        self.environment.containers = lambda: self.containers
        candidate = {'Id': 'sha256:candidate', 'Config': {'Labels': {'org.opencontainers.image.revision': 'revision'}}}
        with patch('environment_config.Environment', return_value=self.environment), patch('environment_config.docker_json', return_value=[candidate]):
            verify_staging_release('ghcr.io/example/iwa:revision', 'revision')
            self.containers[0]['Image'] = 'sha256:rebuilt'
            with self.assertRaises(EnvironmentError):
                verify_staging_release('ghcr.io/example/iwa:revision', 'revision')
            self.containers[0]['Image'] = 'sha256:candidate'
            candidate['Config']['Labels']['org.opencontainers.image.revision'] = 'old'
            with self.assertRaises(EnvironmentError):
                verify_staging_release('ghcr.io/example/iwa:revision', 'revision')


if __name__ == '__main__':
    unittest.main()
