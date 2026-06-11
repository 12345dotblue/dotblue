# DotBlue

DotBlue is an enterprise AI assistant platform for teams that need agent management, runtime isolation, skill governance, enterprise administration, and private deployment in one stack.

Instead of stitching together a chat UI, auth, model settings, runtime hosting, IM integrations, and deployment scripts, DotBlue gives you a full-stack foundation for building, governing, and delivering AI assistants to real users.

## Why DotBlue

- Manage agents, conversations, skills, models, and enterprise access in one product
- Run each agent in an isolated containerized runtime for safer execution boundaries
- Support SaaS-style, self-hosted, and private deployment paths from the same codebase
- Connect assistants to external channels instead of limiting them to an internal web chat

## Who It Is For

- Product and platform teams shipping AI assistants to internal or external users
- Enterprise teams that need organizations, members, roles, invitations, and admin workflows
- Developers and operators who want a self-hosted AI agent platform with a real product surface
- Open-source builders who want a full-stack starting point instead of a thin SDK demo

## Quick Start

The recommended way to try DotBlue is Docker Compose. It gives you the most complete end-to-end setup with the lowest local complexity.

### Prerequisites

- Docker Engine with Docker Compose available
- Enough local resources to run `postgres`, `redis`, `casdoor`, `dotblue`, and `web`
- A valid LLM API key for the provider you want to test
- Free local ports for the default endpoints listed below

### Start With Compose

1. Go to the deployment directory.

```bash
cd deploy/compose
```

2. Copy the environment template.

```bash
cp .env.example .env
```

Windows PowerShell:

```powershell
Copy-Item .env.example .env
```

3. Update the key values in `.env`.

- Database passwords
- Platform admin account information
- `DOTBLUE_LLM_API_KEY`
- Public URLs and ports

4. Generate the local config files.

Linux/macOS:

```bash
./prepare-config.sh
```

Windows PowerShell:

```powershell
.\prepare-config.ps1
```

5. Start the full stack.

```bash
docker compose up -d --build
```

### Default Local Endpoints

- Web: `http://localhost:19000`
- Backend: `http://localhost:18080`
- Casdoor: `http://localhost:18000`

### What Compose Starts

- `postgres`
- `redis`
- `casdoor`
- `dotblue` in all-in-one mode, with the worker loop embedded
- `web`

### Validate The Setup

After the containers are healthy:

1. Open `http://localhost:19000`
2. Confirm that the landing page or login flow is reachable
3. Complete initialization through auto-install or the `/setup` page if needed
4. Log in and verify you can create an agent and open the chat page

For a more detailed walkthrough, see `deploy/compose/README.md`.

## What You Get

### Agent Management

- Create, edit, and delete agents
- Configure a name and system prompt for each agent
- Manage multiple agents for different roles and workflows

### Real-Time Chat

- Conversation list, message history, and auto-generated titles
- Streaming responses and visible thinking states
- Tool call and step records for better execution traceability

### Enterprise Administration

- Create and switch enterprises
- Manage members, org units, roles, and invitations
- Use an enterprise admin console for team collaboration and access control

### Platform Configuration

- Configure platform data paths and runtime ports
- Manage platform-wide model access and provider credentials
- Let platform admins manage global settings in one place

### IM Integration

- Configure and enable enterprise IM connections
- Bind IM channels to agents
- Route assistant interactions through channels such as Feishu, Slack, Telegram, Discord, QQ, Matrix, and Web

### Skill Management

- Create platform skills from scratch
- Import external skills from connected hubs
- Govern the skill lifecycle from catalog to agent installation
- Inject installed skills into both Hermes and Nanobot runtimes

## Current Support

- Backend: Go 1.25, GoFrame, PostgreSQL
- Frontend: React 19, TypeScript, Vite, Ant Design
- Authentication: Casdoor
- Runtime isolation: Docker-based agent runtimes
- Models: centralized platform configuration with OpenAI-compatible and provider-specific runtime handling
- Testing: Go `*_test.go`, Vitest

## Repository Structure

```text
.
├─ backend/          # Go backend: APIs, domain services, schema init, auto-install
├─ web/              # React frontend: landing page, login, setup, dashboard, chat, admin UI
├─ deploy/           # Deployment assets: Casdoor, DotBlue, and Compose quick start
├─ docs/             # Domain and product design notes
└─ README.md
```

Useful directories:

- `backend/internal/domains`: core business domains such as `agent`, `chat`, `conversation`, `enterprise`, `im`, `metering`, and `skill`
- `backend/manifest/config`: runtime and initialization config templates
- `web/src/domains`: frontend pages and business modules
- `deploy/compose`: one-command local or test environment startup

## Local Development

### Frontend

1. Go to `web/`
2. Copy `.env.example` to `.env.local`
3. Configure Casdoor and backend URLs
4. Start the dev server

```bash
cd web
npm install
npm run dev
```

Common commands:

```bash
npm run build
npm run lint
npm run test
```

### Backend

1. Go to `backend/`
2. Prepare PostgreSQL
3. Copy `manifest/config/config.example.yaml` to `manifest/config/config.yaml`
4. If you want auto-install, also copy `manifest/config/init_data.example.json` to `manifest/config/init_data.json`
5. Set the required environment variables
6. Start the service

```bash
cd backend
go run .
```

Recommended for first-time setup:

```bash
export DOTBLUE_ADMIN_PASSWORD='replace-with-a-strong-password'
export DOTBLUE_LLM_API_KEY='replace-with-provider-key'
```

Windows PowerShell:

```powershell
$env:DOTBLUE_ADMIN_PASSWORD='replace-with-a-strong-password'
$env:DOTBLUE_LLM_API_KEY='replace-with-provider-key'
go run .
```

At startup, the backend automatically performs:

- Casdoor initialization
- Database connectivity checks
- Database schema initialization
- Auto-install based on `init_data.json`

## Configuration

### Backend

- Runtime template: `backend/manifest/config/config.example.yaml`
- Initialization template: `backend/manifest/config/init_data.example.json`
- Recommended local files:
  - `backend/manifest/config/config.yaml`
  - `backend/manifest/config/init_data.json`

Notes:

- `organization.name` and `application.name` must match the backend Casdoor configuration
- Do not commit sensitive values; inject them through environment variables when possible
- If no initialization file is provided, the Web setup page at `/setup` still works as a fallback

### Frontend

Key environment variables:

- `VITE_CASDOOR_SERVER_URL`
- `VITE_CASDOOR_CLIENT_ID`
- `VITE_CASDOOR_ORG_NAME`
- `VITE_CASDOOR_APP_NAME`
- `VITE_BACKEND_URL`

The container image also supports runtime overrides through `/runtime-config.js`, so private deployments can reuse the same `dotblue-web` image without rebuilding for each environment.

## API and UI Overview

The frontend includes:

- Landing page
- Login and login callback
- Setup wizard
- Dashboard and agent management
- Chat page
- Enterprise admin console
- Platform settings page
- Skill market, skill builder, skill governance, and dedicated agent skill management pages

The backend exposes:

- `/api/setup/*`: setup and installation
- `/api/agents`: agent CRUD
- `/api/conversations`: conversations and messages
- `/api/chat`: chat interface
- `/api/admin/settings`: platform settings
- `/api/admin/*`: enterprise and IM administration
- `/api/admin/skills*`, `/api/admin/platform/skill-hubs*`, `/api/admin/platform/skill-import-jobs*`: skill governance, connected hubs, and import workflows
- `/api/admin/agents/:agentId/skills*`: install, list, and remove skills on a target agent

Also available by default:

- OpenAPI: `/api.json`
- Swagger: `/swagger`

## Known Caveats

- If you update the platform-level model or provider settings after agent containers have already started, recycle or restart the existing `hermes_*` agent containers before re-testing. Running agent containers keep their existing runtime config until they are recreated.
- For skill-related end-to-end checks, prefer the local Compose environment and real browser verification. A typical validation flow is: import or create a skill -> install it to an agent -> open the chat page -> confirm the assistant actually uses the installed skill.

## Documentation Map

- Root overview: `README.md`
- Compose deployment guide: `deploy/compose/README.md`
- Backend setup notes: `backend/README.MD`
- Frontend environment notes: `web/README.md`
- Domain design docs:
  - `docs/skill-system-design.md`
  - `docs/skill-backend-design.md`
  - `docs/skill-frontend-design.md`
  - `docs/skill-database-design.md`
  - `docs/credit-domain-design.md`

## CI and Container Images

Every push and pull request runs `.github/workflows/ci.yml`, which lints, type-checks, builds, and tests both the backend and the web app, and builds both Docker images as a smoke test.

Cutting a tag of the form `vX.Y.Z` (for example `v0.1.0`) triggers `.github/workflows/release.yml`, which:

- builds and pushes release tags of the backend and web images to GitHub Container Registry
- drafts a GitHub release with auto-generated notes

Once a tag is pushed, users can pull pre-built images directly instead of building from source. Replace `<owner>` with the GitHub owner of this repository:

```bash
docker pull ghcr.io/<owner>/dotblue:v0.1.0
docker pull ghcr.io/<owner>/dotblue-web:v0.1.0
```

The backend image is published as `dotblue`, and the frontend image is published as `dotblue-web`.
