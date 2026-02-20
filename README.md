# Kari

Modern, MIT-licensed server control plane with a split architecture:

- **Kari Brain** (`/brain`): Go orchestration and policy layer
- **Kari Agent** (`/agent`): Rust privileged execution layer
- **Kari Web** (`/web`): Next.js admin and customer UI

## Repository Layout

- `/brain` Go API, providers, pipelines, scheduler, auth, and audit
- `/agent` Rust gRPC worker and system command modules
- `/web` Next.js app, components, hooks, and API client
- `/docs` Architecture and protocol docs
- `/contracts` versioned cross-service protobuf contracts

## Getting Started (Scaffold Stage)

This repository is currently scaffolded for implementation.

1. Implement Brain APIs and orchestration pipelines in `/brain`.
2. Implement Agent gRPC handlers and command executors in `/agent`.
3. Implement Web UI flows against Brain APIs in `/web`.
4. Keep orchestration logic in Brain and privileged system operations in Agent.

See `/docs/ARCHITECTURE.md` for module boundaries and contracts.

## Current API Scaffold

- `POST /v1/sites/activate`
- `POST /v1/servers/onboard`
- `GET /healthz`
