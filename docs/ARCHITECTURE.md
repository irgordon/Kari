# Kari System Architecture

Kari is a hybrid Go + Rust server control panel built with strict separation of concerns:

- `Brain` (Go): orchestration, policy, API, auth, audit, scheduling
- `Agent` (Rust): privileged execution, system mutation, package/service operations
- `Web` (Next.js): operator and tenant UI

## Runtime Topology

- Web communicates with Brain over HTTP APIs.
- Brain communicates with Agent over gRPC (mTLS or UNIX socket).
- Agent performs privileged host operations and returns structured results.

## Module Boundaries

### Brain (`/brain`)

- `api`: request handlers and API contracts
- `providers`: DNS/CDN/SSL and external integrations
- `models`: domain entities
- `pipelines`: orchestration flows (site create, backup, restore)
- `scheduler`: periodic jobs and timers
- `auth`: MFA, session, RBAC
- `audit`: immutable activity/event logs

### Agent (`/agent`)

- `grpc`: server and protobuf handlers
- `commands`: command dispatch and execution wrappers
- `nginx`, `mail`, `db`, `firewall`, `files`, `metrics`, `backup`: infrastructure modules

### Web (`/web`)

- `app`: routes and pages
- `components`: UI primitives and domain components
- `hooks`: state and API hooks
- `api-client`: typed client for Brain APIs

## Architectural Rules

1. Brain never performs raw privileged host mutation.
2. Agent never decides product policy; it executes explicit commands.
3. Pipelines are deterministic and auditable.
4. All cross-boundary contracts are versioned.
