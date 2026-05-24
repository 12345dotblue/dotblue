# DotBlue

DotBlue is an enterprise-grade AI assistant governance platform for teams that want to manage AI assistants with less operational overhead.

It combines enterprise AI assistant management, real-time chat, enterprise administration, platform-wide LLM configuration, IM integration, and deployment workflows in one stack, so teams can build, govern, and deliver AI assistants instead of stitching together disconnected tools.

DotBlue supports self-hosted and cloud-hosted deployment patterns, automatically hosts isolated assistant runtimes, and centralizes LLM provider credentials and access token management at the platform level. It already works with OpenAI-compatible providers and is designed to incrementally support more mainstream AI assistants, models, and enterprise channels over time.

Keywords: enterprise AI assistant platform, AI assistant governance platform, self-hosted AI agent platform, cloud-hosted AI assistant platform, private deployment AI platform, multi-agent platform, LLM ops platform, AI workspace, AI agent management, enterprise chatbot platform, IM-integrated assistant platform, LLM token management.

## Tagline

- Govern enterprise AI assistants with less operational overhead
- Run assistants in self-hosted or cloud-hosted environments with automated isolated runtime hosting
- Support mainstream AI assistant and model ecosystems incrementally
- Centralize LLM provider credentials, access tokens, and platform-wide management in one place

## Why DotBlue

- Built for real teams: more than a chat UI, with organizations, members, invitations, and admin capabilities
- Built for safer runtime isolation: each agent runs in its own container sandbox with isolated runtime data
- Built for extensibility: supports external LLM providers and includes IM integration capabilities
- Built for deployment: ships with frontend, backend, authentication, databases, and Docker Compose setup

## What Makes It Useful

- Self-hosted and private-deployment friendly for internal enterprise environments
- Multi-agent ready, with clear boundaries for agent management and runtime orchestration
- Enterprise-aware, with organizations, members, roles, invitations, and admin workflows
- LLM-provider aware, with centralized platform configuration for model access
- Extensible toward more popular assistants, models, and message channels as the project evolves

## Core Capabilities

### 1. Agent Management

- Create, edit, and delete agents
- Configure a name and system prompt for each agent
- Manage multiple agents for different roles and workflows

### 2. Real-Time Chat

- Conversation list, message history, and auto-generated titles
- Streaming responses and visible thinking states
- Tool call and step records for better execution traceability

### 3. Enterprise Administration

- Create and switch enterprises
- Manage members, org units, roles, and invitations
- Use an enterprise admin console for team collaboration and access control

### 4. Platform Configuration

- Configure platform data paths and runtime ports
- Configure OpenAI, Anthropic, and other LLM providers
- Let platform admins manage global settings in one place

### 5. IM Integration

- Configure and enable enterprise IM connections
- Bind IM channels to agents
- Includes Feishu-related integration as an external message entry point

## Who It Is For

- Product and platform teams that want to deliver AI agents to real users
- Internal enterprise use cases that need organizations, permissions, and member management
- Developers and operators who want to run a self-hosted AI agent platform
- Open-source builders looking for a clean full-stack foundation for further customization

## Tech Stack

- Backend: Go 1.25, GoFrame, PostgreSQL
- Frontend: React 19, TypeScript, Vite, Ant Design
- Authentication: Casdoor
- Runtime: Docker
- Testing: Go `*_test.go`, Vitest

## Repository Structure

```text
.
├─ backend/          # Go backend: APIs, domain services, schema init, auto-install
├─ web/              # React frontend: login, setup wizard, dashboard, chat, admin UI
├─ deploy/           # Deployment assets: Casdoor, DotBlue, and Compose quick start
└─ README.md
```

More detail:

- `backend/internal/domains`: core business domains such as `agent`, `chat`, `conversation`, `enterprise`, and `im`
- `backend/manifest/config`: runtime and initialization config templates
- `web/src/domains`: frontend pages and business modules
- `deploy/compose`: one-command local or test environment startup

## Quick Start

The recommended way to get started is Docker Compose. It gives you the most complete end-to-end setup with the lowest local complexity.

Important reminder:

- If you update the platform-level LLM provider in the admin UI after agent containers have already started, recycle or restart the existing `hermes_*` agent containers before re-testing. Running agent containers keep their existing runtime config until they are recreated.

### Option 1: Docker Compose

1. Go to the deployment directory

```bash
cd deploy/compose
```

2. Copy the environment template

```bash
cp .env.example .env
```

Windows PowerShell:

```powershell
Copy-Item .env.example .env
```

3. Update the key values in `.env`

- Database passwords
- Platform admin account information
- `DOTBLUE_LLM_API_KEY`
- Public URLs and ports

4. Generate local config files

Linux/macOS:

```bash
./prepare-config.sh
```

Windows PowerShell:

```powershell
.\prepare-config.ps1
```

5. Start the full stack

```bash
docker compose up -d --build
```

Default local endpoints:

- Web: `http://localhost:19000`
- Backend: `http://localhost:18080`
- Casdoor: `http://localhost:18000`

The Compose setup starts:

- `postgres`
- `redis`
- `casdoor`
- `dotblue` in all-in-one mode, with the worker loop embedded
- `web`

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

- Template: `backend/manifest/config/config.example.yaml`
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

## API and UI Overview

The frontend includes:

- Landing page
- Login and login callback
- Setup wizard
- Dashboard and agent management
- Chat page
- Enterprise admin console
- Platform settings page

The backend exposes:

- `/api/setup/*`: setup and installation
- `/api/agents`: agent CRUD
- `/api/conversations`: conversations and messages
- `/api/chat`: chat interface
- `/api/admin/settings`: platform settings
- `/api/admin/*`: enterprise and IM administration

Also available by default:

- OpenAPI: `/api.json`
- Swagger: `/swagger`

## Highlights

- Clear separation between frontend, backend, domain services, and deployment assets
- A usable product foundation, not just an SDK or API demo
- Covers local development, automatic installation, and Compose deployment in one repository

## Open Source Checklist

If you plan to publish DotBlue as an open-source project, it is worth adding:

- `LICENSE`
- `CONTRIBUTING.md`
- `CODE_OF_CONDUCT.md`
- `SECURITY.md`

These files make the repository more complete and more welcoming for contributors.
