#!/usr/bin/env python3
"""Validate the committed staging and production Compose examples without starting services."""

import json
import os
from pathlib import Path
import subprocess
from typing import Optional


ROOT = Path(__file__).resolve().parents[1]


def configured(example: str, image: Optional[str] = None, profiles: tuple[str, ...] = ()) -> dict:
    env = {
        key: value
        for key, value in os.environ.items()
        if not key.startswith(("IWA_", "COMPOSE_"))
    }
    if image is not None:
        env["IWA_APP_IMAGE"] = image
    command = ["docker", "compose", "--env-file", example]
    if example == "deploy/production.env.example":
        command += ["-f", "compose.yaml", "-f", "deploy/compose.production.yaml", "-p", "iwa-production"]
    for profile in profiles:
        command += ["--profile", profile]
    output = subprocess.check_output(command + ["config", "--format", "json"], cwd=ROOT, env=env)
    return json.loads(output)


development = configured("deploy/development.env.example")
staging = configured("deploy/staging.env.example")
production_image = "ghcr.io/balestrino/italian-weather-alert@sha256:" + "0" * 64
prepared = configured("deploy/production.env.example", production_image)
production = configured("deploy/production.env.example", production_image, ("production", "production-worker", "application-backup"))
core = configured("deploy/production.env.example", production_image, ("production",))

assert development["name"] == "iwa"
assert staging["name"] == "iwa-staging"
assert production["name"] == "iwa-production"
assert prepared["services"] == {}
assert set(core["services"]) == {"public", "admin", "postgres", "rustfs", "crawl4ai"}
for volume in ("postgres_data", "rustfs_data"):
    names = {config["volumes"][volume]["name"] for config in (development, staging, production)}
    assert len(names) == 3

for secret in development["secrets"]:
    paths = {config["secrets"][secret]["file"] for config in (development, staging, production)}
    assert len(paths) == 3
    assert development["secrets"][secret]["file"].startswith(str(ROOT / ".local/development/secrets"))
    assert staging["secrets"][secret]["file"].startswith(str(ROOT / ".local/staging/secrets"))
    assert production["secrets"][secret]["file"].startswith(str(ROOT / ".local/production/secrets"))

for config in (development, staging, production):
    services = config["services"]
    for name in ("public", "admin"):
        assert len(services[name]["ports"]) == 1
        assert services[name]["ports"][0]["host_ip"] == "127.0.0.1"
    for name in ("postgres", "rustfs", "crawl4ai", "worker", "backup"):
        assert not services[name].get("ports")

listener_ports = {
    (service["ports"][0]["host_ip"], service["ports"][0]["published"])
    for config in (development, staging, production)
    for name, service in config["services"].items()
    if name in ("public", "admin")
}
assert len(listener_ports) == 6

for name in ("public", "admin", "worker", "backup"):
    assert "build" not in production["services"][name]
    assert production["services"][name]["image"] == production_image

for name, service in production["services"].items():
    assert int(service["mem_limit"]) > 0
    assert 0 < service["cpus"] <= 1.5
    assert 0 < service["pids_limit"] <= 512
    assert service["profiles"] == [
        "production-worker" if name == "worker" else "application-backup" if name == "backup" else "production"
    ]

core_memory = sum(int(service["mem_limit"]) for service in core["services"].values())
assert core_memory <= 4 * 1024**3

print("PASS: three isolated Compose projects; prepared production has no active services, bounded resources and digest-only images")
