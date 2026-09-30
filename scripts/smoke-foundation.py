#!/usr/bin/env python3
"""Bounded local smoke: real PostgreSQL, explicitly simulated RustFS/Crawl4AI.
Uses only the existing postgres:16-alpine image; owns/removes its own container.
"""
import http.server
import json
import os
from pathlib import Path
import socket
import subprocess
import tempfile
import threading
import time
import urllib.error
import urllib.request
import uuid

root = Path(__file__).resolve().parents[1]
name = 'iwa-smoke-' + uuid.uuid4().hex[:12]
processes = []
logs = []
responses = []
def docker(*args):
    return subprocess.check_output(['docker', *args], text=True, stderr=subprocess.PIPE).strip()
def request(url, method='GET', headers=None):
    req = urllib.request.Request(url, method=method, headers=headers or {})
    try:
        with urllib.request.urlopen(req, timeout=5) as response:
            body = response.read()
            responses.append(body)
            return response.status, body
    except urllib.error.HTTPError as error:
        body = error.read()
        responses.append(body)
        return error.code, body
def port():
    with socket.socket() as s:
        s.bind(('127.0.0.1', 0))
        return s.getsockname()[1]
class Fixture(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200 if self.path in ('/health', '/health/ready') else 404)
        self.end_headers()
        self.wfile.write(b'{}')
    def log_message(self, *args):
        pass
fixture = http.server.ThreadingHTTPServer(('127.0.0.1', 0), Fixture)
threading.Thread(target=fixture.serve_forever, daemon=True).start()
try:
    docker('run', '--pull=never', '--detach', '--name', name,
           '--tmpfs', '/var/lib/postgresql/data:rw,size=256m',
           '--publish', '127.0.0.1::5432',
           '--mount', f'type=bind,src={root / ".secrets/postgres_password"},dst=/run/secrets/postgres_password,readonly',
           '--env', 'POSTGRES_USER=iwa', '--env', 'POSTGRES_DB=iwa',
           '--env', 'POSTGRES_PASSWORD_FILE=/run/secrets/postgres_password', 'postgres:16-alpine')
    db = docker('port', name, '5432/tcp')
    urls = {}
    for role in ('public', 'admin'):
        listen = f'127.0.0.1:{port()}'
        env = dict(os.environ, IWA_ROLE=role, IWA_LISTEN=listen, IWA_CONTAINER_ADMIN='false',
                   IWA_POSTGRES_HOST=db,
                   IWA_POSTGRES_PASSWORD_FILE=str(root / '.secrets/postgres_password'),
                   IWA_CRAWL_TOKEN_FILE=str(root / '.secrets/crawl_token'),
                   IWA_PUBLIC_CURSOR_KEY_FILE=str(root / '.secrets/public_cursor_key'),
                   IWA_RUSTFS_URL=f'http://127.0.0.1:{fixture.server_port}',
                   IWA_RUSTFS_ACCESS_KEY_FILE=str(root / '.secrets/rustfs_access_key'),
                   IWA_RUSTFS_SECRET_KEY_FILE=str(root / '.secrets/rustfs_secret_key'),
                   IWA_CRAWL_URL=f'http://127.0.0.1:{fixture.server_port}')
        log = tempfile.TemporaryFile(); logs.append(log)
        proc = subprocess.Popen([str(root / 'bin/iwa')], env=env, stdout=log, stderr=log)
        processes.append(proc); urls[role] = 'http://' + listen
        deadline = time.monotonic()+25
        while True:
            try:
                if request(urls[role]+'/health/ready')[0] == 200:
                    break
            except (OSError, urllib.error.URLError):
                pass
            if proc.poll() is not None or time.monotonic() > deadline:
                raise RuntimeError('service did not become ready; logs withheld to protect credentials')
            time.sleep(.2)
        subprocess.run([str(root / 'bin/iwa'), 'check'], env=env, check=True)
    assert request(urls['public']+'/admin/status')[0] == 404
    assert request(urls['admin']+'/admin/status')[0] == 200
    assert request(urls['admin']+'/admin/')[0] == 200
    for method in ('GET', 'POST', 'PUT', 'PATCH', 'DELETE'):
        assert request(urls['admin']+'/admin/status', method, {'Host': 'evil.example'})[0] == 403
        assert request(urls['admin']+'/admin/status', method, {'Origin': 'https://evil.example'})[0] == 403
        assert request(urls['admin']+'/admin/status', method, {'Origin': urls['public']})[0] == 403
        for path in ('/admin/', '/admin/sources', '/admin/config', '/admin/reprocess', '/admin/notifications', '/admin/backups', '/admin/processing-evaluations', '/config'):
            assert request(urls['public']+path, method)[0] == 404
    # Exercise the real executable's role checks, including configuration writes.
    public_env = dict(env, IWA_ROLE='public', IWA_LISTEN=urls['public'].removeprefix('http://'))
    for args in (['report-preview'], ['report-status'], ['interpretation-preflight', '0', '1'], ['migrate'], ['documents-archive-pending', '2026-01-01T00:00:00Z', 'test'], ['jobs-archive', '2026-01-01T00:00:00Z', 'test'], ['inference-catalog-check'], ['storage-init'], ['storage-recover'], ['retention-policy', '2'], ['retention-cleanup'], ['preview', 'fixture', '1'], ['backup-result', 'fixture', 'run', 'failed', '2026-01-01T00:00:00Z']):
        result = subprocess.run([str(root / 'bin/iwa'), *args], env=public_env, capture_output=True, timeout=8)
        assert result.returncode != 0 and b'admin role' in result.stderr, 'public command bypassed administrative role'
        responses.extend((result.stdout, result.stderr))
    result = subprocess.run([str(root / 'bin/iwa'), 'backup-worker'], env=public_env, capture_output=True, timeout=8)
    assert result.returncode != 0 and b'worker role' in result.stderr
    responses.extend((result.stdout, result.stderr))
    cfg = docker('compose', 'config', '--format', 'json')
    conf = json.loads(cfg)
    for svc in ('postgres', 'rustfs', 'crawl4ai'):
        assert not conf['services'][svc].get('ports'), svc
    assert conf['services']['admin']['ports'][0]['host_ip'] == '127.0.0.1'
    docker('stop', '--time', '5', name)
    assert request(urls['public']+'/health/ready')[0] == 503
    assert request(urls['public']+'/health/live')[0] == 200
    for proc in processes:
        proc.terminate(); proc.wait(timeout=8); assert proc.returncode == 0
    for log in logs:
        log.seek(0)
        output = log.read()
        for path in (root / '.secrets').iterdir():
            if path.is_file():
                value = path.read_bytes().strip()
                assert value not in output and value.decode() not in cfg and all(value not in body for body in responses), 'secret disclosure'
    print('PASS: native listeners, real PostgreSQL, dependency failure, shutdown, Compose boundaries, browser/route/CLI administration isolation and secret redaction.')
    print('RustFS/Crawl4AI responses were simulated; complete Compose startup remains unverified.')
finally:
    for proc in processes:
        if proc.poll() is None:
            proc.terminate()
            try:
                proc.wait(timeout=8)
            except subprocess.TimeoutExpired:
                proc.kill(); proc.wait()
    subprocess.run(['docker', 'rm', '-f', name], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    fixture.shutdown(); fixture.server_close()
    for log in logs:
        log.close()
