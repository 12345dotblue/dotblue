# Compose Quick Start

## Files

- `.env.example`: user-editable deployment variables
- `prepare-config.sh`: Linux/macOS config generator
- `prepare-config.ps1`: Windows PowerShell config generator
- `smoke-test.sh`: WSL/Linux post-start health check
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

5. Run a smoke test:

```bash
./smoke-test.sh
```

6. If agent containers do not start, run runtime diagnostics:

```bash
./runtime-doctor.sh
```

or on Windows:

```powershell
.\runtime-doctor.ps1
```

## Services

- `casdoor-db`: PostgreSQL for Casdoor
- `db`: PostgreSQL for dotblue backend
- `casdoor`: auth server with first-boot `init_data.json`
- `backend`: Go backend using generated runtime config
- `web`: static frontend image built with generated `VITE_*` values

## Important Notes

- Run the prepare script before `docker compose up`
- `.env` and `.generated/` are ignored and should stay local
- `CASDOOR_PUBLIC_URL` is for browser access
- `CASDOOR_INTERNAL_URL` is for container-to-container access
- `DOTBLUE_BACKEND_INTERNAL_URL` defaults to `http://backend:8000`
- All published ports are user-configurable in `.env`
- `postgres:18-alpine` expects the data volume to be mounted at `/var/lib/postgresql`
- When accessing from another machine or via host IP, set `CASDOOR_PUBLIC_URL`, `DOTBLUE_PUBLIC_URL`, and `DOTBLUE_BACKEND_PUBLIC_URL` to that reachable IP or domain instead of `localhost`
- Container runtime mode auto-detects the host `docker.sock` group id during `prepare-config` and injects it into the `backend` service so non-root backend processes can talk to Docker
- If auto-detection is unavailable in your environment, set `DOTBLUE_ENGINE_DOCKER_SOCKET_GID` manually in `.env` before running the prepare script
- `runtime-doctor` checks the effective runtime mode, backend container group wiring, `docker.sock` metadata, compose status, and currently running `hermes_*` containers

## Runtime Modes

- `runtimeMode=host`: suitable when the backend binary runs directly on the host and agent endpoints return `127.0.0.1:<random-port>`
- `runtimeMode=container`: suitable when the backend runs in Docker and reaches agent containers over Docker DNS
- `endpointMode=host_loopback`: publish the helper container to host loopback and let the backend call back through mapped host ports
- `endpointMode=docker_dns`: keep the helper container on the Docker network and let the backend call `http://hermes_<agent-id>:<port>`
- For the compose deployment in this folder, the recommended production default is `runtimeMode=container` with `endpointMode=docker_dns`
