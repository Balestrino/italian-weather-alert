#!/usr/bin/env python3
"""Validate the committed staging and production Compose examples without starting services."""

import json
import os
from pathlib import Path
import subprocess
from typing import Optional


ROOT = Path(__file__).resolve().parents[1]


def configured(example: str, image: Optional[str] = None) -> dict:
    env = {
        key: value
        for key, value in os.environ.items()
        if not key.startswith(("IWA_", "COMPOSE_"))
    }
    if image is not None:
        env["IWA_APP_IMAGE"] = image
    output = subprocess.check_output(
        ["docker", "compose", "--env-file", example, "config", "--format", "json"],
        cwd=ROOT,
        env=env,
    )
    return json.loads(output)


development = configured("deploy/development.env.example")
staging = configured("deploy/staging.env.example")
production_image = "ghcr.io/balestrino/italian-weather-alert@sha256:" + "0" * 64
production = configured("deploy/production.env.example", production_image)

assert development["name"] == "iwa"
assert staging["name"] == "iwa-staging"
assert production["name"] == "iwa-production"
for volume in ("postgres_data", "rustfs_data"):
    names = {config["volumes"][volume]["name"] for config in (development, staging, production)}
    assert len(names) == 3

for secret in development["secrets"]:
    paths = {config["secrets"][secret]["file"] for config in (development, staging, production)}
    assert len(paths) == 3
    assert development["secrets"][secret]["file"].startswith(str(ROOT / ".local/development/secrets"))
    assert staging["secrets"][secret]["file"].startswith(str(ROOT / ".local/staging/secrets"))

for config in (development, staging, production):
    services = config["services"]
    for name in ("public", "admin"):
        assert len(services[name]["ports"]) == 1
        assert services[name]["ports"][0]["host_ip"] == "127.0.0.1"
    for name in ("postgres", "rustfs", "crawl4ai", "worker", "backup"):
        assert not services[name].get("ports")

for name in ("public", "admin", "worker", "backup"):
    assert "build" not in production["services"][name]
    assert production["services"][name]["image"] == production_image

print("PASS: development, staging and production Compose examples have isolated resources and production is digest-only")
