<div align="center">
  
  
  <h1>Karı — Made Simple. Designed Secure. </h1>
  <p>A fast, friendly control panel that installs in minutes and makes server management effortless, safe, and actually enjoyable. Get powerful tools, a clean interface, and complete control without the clutter.</p>

  <p>
    <img src="https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white" alt="Go" />
    <img src="https://img.shields.io/badge/rust-%23000000.svg?style=for-the-badge&logo=rust&logoColor=white" alt="Rust" />
    <img src="https://img.shields.io/badge/svelte-%23f1413d.svg?style=for-the-badge&logo=svelte&logoColor=white" alt="Svelte" />
    <img src="https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL" />
    <img src="https://img.shields.io/badge/nginx-%23009639.svg?style=for-the-badge&logo=nginx&logoColor=white" alt="Nginx" />
    <br/>
    <img src="https://img.shields.io/badge/gRPC-%23244c5a.svg?style=for-the-badge&logo=grpc&logoColor=white" alt="gRPC" />
    <img src="https://img.shields.io/badge/GitHub_Actions-2088FF?style=for-the-badge&logo=github-actions&logoColor=white" alt="GitHub Actions" />
    <img src="https://img.shields.io/badge/license-MIT-blue.svg?style=for-the-badge" alt="MIT License" />
  </p>
</div>

---

Karı is a next-generation server control panel built for the workflows of 2026 and beyond. Designed to replace legacy monolithic panels, Karı brings the seamless, GitOps-driven developer experience of platforms like Vercel or Railway directly to your own infrastructure. 

Built with an unprivileged **Go** REST API and a memory-safe, root-level **Rust** system agent, Karı offers blisteringly fast performance and an impenetrable security boundary.

## ✨ Core Features

* **GitOps by Default:** Native webhooks for GitHub/GitLab. Push to `main`, and Karı automatically clones, builds, and swaps your app with zero-downtime symlinks.
* **Systemd User Jails:** First-class support for Node.js, Python, and Ruby. Apps run isolated under unprivileged system users with `ProtectSystem=full` and `PrivateTmp=true`, ensuring zero cross-tenant contamination.
* **Automated Auto-Renewing SSL:** Native Let's Encrypt integration. Certificates are provisioned securely in memory, written directly to root-owned files via Rust, and auto-renewed by a background Go worker 30 days before expiration.
* **Dynamic RBAC:** Shift beyond static roles. Create custom permission sets (e.g., "Junior Dev", "Auditor") with mathematical safeguards to prevent Super Admin lockouts.
* **Privacy-First Audit Logs:** Centralized PostgreSQL logging separates tenant activity from global server alerts, surfaced via a proactive UI Action Center.
* **Real-Time Observability:** End-to-end WebSockets stream deployment build logs directly to an XSS-proof `xterm.js` terminal UI in real-time.
* **Secure by Design:** Strict privilege separation. The API runs unprivileged; the Rust agent runs as root, bypassing shell execution entirely (no `bash` injection) and communicating exclusively via gRPC over a locked-down Unix Domain Socket.

---

## 🏗️ Architecture



Karı uses a Monorepo structure, split into three distinct boundaries:

1. **The UI (`/frontend`):** A decoupled, reactive Single Page Application built with SvelteKit.
2. **The Brain (`/api`):** A Go-based REST API that manages state in PostgreSQL, handles RBAC authentication, and orchestrates workflows following strict SOLID principles.
3. **The Muscle (`/agent`):** A Rust-based gRPC daemon running as root that executes highly validated system mutations (package management, systemd, file writes).

---

```markdown
# Karı Monorepo File Structure

kari/
├── .github/
│   └── workflows/
│       └── release.yml                 # CI/CD pipeline (Go build, Rust cross-compile, Svelte build)
├── agent/                              # The Muscle (Rust gRPC Daemon)
│   ├── build.rs                        
│   ├── Cargo.toml                      
│   └── src/
│       ├── main.rs                     # Entrypoint, secure Unix socket binding (0o660)
│       ├── server.rs                   # gRPC SystemAgent implementation 
│       └── sys/                        # System Integrations (SOLID SLAs)
│           ├── jail.rs                 # Linux user creation and filesystem lockdown
│           └── systemd.rs              # Generates secure systemd unit files (ProtectSystem=full)
├── api/                                # The Brain (Go REST API)
│   ├── cmd/
│   │   └── kari-api/
│   │       └── main.go                 # App entrypoint (wires dependencies, starts workers/router)
│   ├── internal/
│   │   ├── adapters/                   # Concrete implementations (SLA)
│   │   │   ├── nginx_manager.go        # text/template generation and Rust gRPC execution
│   │   │   └── acme_provider.go        # Let's Encrypt / lego adapter for SSL
│   │   ├── api/                        # HTTP Transport Layer
│   │   │   ├── handlers/
│   │   │   │   ├── application.go      
│   │   │   │   ├── websocket.go        
│   │   │   │   └── audit.go            # Privacy-first endpoints (tenant vs admin alerts)
│   │   │   ├── middleware/
│   │   │   │   └── auth.go             # JWT validation, Rate limiting, RequirePermission (RBAC)
│   │   │   └── router/
│   │   │       └── router.go           
│   │   ├── core/                       # Business Logic (SOLID)
│   │   │   ├── domain/                 # Structs & Repository Interfaces
│   │   │   │   ├── application.go      
│   │   │   │   ├── audit.go            
│   │   │   │   ├── ssl.go              
│   │   │   │   └── webserver.go        
│   │   │   ├── services/               # Orchestrators
│   │   │   │   ├── audit_service.go    # Enforces tenant data isolation
│   │   │   │   ├── ssl_service.go      # Orchestrates Let's Encrypt & Rust file writes
│   │   │   │   └── user_service.go     # RBAC logic (prevents Super Admin lockout)
│   │   │   └── utils/
│   │   │       └── cert_parser.go      # Reads PEM file expiration dates
│   │   ├── db/                         # Database Layer
│   │   │   ├── migrations/
│   │   │   │   ├── 001_initial_schema.sql # Postgres tables (users, domains, apps)
│   │   │   │   ├── 002_audit_logs.sql     # Centralized logging & system alerts
│   │   │   │   └── 003_dynamic_rbac.sql   # Roles, Permissions, and Mapping tables
│   │   │   └── postgres/
│   │   │       ├── application_repo.go 
│   │   │       └── audit_repo.go       # Dynamically built SQL queries for logs
│   │   ├── workers/
│   │   │   └── ssl_renewer.go          # Background cron job for automated certificate renewals
│   │   └── grpc/                       # Generated Go gRPC client (from proto)
├── frontend/                           # The UI (SvelteKit SPA)
│   ├── package.json
│   └── src/
│       ├── hooks.server.ts             # Server-side JWT gatekeeper, silent refresh logic
│       ├── lib/                        # Shared UI utilities and components
│       │   ├── api/                    # Frontend SLA Layer
│       │   │   ├── domains.ts          
│       │   │   └── terminalStream.ts   
│       │   └── components/             # SRP UI Components
│       │       ├── admin/
│       │       │   └── ActionCenter.svelte # Displays unresolved critical system alerts
│       │       ├── DeploymentTerminal.svelte 
│       │       └── DomainList.svelte   
│       └── routes/                     # Filesystem Routing
│           ├── (app)/                  # Authenticated routes 
│           │   ├── +layout.server.ts   
│           │   ├── dashboard/          # Includes ActionCenter for Admins
│           │   └── domains/            
│           └── (auth)/                 
│               └── login/              
├── proto/                              # The Contract
│   └── kari/agent/v1/agent.proto       
├── scripts/                            # DevOps & DX
│   ├── dev.sh                          
│   └── install.sh                      # Idempotent installer with CDN failover
├── docker-compose.yml                  
├── README.md                           
└── TECHNICAL_SPEC.md                   

```

---

## 🚀 Quick Install

To install Karı on a fresh Linux server, run our idempotent bootstrap script as `root`. This will automatically detect your OS, install baseline dependencies, configure PostgreSQL, and download the pre-compiled static binaries with an automatic CDN failover.

```bash
curl -sSL [https://raw.githubusercontent.com/irgordon/kari/main/scripts/install.sh](https://raw.githubusercontent.com/irgordon/kari/main/scripts/install.sh) | sudo bash

```

*(Supports Ubuntu 22.04/24.04, Debian 12, AlmaLinux 9, and Fedora)*

---

## 🛠️ Local Development

### Prerequisites

* Go 1.22+
* Rust (Stable) + Cargo
* Node.js 20+
* PostgreSQL 16+
* Protocol Buffers Compiler (`protoc`)

### Getting Started

1. **Clone the repository:**
```bash
git clone [https://github.com/irgordon/kari.git](https://github.com/irgordon/kari.git)
cd kari

```


2. **Generate the gRPC Protobufs:**
Ensure the contract between Go and Rust is up to date.
```bash
make proto-gen

```


3. **Start the development services:**
You can run the full stack locally using our provided script:
```bash
./scripts/dev.sh

```


* *Frontend:* `http://localhost:5173`
* *Go API:* `http://localhost:8080`



---

## 🛡️ Security

Security is the foundational principle of Karı. We utilize a strict two-token JWT architecture (HttpOnly cookies for the browser UI, and Personal Access Tokens for CLI usage).

If you discover a security vulnerability, please do **NOT** open a public issue. Email `security@kariapp.dev` directly.

---

## 📄 License

This project is licensed under the **[MIT License](https://mit-license.org/)**.

© 2026 Karı Project - *Made Simple. Designed Secure.*

```

```
