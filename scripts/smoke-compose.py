#!/usr/bin/env python3
"""Check one explicit environment; fault injection requires --allow-interruption."""
from __future__ import annotations

import argparse
import json
import sys
import time

from environment_config import (Environment, EnvironmentError, PROJECTS,
                                inspect_boundaries, listener_origin)
from deployment_checks import (check_http, check_secret_files, request,
                               used_secret_paths, validate_config)


def ready(origin: str) -> None:
    deadline = time.monotonic() + 90
    while time.monotonic() < deadline:
        try:
            if request(origin, '/health/ready')[0] == 200:
                return
        except OSError:
            pass
        time.sleep(1)
    raise EnvironmentError('Selected environment readiness did not recover.')


def fault_check(environment: Environment, config: dict) -> list[bytes]:
    public = listener_origin(config, 'public')
    admin = listener_origin(config, 'admin')
    responses = []
    try:
        environment.capture('stop', 'rustfs')
        status, body = request(public, '/health/ready')
        responses.append(body)
        if status != 503 or request(public, '/health/live')[0] != 200:
            raise EnvironmentError('Dependency outage was not reported truthfully.')
        status, body = request(admin, '/health/ready')
        responses.append(body)
        if status != 503 or json.loads(body) != {'postgres': True, 'rustfs': False, 'crawl4ai': True}:
            raise EnvironmentError('Administrative dependency state did not show the selected outage.')
    finally:
        environment.capture('start', '--wait', '--wait-timeout', '180', 'rustfs')
        ready(public)
        ready(admin)
    return responses


def smoke(environment: Environment, allow_interruption: bool = False) -> None:
    if allow_interruption and environment.name == 'production':
        raise EnvironmentError('Production fault injection is not supported; use the ordinary nondisruptive check.')
    config = environment.config(all_services=True)
    validate_config(environment, config)
    containers = environment.containers()
    running = inspect_boundaries(environment, config, containers)
    check_secret_files(config, running, containers)
    responses = check_http(config)
    if allow_interruption:
        responses += fault_check(environment, config)
    outputs = [json.dumps(config).encode(), json.dumps(containers).encode(), *responses]
    for name in sorted(running):
        outputs.append(environment.capture('logs', '--no-color', '--tail=100', name))
    for path in used_secret_paths(config, running):
        value = path.read_bytes().strip()
        if value and any(value in output for output in outputs):
            raise EnvironmentError('A secret appeared in configuration, runtime inspection, response or recent logs (value withheld).')
    print(f'PASS: {environment.name} HTTP readiness, public/admin separation, runtime project/port/mount identity and secret checks.')
    print('Selected RustFS outage/recovery verified.' if allow_interruption else 'No services stopped, started or recreated.')


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('environment', choices=PROJECTS)
    parser.add_argument('--allow-interruption', action='store_true', help='Planned development/staging RustFS outage and recovery; never production')
    args = parser.parse_args()
    # Reject interruption before resolving a local production file or invoking Docker.
    if args.environment == 'production' and args.allow_interruption:
        parser.error('production fault injection is not supported')
    try:
        smoke(Environment(args.environment), args.allow_interruption)
        return 0
    except (EnvironmentError, OSError, ValueError) as error:
        print(str(error) if isinstance(error, EnvironmentError) else 'Smoke check failed; inspect the selected environment (private output withheld).', file=sys.stderr)
        return 1


if __name__ == '__main__':
    sys.exit(main())
