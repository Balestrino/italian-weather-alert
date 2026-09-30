#!/usr/bin/env python3
"""Generate local service secrets once, without printing or replacing them."""
import os
from pathlib import Path
import secrets

root = Path(__file__).resolve().parents[1] / '.secrets'
root.mkdir(mode=0o700, exist_ok=True)
root.chmod(0o700)
for name in ('postgres_password', 'rustfs_access_key', 'rustfs_secret_key', 'crawl_token', 'public_cursor_key'):
    try:
        fd = os.open(root / name, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o644)
    except FileExistsError:
        continue
    # Parent 0700 protects host access; container services run with distinct UIDs
    # and Compose bind-mounted secrets must be readable by their runtime user.
    with os.fdopen(fd, 'w') as out:
        out.write(secrets.token_hex(32) + '\n')
for name in ('notifications_config.json', 'backups_config.json'):
    try:
        fd = os.open(root / name, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o644)
    except FileExistsError:
        continue
    with os.fdopen(fd, 'w') as out:
        out.write('{"enabled":false}\n')
print('Local service secret files ready; existing values preserved.')
