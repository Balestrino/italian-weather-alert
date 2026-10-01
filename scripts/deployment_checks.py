#!/usr/bin/env python3
"""Validate templates or inspect one selected installation without changing it."""

from __future__ import annotations

import argparse
import json
import os
from pathlib import Path
import subprocess
import sys
import urllib.error
import urllib.request

from environment_config import (APPLICATIONS, CORE, ROOT, PROJECTS, DIGEST,
                                Environment, EnvironmentError, docker_json,
                                inspect_boundaries, listener_origin)


def validate_config(environment: Environment, config: dict) -> None:
    if config['name'] != environment.project:
        raise EnvironmentError('Resolved Compose project differs from the selected environment.')
    for name in ('public', 'admin'):
        listener_origin(config, name)
    for name, service in config['services'].items():
        if service.get('restart') != 'unless-stopped':
            raise EnvironmentError(f'{name} lacks the expected restart policy.')
        if name not in ('public', 'admin') and service.get('ports'):
            raise EnvironmentError(f'{name} publishes an internal host port.')
        if name in APPLICATIONS and environment.name != 'development' and service.get('build'):
            raise EnvironmentError('Release environments must not build application images.')
        if environment.name == 'production':
            if int(service.get('mem_limit', 0)) <= 0 or not 0 < service.get('cpus', 0) <= 1.5 or not 0 < service.get('pids_limit', 0) <= 512:
                raise EnvironmentError(f'Production {name} resource ceilings are missing or invalid.')
    for name, volume in config.get('volumes', {}).items():
        if volume['name'] != environment.project + '_' + name:
            raise EnvironmentError('Named volume differs from the selected project namespace.')


def validate_background_selection(environment: Environment, config: dict) -> None:
    """Resolve profiles without starting services or requiring provider keys."""
    production = environment.name == 'production'
    worker_profile = 'production-worker' if production else 'processing-worker'
    for service, profile in (('worker', worker_profile), ('backup', 'application-backup')):
        if config['services'].get(service, {}).get('profiles') != [profile]:
            raise EnvironmentError(f'{environment.name} {service} must use only the {profile} profile.')
    default = set() if production else set(CORE)
    core_options = ('--profile', 'production') if production else ()
    selections = [((), default),
                  (core_options + ('--profile', worker_profile), set(CORE) | {'worker'}),
                  (core_options + ('--profile', 'application-backup'), set(CORE) | {'backup'})]
    if production:
        selections += [(('--profile', 'production'), set(CORE)),
                       (('--profile', 'processing-worker'), set())]
    for options, expected in selections:
        selected = set(json.loads(environment.capture(*options, 'config', '--format', 'json'))['services'])
        if selected != expected:
            raise EnvironmentError(f'{environment.name} background service selection differs for {options or "default startup"}: ' + ', '.join(sorted(selected ^ expected)))


def templates() -> None:
    configs = {}
    for name in PROJECTS:
        environment = Environment(name, example=True)
        config = environment.config(all_services=True)
        validate_config(environment, config)
        validate_background_selection(environment, config)
        configs[name] = config
    production = Environment('production', example=True)
    if production.config()['services']:
        raise EnvironmentError('Production preparation activates services by default.')
    for volume in ('postgres_data', 'rustfs_data'):
        if len({c['volumes'][volume]['name'] for c in configs.values()}) != 3:
            raise EnvironmentError('Example environments share a named volume.')
    for secret in configs['development']['secrets']:
        paths = {c['secrets'][secret]['file'] for c in configs.values()}
        if len(paths) != 3:
            raise EnvironmentError('Example environments share a secret path.')
        for name, config in configs.items():
            if not Path(config['secrets'][secret]['file']).is_relative_to(ROOT / f'.local/{name}/secrets'):
                raise EnvironmentError('Example secret path is outside its environment directory.')
    ports = {listener_origin(c, role) for c in configs.values() for role in ('public', 'admin')}
    if len(ports) != 6:
        raise EnvironmentError('Example environments share a listener port.')
    core_memory = sum(int(configs['production']['services'][s]['mem_limit']) for s in CORE)
    if core_memory > 4 * 1024**3:
        raise EnvironmentError('Production core memory ceilings exceed the template budget.')
    print('PASS: example topology, isolated paths/volumes/ports, restart policies, opt-in workers, image-only staging/production and production ceilings.')
    print('Production example is prepared with no default active services; its image placeholder is not deployment-ready. Host recovery and capacity are not verified by this check.')


def request(origin: str, path: str) -> tuple[int, bytes]:
    try:
        with urllib.request.urlopen(origin + path, timeout=8) as response:
            return response.status, response.read()
    except urllib.error.HTTPError as response:
        return response.code, response.read()


def check_http(config: dict) -> list[bytes]:
    responses = []
    for role in ('public', 'admin'):
        origin = listener_origin(config, role)
        paths = [('/health/ready', 200), ('/admin/status', 404 if role == 'public' else 200)]
        if role == 'public':
            paths += [('/config', 404), ('/debug/vars', 404)]
        for path, expected in paths:
            status, body = request(origin, path)
            if status != expected:
                raise EnvironmentError(f'{role} {path} returned {status}; expected {expected}.')
            responses.append(body)
    return responses


def used_secret_paths(config: dict, running: set[str]) -> set[Path]:
    return {Path(config['secrets'][s['source']]['file']) for role in running
            for s in config['services'][role].get('secrets', [])}


def check_secret_files(config: dict, running: set[str], containers: list[dict] = ()) -> None:
    for path in used_secret_paths(config, running):
        if not path.is_file() or not os.access(path, os.R_OK) or not path.read_bytes().strip():
            raise EnvironmentError('An active service secret is missing, empty or unreadable (value withheld).')
        # Container secrets are bind mounted; ownership/modes or an ACL must
        # allow the container UID to read. Do not alter existing custom ACLs.
        if not containers and not path.stat().st_mode & 0o004:
            print('CHECK: a secret lacks world-read permission; verify the mounted file is readable by its container UID or custom ACL (path/value withheld).')
    for container in containers:
        labels = container['Config'].get('Labels') or {}
        role = labels.get('com.docker.compose.service')
        if not container['State']['Running'] or labels.get('com.docker.compose.oneoff', 'false').lower() == 'true':
            continue
        for secret in config['services'][role].get('secrets', []):
            target = secret.get('target', secret['source'])
            target = target if target.startswith('/') else '/run/secrets/' + target
            result = subprocess.run(['docker', 'exec', container['Id'], 'sh', '-c', 'test -r "$1"', 'sh', target], capture_output=True)
            if result.returncode:
                raise EnvironmentError(f'{role} cannot read a mounted secret as its configured container user; repair that file\'s permissions or ACL (value withheld).')


def check_runtime(environment: Environment, config: dict) -> None:
    containers = environment.containers()
    if environment.name == 'production' and not containers:
        print('PREPARED: no production containers; runtime readiness and reboot recovery remain unverified.')
        return
    running = inspect_boundaries(environment, config, containers)
    check_secret_files(config, running, containers)
    check_http(config)
    revision = 'unavailable (archive download)'
    result = subprocess.run(['git', '-C', str(environment.root), 'rev-parse', 'HEAD'], capture_output=True, text=True) if (environment.root / '.git').exists() else None
    if result and result.returncode == 0:
        revision = result.stdout.strip()
    print('Checkout revision:', revision)
    policy_mismatches = []
    for container in containers:
        labels = container['Config'].get('Labels') or {}
        role = labels['com.docker.compose.service']
        if not container['State']['Running'] or labels.get('com.docker.compose.oneoff', 'false').lower() == 'true':
            continue
        actual_policy = container['HostConfig']['RestartPolicy']['Name']
        if actual_policy != config['services'][role]['restart']:
            policy_mismatches.append(role)
        health = container['State'].get('Health', {}).get('Status')
        if health and health != 'healthy':
            raise EnvironmentError(f'{role} container health is {health}.')
        if role in APPLICATIONS:
            image = docker_json('image', 'inspect', container['Image'])[0]
            image_revision = (image['Config'].get('Labels') or {}).get('org.opencontainers.image.revision', 'unknown')
            configured = config['services'][role]['image']
            configured_image = docker_json('image', 'inspect', configured)[0]
            print(f'{role}: configured={configured}; running_revision={image_revision}; matches_checkout={image_revision == revision}; matches_configured_image={container["Image"] == configured_image["Id"]}')
            if container['Image'] != configured_image['Id']:
                raise EnvironmentError(f'{role} running image differs from its configured image; deploy deliberately before release validation.')
    if policy_mismatches:
        raise EnvironmentError('Runtime restart policies need a planned rollout for: ' + ', '.join(sorted(policy_mismatches)))
    print(f'PASS: {environment.name} runtime readiness, project/mount/port identity, health and restart policies. No services changed; a host reboot is not proven by this check.')


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--environment', choices=PROJECTS)
    parser.add_argument('--runtime', action='store_true')
    args = parser.parse_args()
    if args.runtime and not args.environment:
        parser.error('--runtime requires --environment')
    try:
        if args.environment:
            environment = Environment(args.environment)
            config = environment.config(all_services=True)
            validate_config(environment, config)
            validate_background_selection(environment, config)
            selected = config['services']['admin']['image']
            print(f'PASS: {environment.name} local configuration; project={environment.project}; application_image={selected}')
            if environment.name == 'production':
                if not DIGEST.fullmatch(selected) and selected != 'ghcr.io/balestrino/italian-weather-alert@sha256:REPLACE_WITH_APPROVED_DIGEST':
                    raise EnvironmentError('Production image selection is invalid; use an approved GHCR SHA-256 digest or the documented preparation placeholder.')
                print('IMAGE SELECTED: verify registry availability and operator approval before deployment.' if DIGEST.fullmatch(selected) else 'PREPARED: select an approved production digest before image operations.')
            if args.runtime:
                check_runtime(environment, config)
        else:
            templates()
        return 0
    except (EnvironmentError, OSError, ValueError) as error:
        print(str(error) if isinstance(error, EnvironmentError) else 'Check failed; inspect Docker availability and selected environment readiness (private output withheld).', file=sys.stderr)
        return 1


if __name__ == '__main__':
    sys.exit(main())
