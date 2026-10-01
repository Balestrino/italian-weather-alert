# Developer setup

This procedure runs development (`iwa`) with PostgreSQL, RustFS, Crawl4AI and the public/admin HTTP listeners. It works from a Git clone or an extracted ZIP; no agent tool is required. Read [environment prerequisites and existing-installation guidance](environments.md) first. If this Docker host already has an `iwa` project or volumes, preserve its credentials/settings instead of following the fresh-copy steps.

## 1. Prepare a fresh development project

Run from the repository root:

```sh
mkdir -p .local/development
install -m 600 deploy/development.env.example .local/development.env
python3 scripts/init-secrets.py --directory .local/development/secrets
scripts/compose-env.sh development config --quiet
```

The ignored environment file selects development's image and loopback ports, defaulting to `8080` and `8081`. Edit these ports if occupied. The wrapper clears shell overrides; persistent values belong in `.local/development.env`. Secret generation preserves existing values and creates disabled notification/backup configurations. It does not create an inference-provider key.

## 2. Start dependencies and initialize storage

```sh
scripts/compose-env.sh development up -d --wait --wait-timeout 180 postgres rustfs crawl4ai
scripts/compose-env.sh development build admin
scripts/compose-env.sh development run --rm --no-deps --pull never admin migrate
scripts/compose-env.sh development run --rm --no-deps --pull never admin storage-init
```

Migration prepares the application schema; storage initialization prepares the private object bucket. Both commands require the admin role. They do not import or enable sources. Preserve existing PostgreSQL credentials when retaining database volumes.

## 3. Start and check the listeners

```sh
scripts/compose-env.sh development up -d --no-build --pull never --wait --wait-timeout 180 public admin
scripts/compose-env.sh development ps
python3 scripts/smoke-compose.py development
python3 scripts/check-deployment-config.py --environment development --runtime
curl --fail http://127.0.0.1:8080/health/ready
curl --fail http://127.0.0.1:8081/health/ready
curl --fail http://127.0.0.1:8081/admin/status
```

Open [the administration UI](http://127.0.0.1:8081/admin/). The public JSON API is `http://127.0.0.1:8080/v1/` and MCP is `http://127.0.0.1:8080/mcp`. Substitute configured ports in URLs. A new database has no accepted sources, so empty public results are expected. The [coverage tracker](coverage.md) does not import runtime sources automatically.

## 4. Configure source work and optional collection

Administration at `/admin/sources` supports authority/channel records, source drafts, acquisition previews and separate collection/publication controls. Use synthetic or authorized fixtures. Record official referrals, product/territory scope, access/reuse conditions and representative examples; preview before collection activation. A preview is not source acceptance or public enablement. Geography data in `docs/` are research inputs; runtime dataset registration and selection are explicit operations.

Only start `worker` when a controlled source and the intended provider are configured. Put the real provider key in `.local/development/secrets/regolo_api_key`, or set `IWA_REGOLO_API_KEY_PATH` in `.local/development.env` to another private readable file. The file must contain one nonempty line and be readable by application UID 65532, through its mode or a deliberate ACL. Keep it inside an owner-private directory; never put the key in `.env` or a command argument.

```sh
scripts/compose-env.sh development up -d --no-build --pull never worker
scripts/compose-env.sh development logs --tail=100 worker
```

Collection/provider requests can incur charges. Notifications, semantic linking, public retained-copy access and application backups start disabled. The optional `backup` service needs its own reviewed configuration; the whole-VM production strategy uses off-host PBS instead.

## 5. Test, update and manage development

```sh
python3 -m unittest scripts/test_environment_operations.py
python3 scripts/check-deployment-config.py
go test ./...
go vet ./...
go build ./cmd/...
go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...
```

Before changing a capability, read [OpenSpec plans and progress](../openspec/README.md). [Agent tools](agent-tools.md) are optional contributor tools. Integration coverage instructions are in [CONTRIBUTING.md](../CONTRIBUTING.md). For an opt-in fresh-snapshot Compose rehearsal using unique temporary projects and generated credentials, run `python3 scripts/rehearse-environments.py --run` on a host with spare capacity. It builds an image, starts disposable services, exercises a disposable RustFS outage and removes only its own containers/volumes/image afterward. It does not prove a VM reboot or production readiness.

An ordinary restart does not rebuild your code. To deploy changed development code deliberately, first rerun applicable tests, then:

```sh
scripts/compose-env.sh development build admin
scripts/compose-env.sh development run --rm --no-deps --pull never admin migrate
scripts/compose-env.sh development up -d --no-build --pull never --wait --wait-timeout 180 public admin
```

Update any already activated worker separately with the same selected image and its reviewed configuration. Keep a compatible prior image/configuration for rollback and preserve data volumes and additive migrations. For releases, use the separate [staging/production procedure](../deploy/release-operations.md).

```sh
scripts/compose-env.sh development logs --tail=100 public admin postgres rustfs crawl4ai
scripts/compose-env.sh development stop
scripts/compose-env.sh development start --wait --wait-timeout 180 postgres rustfs crawl4ai public admin
scripts/compose-env.sh development down
```

`stop`/`start` operate on existing containers. `down` retains PostgreSQL/RustFS volumes and private files. A readiness response of 503 means a dependency is unavailable; inspect the selected project's status and logs. Do not use `down -v` as routine recovery. Ordinary smoke checks change no services; [planned fault injection](environments.md#planned-failure-checks) requires an explicit interruption option.
