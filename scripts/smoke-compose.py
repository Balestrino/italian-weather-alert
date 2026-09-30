#!/usr/bin/env python3
"""Verify the running local Compose stack without printing credentials.
Temporarily stops RustFS to verify failure/recovery; restarts it in finally.
"""
import json
import os
from pathlib import Path
import subprocess
import time
import urllib.error
import urllib.request

root = Path(__file__).resolve().parents[1]

def compose(*args):
    return subprocess.check_output(['docker', 'compose', *args], cwd=root, stderr=subprocess.PIPE)

def request(base, path):
    try:
        with urllib.request.urlopen(base + path, timeout=6) as res:
            return res.status, res.read()
    except urllib.error.HTTPError as res:
        return res.code, res.read()

def ready(base):
    deadline = time.monotonic() + 90
    while time.monotonic() < deadline:
        try:
            if request(base, '/health/ready')[0] == 200:
                return
        except OSError:
            pass
        time.sleep(1)
    raise RuntimeError('readiness did not recover')

public = 'http://127.0.0.1:' + os.environ.get('IWA_PUBLIC_PORT', '8080')
admin = 'http://127.0.0.1:' + os.environ.get('IWA_ADMIN_PORT', '8081')
ready(public)
ready(admin)
responses = []
for base, path, code in [(public, '/health/ready', 200), (admin, '/health/ready', 200),
                         (public, '/admin/status', 404), (admin, '/admin/status', 200),
                         (public, '/config', 404), (public, '/debug/vars', 404)]:
    status, body = request(base, path)
    assert status == code, f'unexpected status for {path}'
    responses.append(body)
assert json.loads(request(admin, '/health/ready')[1]) == {'postgres': True, 'rustfs': True, 'crawl4ai': True}
try:
    compose('stop', 'rustfs')
    status, body = request(public, '/health/ready')
    assert status == 503, 'dependency outage was hidden'
    responses.append(body)
    assert request(public, '/health/live')[0] == 200
    state = json.loads(request(admin, '/health/ready')[1])
    assert state == {'postgres': True, 'rustfs': False, 'crawl4ai': True}
finally:
    compose('start', 'rustfs')
ready(public)
ready(admin)
config = compose('config', '--format', 'json')
parsed = json.loads(config)
listeners = ('public', 'admin')
internal = ('worker', 'backup', 'postgres', 'rustfs', 'crawl4ai')
for name in internal:
    assert not parsed['services'][name].get('ports'), f'{name} has a host port'
for name in listeners:
    assert all(p['host_ip'] == '127.0.0.1' for p in parsed['services'][name]['ports'])
# Inspect actual runtime bindings, not only the intended YAML.
ids = compose('ps', '-q').decode().split()
inspected = subprocess.check_output(['docker', 'inspect', *ids])
for container in json.loads(inspected):
    role = container['Config']['Labels']['com.docker.compose.service']
    bindings = container['HostConfig'].get('PortBindings') or {}
    actual = container['NetworkSettings'].get('Ports') or {}
    published = [b for entries in actual.values() if entries for b in entries]
    if role in internal:
        assert not bindings and not published, f'{role} internal runtime port exposed'
    elif role in listeners:
        assert published and all(b['HostIp'] == '127.0.0.1' for b in published), f'{role} runtime listener mapping missing or nonlocal'
    else:
        raise AssertionError(f'unexpected Compose service {role}')
outputs = [config, inspected, *responses]
for name in (*listeners, *internal):
    outputs.append(compose('logs', '--no-color', name))
for path in (root / '.secrets').iterdir():
    if path.is_file():
        secret = path.read_bytes().strip()
        assert secret and all(secret not in output for output in outputs), 'secret found in logs/config/response (value withheld)'
print('PASS: actual PostgreSQL/RustFS/Crawl4AI readiness, public/admin isolation, dependency outage/recovery, runtime port bindings and secret checks.')
