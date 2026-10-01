#!/usr/bin/env python3
"""Opt-in disposable Compose rehearsal with unique projects; never reuses live data."""
from __future__ import annotations

import os
from pathlib import Path
import shutil
import socket
import subprocess
import sys
import tempfile
import uuid

from environment_config import Environment, EnvironmentError, ROOT, CORE
from deployment_checks import check_runtime
import importlib.util


def free_port() -> int:
    with socket.socket() as listener:
        listener.bind(('127.0.0.1', 0))
        return listener.getsockname()[1]


def main() -> int:
    parser = __import__('argparse').ArgumentParser(description=__doc__)
    parser.add_argument('--run', action='store_true', help='Build an image, start disposable services and interrupt disposable RustFS; clean up afterward')
    args = parser.parse_args()
    if not args.run:
        parser.error('use --run explicitly; this rehearsal needs Docker, network/build access and spare host capacity')
    identifier = uuid.uuid4().hex[:12]
    image = 'iwa-environments-rehearsal:' + identifier
    projects = []
    temporary = tempfile.TemporaryDirectory(prefix='iwa-environments-rehearsal-')
    directory = temporary.name
    try:
        root = Path(directory)
        # Include reviewed Git files and current nonignored implementation,
        # never local settings, captures, credentials or machine indexes.
        files = subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard', '-z'], cwd=ROOT).decode().split('\0')
        for name in filter(None, files):
            source = ROOT / name
            if source.is_file():
                destination = root / name
                destination.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(source, destination)
        subprocess.run(['git', 'init', '-q'], cwd=root, check=True)
        subprocess.run(['git', 'add', '.'], cwd=root, check=True)
        subprocess.run(['git', '-c', 'user.name=Fixture', '-c', 'user.email=fixture@example.test', '-c', 'core.hooksPath=/dev/null', 'commit', '-qm', 'Synthetic environment rehearsal snapshot'], cwd=root, check=True)
        revision = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root).decode().strip()
        print('Building disposable application image from the fresh rehearsal snapshot.', flush=True)
        subprocess.run(['docker', 'build', '--build-arg', 'VCS_REF=' + revision, '--tag', image, str(root)], check=True)
        spec = importlib.util.spec_from_file_location('rehearsal_smoke', root / 'scripts/smoke-compose.py')
        smoke_module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(smoke_module)
        original_image = subprocess.check_output(['docker', 'image', 'inspect', '--format', '{{.Id}}', image]).strip()
        configs = []
        for name in ('development', 'staging', 'production'):
            (root / '.local' / name).mkdir(parents=True)
            contents = (root / f'deploy/{name}.env.example').read_text()
            if name != 'production':
                contents += f'\nIWA_APP_IMAGE={image}\nIWA_PUBLIC_PORT={free_port()}\nIWA_ADMIN_PORT={free_port()}\n'
            (root / f'.local/{name}.env').write_text(contents)
            old = os.umask(0o077)
            try:
                subprocess.run([sys.executable, str(root / 'scripts/init-secrets.py'), '--directory', str(root / f'.local/{name}/secrets')], check=True)
            finally:
                os.umask(old)
            selected = Environment(name, root=root)
            # Only the test harness assigns an isolated namespace. The public
            # wrapper still refuses project overrides. Never touch iwa* data.
            selected.project = 'iwa-rehearsal-' + identifier + '-' + name
            projects.append(selected)
            config = selected.config(all_services=True)
            configs.append(config)
            if name == 'production':
                assert not selected.config()['services']
                assert not selected.containers()
                continue
            print(f'Rehearsing {name} dependency, migration, storage and listener commands.', flush=True)
            selected.capture('up', '-d', '--wait', '--wait-timeout', '180', 'postgres', 'rustfs', 'crawl4ai')
            selected.capture('run', '--rm', '--no-deps', '--pull', 'never', 'admin', 'migrate')
            selected.capture('run', '--rm', '--no-deps', '--pull', 'never', 'admin', 'storage-init')
            selected.capture('up', '-d', '--no-build', '--pull', 'never', '--wait', '--wait-timeout', '180', 'public', 'admin')
            smoke_module.smoke(selected, allow_interruption=True)
            check_runtime(selected, config)
            assert subprocess.check_output(['docker', 'image', 'inspect', '--format', '{{.Id}}', image]).strip() == original_image
            selected.capture('stop')
            selected.capture('start', '--wait', '--wait-timeout', '180', *CORE)
            smoke_module.smoke(selected)
            selected.capture('stop')  # Bound concurrent memory use on the shared host.
        for volume in ('postgres_data', 'rustfs_data'):
            assert len({c['volumes'][volume]['name'] for c in configs}) == 3
        for secret in ('postgres_password', 'rustfs_access_key', 'rustfs_secret_key', 'crawl_token', 'public_cursor_key'):
            assert len({Path(c['secrets'][secret]['file']).read_bytes() for c in configs}) == 3
        print('PASS: disposable fresh-snapshot initialization, strict-umask mounted-secret readability, image-preserving staging, listener/volume/secret isolation, fault recovery, stop/start and inactive production. This is not a VM reboot or production readiness proof.', flush=True)
        # Clean up before removing files used by bind-mounted secrets.
        for selected in reversed(projects):
            selected.capture('--profile', '*', 'down', '-v', '--remove-orphans')
        projects.clear()
        return 0
    except (EnvironmentError, subprocess.CalledProcessError, OSError, AssertionError) as error:
        if isinstance(error, EnvironmentError) and isinstance(error.__cause__, subprocess.CalledProcessError):
            detail = error.__cause__.stderr.decode(errors='replace')
            for path in (root / '.local').glob('*/secrets/*'):
                if path.is_file():
                    value = path.read_text().strip()
                    if value:
                        detail = detail.replace(value, '<synthetic-secret>')
            print(detail, file=sys.stderr)
        print(str(error) if isinstance(error, EnvironmentError) else 'Disposable rehearsal failed; inspect its build/status output. Operational projects were not selected.', file=sys.stderr)
        return 1
    finally:
        for selected in reversed(projects):
            try:
                selected.capture('--profile', '*', 'down', '-v', '--remove-orphans')
            except (EnvironmentError, OSError):
                print('Rehearsal cleanup needs attention for ' + selected.project, file=sys.stderr)
        temporary.cleanup()
        subprocess.run(['docker', 'image', 'rm', image], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)


if __name__ == '__main__':
    sys.exit(main())
