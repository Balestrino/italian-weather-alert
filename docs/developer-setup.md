# Developer setup

This guide brings up an isolated local IWA stack for development. It uses the Compose services in this repository: PostgreSQL, RustFS, Crawl4AI, the read-only public API/MCP listener, and the private administration listener. The worker is a separate step because it needs a real inference-provider credential.

Commands below assume a POSIX shell on a machine with Docker Engine and the Docker Compose plugin. Run them from the repository root. Python 3 and `curl` are used by the setup and verification commands; Go 1.25 is needed for native builds and tests. Allow several gigabytes of free disk space for the container images, especially Crawl4AI.

## 1. Choose an isolated local project

Use a separate checkout if the machine already hosts an IWA installation. In the shell used for all commands in this guide, set a distinct Compose project, image tag, and host ports:

```sh
export COMPOSE_PROJECT_NAME=iwa-dev
export IWA_APP_IMAGE=iwa-app:dev
export IWA_PUBLIC_PORT=18080
export IWA_ADMIN_PORT=18081
```

Compose binds the two HTTP ports to `127.0.0.1`. PostgreSQL, RustFS, Crawl4AI, and the workers have no host ports. The export values above are examples; choose free ports if 18080 or 18081 are occupied. Keep the same values when you later run `docker compose` commands.

## 2. Initialize local configuration

```sh
python3 scripts/init-secrets.py
docker compose config --quiet
```

The script creates `.secrets/` with PostgreSQL, RustFS, Crawl4AI, and public cursor secrets, plus disabled notification and backup configurations. It preserves existing values. Both `.secrets/` and the optional `.env` file are ignored by Git. Compose defaults are enough for this local setup; put persistent, nonsecret overrides such as the host ports in your own `.env` if you prefer.

| Setting | Default | Purpose |
| --- | --- | --- |
| `COMPOSE_PROJECT_NAME` | `iwa` | Compose project and volume namespace; use `iwa-dev` for an isolated checkout. |
| `IWA_APP_IMAGE` | `iwa-app:local` | Local application image tag. |
| `IWA_PUBLIC_PORT` / `IWA_ADMIN_PORT` | `8080` / `8081` | Loopback host ports for API/MCP and administration. |
| `IWA_REGOLO_API_KEY_PATH` | `./.secrets/regolo_api_key` | Existing provider key file, needed only when starting the worker. |
| `IWA_SEMANTIC_LINKING_ENABLED` | `false` | Optional semantic retrieval in the worker. |
| `IWA_PUBLIC_COPY_ACCESS_ENABLED` | `false` | Application-mediated retained copies on the public listener. |

Do not replace the PostgreSQL password while keeping an existing PostgreSQL volume: the database was initialized with the earlier value. For a local run, leave public copy access, semantic linking, notifications, and backups at their disabled defaults.

## 3. Start dependencies and initialize storage

```sh
docker compose up -d --wait --wait-timeout 180 postgres rustfs crawl4ai
docker compose build admin
docker compose run --rm --no-deps admin migrate
docker compose run --rm --no-deps admin storage-init
```

The migration creates the application schema. `storage-init` prepares the object store for retained documents. Both are admin-only commands. These steps do not import or publicly enable any alert source.

## 4. Start the two HTTP listeners

```sh
docker compose up -d --wait --wait-timeout 180 public admin
docker compose ps
curl --fail http://127.0.0.1:18080/health/ready
curl --fail http://127.0.0.1:18081/health/ready
curl --fail http://127.0.0.1:18081/admin/status
```

Open the administration UI at [http://127.0.0.1:18081/admin/](http://127.0.0.1:18081/admin/). The JSON API uses `http://127.0.0.1:18080/v1/`; the MCP endpoint is `http://127.0.0.1:18080/mcp`. Substitute your chosen host ports in these URLs. See the [verified API and MCP examples](../deploy/public-usage.md) for requests against a populated local fixture.

A new database has no accepted alert sources. Empty coverage or search results are expected until source data is explicitly configured and reviewed. The [national coverage tracker](coverage.md) is documentation, not an automatic source import.

## 5. Configure sources for development

The administration UI at `/admin/sources` supports authority and channel records, source drafts, previews, and separate collection and publication controls.

For source work, record the official referral, territory, product, access and reuse conditions, sections to check, and representative examples. Preview the configured revision before enabling collection. A successful preview does not certify complete regional or municipal coverage; observation, acceptance, and public enablement are separate steps. Use synthetic or otherwise authorized fixtures for ordinary development.

Municipality and CAP data in `docs/` are research inputs. They are not automatically selected as runtime geography datasets by this setup; dataset registration and selection are explicit administration steps.

## 6. Add the worker only when needed

`scripts/init-secrets.py` does **not** create a Regolo API key. The `worker` service requires a real key file containing one nonempty line. Put it at the default ignored path `.secrets/regolo_api_key`, or set `IWA_REGOLO_API_KEY_PATH` to an existing private file that the container can read. The generated secret files use mode `0644` inside a host directory with mode `0700`; a custom key file must likewise be readable by the container user. Do not put the key in `.env` or a command argument.

After configuring an intended test source and the key file:

```sh
docker compose up -d worker
docker compose ps worker
docker compose logs --tail=100 worker
```

The worker can fetch configured sources and call external inference services, which may incur charges. The `backup` service is separate and its generated configuration is disabled. Review the configured source, model provider, and OCR settings before testing real acquisition or interpretation.

## 7. Run checks and stop the stack

```sh
go test ./...
go vet ./...
go build ./cmd/...
docker compose config --quiet
```

For a running **throwaway** stack, `python3 scripts/smoke-compose.py` exercises dependency failure and recovery; it temporarily stops RustFS and starts it again. Do not run that smoke test against an operational instance.

To stop this local project:

```sh
docker compose down
```

`down` keeps the PostgreSQL and RustFS volumes and leaves `.secrets/` intact. To diagnose startup trouble, run `docker compose ps` and `docker compose logs --tail=100 public admin postgres rustfs crawl4ai`. A readiness response of 503 means at least one dependency is unavailable; a healthy container by itself does not prove source coverage or interpretation quality.
