#!/usr/bin/env python3
"""Shared environment selection for Compose, validation and smoke checks."""

from __future__ import annotations

import json
import os
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PROJECTS = {"development": "iwa", "staging": "iwa-staging", "production": "iwa-production"}
APPLICATIONS = ("public", "admin", "worker", "backup")
CORE = ("public", "admin", "postgres", "rustfs", "crawl4ai")
DIGEST = re.compile(r"ghcr\.io/[a-z0-9][a-z0-9._/-]*@sha256:[0-9a-fA-F]{64}\Z")
FORBIDDEN = ("--project-name", "--project-directory", "--env-file", "--file")
BOOLEAN_OPTIONS = {"--dry-run", "--compatibility", "--all-resources"}
VALUE_OPTIONS = {"--profile", "--ansi", "--progress", "--parallel"}


class EnvironmentError(RuntimeError):
    """An operator-correctable error safe to print without private output."""


def clean_environment() -> dict[str, str]:
    return {key: value for key, value in os.environ.items()
            if not key.startswith(("IWA_", "COMPOSE_"))}


def compose_command(args: list[str]) -> str:
    """Validate wrapper arguments, including persistent flags after a command."""
    for argument in args:
        option = argument.split("=", 1)[0]
        follow_logs = args[0] == "logs" and argument == "-f"
        if option in FORBIDDEN or (argument.startswith(("-p", "-f")) and not follow_logs):
            raise EnvironmentError("Project, file, environment-file and directory overrides are not supported.")
    index = 0
    while index < len(args) and args[index].startswith("-"):
        option, separator, value = args[index].partition("=")
        if option in BOOLEAN_OPTIONS and not separator:
            index += 1
        elif option in VALUE_OPTIONS:
            if not separator:
                index += 1
                if index >= len(args) or args[index].startswith("--"):
                    raise EnvironmentError(f"Missing value for {option}.")
                value = args[index]
            if not value:
                raise EnvironmentError(f"Missing value for {option}.")
            index += 1
        else:
            raise EnvironmentError("Unsupported leading Compose option; use --profile, --dry-run, --ansi, --progress or --parallel.")
    if index >= len(args):
        raise EnvironmentError("Missing Docker Compose command.")
    return args[index]


def require_production_image(config: dict) -> None:
    for name in APPLICATIONS:
        service = config["services"][name]
        if service.get("build") or not DIGEST.fullmatch(service.get("image", "")):
            raise EnvironmentError("Production requires an approved ghcr.io/...@sha256:<64 hex characters> application image; replace the placeholder before pulling or starting.")


class Environment:
    def __init__(self, name: str, root: Path = ROOT, example: bool = False):
        if name not in PROJECTS:
            raise EnvironmentError("Choose development, staging or production explicitly.")
        self.name = name
        self.root = root.resolve()
        self.project = PROJECTS[name]
        self.env_file = self.root / (f"deploy/{name}.env.example" if example else f".local/{name}.env")
        if not self.env_file.is_file():
            raise EnvironmentError(f"Missing {self.env_file.relative_to(self.root)}; follow docs/operations/environments.md to prepare this environment.")
        self.process_environment = clean_environment()

    def command(self, *args: str) -> list[str]:
        command = ["docker", "compose", "--env-file", str(self.env_file),
                   "--project-directory", str(self.root), "-f", str(self.root / "compose.yaml")]
        if self.name != "development":
            command += ["-f", str(self.root / f"deploy/compose.{self.name}.yaml")]
        return command + ["-p", self.project, *args]

    def capture(self, *args: str) -> bytes:
        try:
            return subprocess.check_output(self.command(*args), cwd=self.root,
                                           env=self.process_environment, stderr=subprocess.PIPE)
        except subprocess.CalledProcessError as error:
            raise EnvironmentError(f"Compose check failed for {self.name}; inspect its configuration and service status (private output withheld).") from error

    def config(self, all_services: bool = False) -> dict:
        args = ["--profile", "*"] if all_services else []
        return json.loads(self.capture(*args, "config", "--format", "json"))

    def containers(self) -> list[dict]:
        ids = self.capture("ps", "--all", "-q").decode().split()
        return docker_json("inspect", *ids) if ids else []


def docker_json(*args: str):
    try:
        return json.loads(subprocess.check_output(["docker", *args], stderr=subprocess.PIPE))
    except subprocess.CalledProcessError as error:
        raise EnvironmentError("Docker inspection failed (private output withheld).") from error


def listener_host_ip(environment: str, name: str) -> str:
    return "0.0.0.0" if environment == "development" and name == "public" else "127.0.0.1"


def listener_origin(config: dict, name: str) -> str:
    ports = config["services"][name].get("ports", [])
    environment = config["services"][name].get("environment", {}).get("IWA_ENVIRONMENT", "production")
    expected = listener_host_ip(environment, name)
    if len(ports) != 1 or ports[0].get("host_ip") != expected:
        raise EnvironmentError(f"{name} must publish one listener bound to {expected}.")
    return f"http://127.0.0.1:{ports[0]['published']}"


def inspect_boundaries(environment: Environment, config: dict, containers: list[dict],
                       require_running: bool = True) -> set[str]:
    running = set()
    for container in containers:
        labels = container["Config"].get("Labels") or {}
        name = labels.get("com.docker.compose.service")
        if labels.get("com.docker.compose.project") != environment.project or name not in config["services"]:
            raise EnvironmentError("Runtime container belongs to an unexpected project or service.")
        if labels.get("com.docker.compose.oneoff", "false").lower() == "true":
            continue
        if not container["State"]["Running"]:
            continue
        running.add(name)
        bindings = container["HostConfig"].get("PortBindings") or {}
        ports = container["NetworkSettings"].get("Ports") or {}
        published = [entry for entries in ports.values() if entries for entry in entries]
        if name in ("public", "admin"):
            listener_origin(config, name)
            expected = str(config["services"][name]["ports"][0]["published"])
            expected_ip = listener_host_ip(environment.name, name)
            if len(published) != 1 or published[0]["HostIp"] != expected_ip or published[0]["HostPort"] != expected:
                raise EnvironmentError(f"Unexpected actual listener binding for {name}.")
        elif bindings or published:
            raise EnvironmentError(f"Internal service {name} publishes a host port.")
        mounts = {mount["Destination"]: mount for mount in container["Mounts"]}
        for volume in config["services"][name].get("volumes", []):
            if volume["type"] == "volume":
                mount = mounts.get(volume["target"], {})
                if mount.get("Name") != config["volumes"][volume["source"]]["name"]:
                    raise EnvironmentError(f"Unexpected named volume mounted by {name}.")
        for secret in config["services"][name].get("secrets", []):
            target = secret.get("target", secret["source"])
            mount = mounts.get(target if target.startswith("/") else "/run/secrets/" + target, {})
            expected = Path(config["secrets"][secret["source"]]["file"]).resolve()
            if Path(mount.get("Source", "/missing")).resolve() != expected or mount.get("RW", True):
                raise EnvironmentError(f"Unexpected secret mount for {name} (value withheld).")
    if require_running and not set(CORE).issubset(running):
        raise EnvironmentError("Core services are not all running; inspect the selected environment's ps and logs.")
    return running


def verify_staging_release(image_tag: str, revision: str) -> None:
    environment = Environment("staging")
    config = environment.config(all_services=True)
    candidate = docker_json("image", "inspect", image_tag)[0]
    labels = candidate["Config"].get("Labels") or {}
    if labels.get("org.opencontainers.image.revision") != revision:
        raise EnvironmentError("Local release image revision differs from the clean checkout.")
    containers = environment.containers()
    running = inspect_boundaries(environment, config, containers)
    for container in containers:
        labels = container["Config"].get("Labels") or {}
        role = labels.get("com.docker.compose.service")
        if not container["State"]["Running"] or role not in APPLICATIONS or labels.get("com.docker.compose.oneoff", "false").lower() == "true":
            continue
        if config["services"][role]["image"] != image_tag or container["Image"] != candidate["Id"]:
            raise EnvironmentError("Staging application image differs from the release candidate; validate that exact image before publishing.")
    if not {"public", "admin"}.issubset(running):
        raise EnvironmentError("Both staging application listeners must be running before publication.")


def main() -> int:
    try:
        if len(sys.argv) == 4 and sys.argv[1] == "--verify-staging-release":
            verify_staging_release(sys.argv[2], sys.argv[3])
            return 0
        if len(sys.argv) < 3:
            raise EnvironmentError("Usage: scripts/compose-env.sh development|staging|production <docker compose command>")
        environment = Environment(sys.argv[1])
        args = sys.argv[2:]
        command = compose_command(args)
        if environment.name != "development" and (command in ("build", "watch") or any(a.split("=", 1)[0] == "--build" for a in args)):
            raise EnvironmentError("Staging and production reuse release images; build the candidate with scripts/release-image.sh build.")
        if environment.name == "production" and command in ("pull", "up", "run", "create", "start", "restart", "scale", "unpause"):
            require_production_image(environment.config(all_services=True))
        os.chdir(environment.root)
        os.execvpe("docker", environment.command(*args), environment.process_environment)
    except (EnvironmentError, OSError, ValueError) as error:
        print(str(error) if isinstance(error, EnvironmentError) else "Environment command failed; check Docker/Python availability and local configuration.", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main())
