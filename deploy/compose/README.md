# Compose Quick Start

Local all-in-one startup for DotBlue, an enterprise-grade self-hosted AI assistant platform for private deployment, team collaboration, and incremental support for mainstream assistant ecosystems.

## Important Reminder

- If you change the platform-level LLM provider or model in the admin UI after `hermes_*` agent containers are already running, recycle or restart those existing agent containers before re-testing chat.
- Reason: each running agent container keeps its own generated runtime config, so provider changes are not applied to that container until it is recreated.
- In the Compose setup here, this is easy to miss during end-to-end testing because the backend and worker loop are all-in-one, while agent runtimes are still separate `hermes_*` containers.

## Files

- `.env.example`: user-editable deployment variables
- `prepare-config.sh`: Linux/macOS config generator
- `prepare-config.ps1`: Windows PowerShell config generator
- `smoke-test.sh`: WSL/Linux post-start health check
- `platform-skills-e2e.py`: real UI E2E for the platform skill reference-cycle flow
- `runtime-doctor.sh`: WSL/Linux runtime diagnostics for Docker dual-mode wiring
- `runtime-doctor.ps1`: Windows runtime diagnostics for Docker dual-mode wiring
- `docker-compose.yml`: local one-command stack

## What Gets Generated

The prepare scripts write ignored local files into `.generated/`:

- `casdoor/app.conf`
- `casdoor/init_data.json`
- `dotblue/config.yaml`
- `dotblue/init_data.json`
- They also append generated `DOTBLUE_CASDOOR_*` values to `.env`

## Quick Start

This Compose setup is optimized for local demo and quick testing:

- one `dotblue` service
- embedded worker loop inside that `dotblue` service
- one `redis` dependency for async session/queue/event flow
- one shared `postgres` service for both Casdoor and DotBlue
- no extra `worker` container to start or explain

1. Copy `.env.example` to `.env`
2. Edit `.env` and set:
   - `GO_MODULE_PROXY` and `NPM_REGISTRY` if your network environment needs mirrors
   - `POSTGRES_IMAGE` if you need a different PostgreSQL runtime, default is `postgres:18-alpine`
   - database passwords
   - `DOTBLUE_ADMIN_*`
   - `DOTBLUE_LLM_API_KEY` if needed
   - public URLs and ports
3. Generate local config files:

```bash
./prepare-config.sh
```

or on Windows:

```powershell
.\prepare-config.ps1
```

4. Start the stack:

```bash
docker compose up -d --build
```

If you want to test the S3 driver with the bundled MinIO service:

```bash
docker compose --profile s3 up -d --build
```

5. Run a smoke test:

```bash
./smoke-test.sh
```

6. Run the platform skill UI E2E when you need a real login + reference edit + publish-cycle regression check:

```powershell
python .\platform-skills-e2e.py
```

Notes:

- Start compose in WSL first, then run this script from Windows so Playwright can open `http://172.22.3.181:<port>` directly.
- The script reads `DOTBLUE_PUBLIC_URL`, `DOTBLUE_BACKEND_PUBLIC_URL`, `CASDOOR_PUBLIC_URL`, and admin credentials from `.env`.
- It reads the generated Casdoor OAuth client from `.generated/dotblue/config.yaml`.
- It saves HTML and screenshot artifacts under `.tmp/webapp-testing/platform-skills-e2e/`.
- It resets the edited skill references in a `finally` block so repeated runs start from a clean state.

7. If agent containers do not start, run runtime diagnostics:

```bash
./runtime-doctor.sh
```

or on Windows:

```powershell
.\runtime-doctor.ps1
```

## Services

- `postgres`: shared PostgreSQL for Casdoor and DotBlue
- `redis`: Redis for session control plane and dataplane
- `casdoor`: auth server with first-boot `init_data.json`
- `minio`: optional local S3-compatible object storage for driver verification, enabled with `--profile s3`
- `dotblue`: Go application service using generated runtime config and an embedded worker loop
- `web`: static frontend image built with generated `VITE_*` values

## Important Notes

- Run the prepare script before `docker compose up`
- `.env` and `.generated/` are ignored and should stay local
- `CASDOOR_PUBLIC_URL` is for browser access
- `CASDOOR_INTERNAL_URL` is for container-to-container access
- `DOTBLUE_BACKEND_INTERNAL_URL` defaults to `http://dotblue:8000`
- All published ports are user-configurable in `.env`
- `postgres:18-alpine` expects the data volume to be mounted at `/var/lib/postgresql`
- The generated DotBlue config enables `worker.embedded=true`, so local async chat works without a second app role
- File storage defaults to `local`, and the backend upload directory is persisted through `DOTBLUE_FILES_HOST_PATH -> DOTBLUE_FILES_LOCAL_ROOT`
- To switch to S3, set `DOTBLUE_FILES_DRIVER=s3` and fill `DOTBLUE_S3_*`
- For local MinIO testing, use `DOTBLUE_S3_ENDPOINT=http://minio:9000` and usually set `DOTBLUE_S3_FORCE_PATH_STYLE=true`; `.env.example` already aligns the sample bucket and credentials with the bundled MinIO defaults
- `DOTBLUE_S3_AUTO_CREATE_BUCKET=true` lets the backend create the bucket automatically on first upload, which is convenient for MinIO-based local verification
- When accessing from another machine or via host IP, set `CASDOOR_PUBLIC_URL`, `DOTBLUE_PUBLIC_URL`, and `DOTBLUE_BACKEND_PUBLIC_URL` to that reachable IP or domain instead of `localhost`
- The prepare scripts register multiple Casdoor OAuth callbacks by default: `${DOTBLUE_PUBLIC_URL}/callback`, `http://localhost:9000/callback`, and `http://127.0.0.1:9000/callback`
- Add more callbacks with `DOTBLUE_CASDOOR_EXTRA_REDIRECT_URIS`, using a comma-separated list in `.env`
- Container runtime mode auto-detects the host `docker.sock` group id during `prepare-config` and injects it into the `dotblue` service so non-root app processes can talk to Docker
- If auto-detection is unavailable in your environment, set `DOTBLUE_ENGINE_DOCKER_SOCKET_GID` manually in `.env` before running the prepare script
- `runtime-doctor` checks the effective runtime mode, dotblue container group wiring, `docker.sock` metadata, compose status, and currently running `hermes_*` containers

## Runtime Modes

- `runtimeMode=host`: suitable when the backend binary runs directly on the host and agent endpoints return `127.0.0.1:<random-port>`
- `runtimeMode=container`: suitable when the backend runs in Docker and reaches agent containers over Docker DNS
- `endpointMode=host_loopback`: publish the helper container to host loopback and let the backend call back through mapped host ports
- `endpointMode=docker_dns`: keep the helper container on the Docker network and let the backend call `http://hermes_<agent-id>:<port>`
- For the compose deployment in this folder, the recommended production default is `runtimeMode=container` with `endpointMode=docker_dns`
