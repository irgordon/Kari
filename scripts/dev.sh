#!/usr/bin/env bash
# Karı - Local Development Bootstrapper
# Hardened for 2026 SLA Compliance & Deterministic Execution.

set -euo pipefail

# --- Color formatting ---
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m'

# -----------------------------------------------------------------------------
# 1. Platform-Agnostic Setup
# -----------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
DEV_DIR="$ROOT_DIR/.local-dev"

DOCKER_COMPOSE_CMD="docker compose"
if ! docker compose version &>/dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
fi

echo -e "${CYAN}======================================================${NC}"
echo -e "${CYAN}🚀 Starting Karı Local Development Environment...${NC}"
echo -e "${CYAN}======================================================${NC}"

# -----------------------------------------------------------------------------
# 2. Pre-Flight Checks
# -----------------------------------------------------------------------------
check_cmd() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}❌ Missing dependency: $1 is required.${NC}"
        exit 1
    fi
}

echo -e "${YELLOW}🔍 Checking prerequisites...${NC}"
check_cmd "docker"
check_cmd "go"
check_cmd "cargo"
check_cmd "npm"
echo -e "${GREEN}✅ All dependencies found.${NC}"

# -----------------------------------------------------------------------------
# 3. Local Environment Preparation
# -----------------------------------------------------------------------------
# 🛡️ SLA: Create mock OS directories so the Rust Agent doesn't need root locally
mkdir -p "$DEV_DIR/tmp" "$DEV_DIR/logs" "$DEV_DIR/etc/nginx/sites-enabled" "$DEV_DIR/etc/ssl/kari"

if [ ! -d "$ROOT_DIR/frontend/node_modules" ]; then
    echo -e "${YELLOW}📦 Installing SvelteKit dependencies...${NC}"
    (cd "$ROOT_DIR/frontend" && npm install)
fi

export KARI_SOCKET_PATH="$DEV_DIR/tmp/agent.sock"
rm -f "$KARI_SOCKET_PATH"

# -----------------------------------------------------------------------------
# 4. Database Bootstrapping
# -----------------------------------------------------------------------------
echo -e "${YELLOW}🗄️ Starting Infrastructure via $DOCKER_COMPOSE_CMD...${NC}"
# Note: Ensure your docker-compose exposes 5432 for this local-host dev script!
$DOCKER_COMPOSE_CMD up -d db

echo -e "${YELLOW}⏳ Waiting for database to accept connections...${NC}"
MAX_RETRIES=30
COUNT=0
# 🛡️ Stability: Use docker compose exec to ensure we are talking to the right project network
until $DOCKER_COMPOSE_CMD exec db pg_isready -U kari_admin &>/dev/null || [ $COUNT -eq $MAX_RETRIES ]; do
    sleep 1
    ((COUNT++))
done

if [ $COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}❌ Database failed to start in time.${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Database is ready.${NC}"

# -----------------------------------------------------------------------------
# 5. Process Management (The Trap)
# -----------------------------------------------------------------------------
declare -a PIDS=()

cleanup() {
    echo -e "\n${RED}🛑 Shutting down Karı Development Environment...${NC}"
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid" 2>/dev/null || true
        fi
    done
    $DOCKER_COMPOSE_CMD stop db
    echo -e "${GREEN}✅ All processes terminated gracefully. Goodbye!${NC}"
    exit 0
}

trap cleanup SIGINT SIGTERM EXIT

# -----------------------------------------------------------------------------
# 6. Launch the Monorepo Services
# -----------------------------------------------------------------------------

# A. Start the Rust Agent (The Muscle)
echo -e "${YELLOW}⚙️ Starting Rust Agent...${NC}"
(cd "$ROOT_DIR/agent" && \
 KARI_SOCKET_PATH="$KARI_SOCKET_PATH" \
 KARI_SSL_DIR="$DEV_DIR/etc/ssl/kari" \
 NGINX_VHOST_DIR="$DEV_DIR/etc/nginx/sites-enabled" \
 RUST_LOG="debug" \
 cargo run > "$DEV_DIR/logs/agent.log" 2>&1) &
PIDS+=($!)

# B. Start the Go API (The Brain)
echo -e "${YELLOW}🧠 Starting Go API...${NC}"
(cd "$ROOT_DIR/api" && \
 DATABASE_URL="postgres://kari_admin:dev_password_only@localhost:5432/kari?sslmode=disable" \
 JWT_SECRET="dev_secret_for_testing_only" \
 AGENT_SOCKET="$KARI_SOCKET_PATH" \
 PORT="8080" \
 go run ./cmd/kari-api/main.go > "$DEV_DIR/logs/api.log" 2>&1) &
PIDS+=($!)

# C. Start the SvelteKit Frontend (The UI)
echo -e "${YELLOW}🎨 Starting SvelteKit UI...${NC}"
(cd "$ROOT_DIR/frontend" && \
 INTERNAL_API_URL="http://localhost:8080" \
 PUBLIC_API_URL="http://localhost:8080" \
 npm run dev > "$DEV_DIR/logs/frontend.log" 2>&1) &
PIDS+=($!)

# -----------------------------------------------------------------------------
# 7. Monitoring & Logging
# -----------------------------------------------------------------------------
echo -e "${GREEN}======================================================${NC}"
echo -e "${GREEN}🎉 Karı is running locally!${NC}"
echo -e "    🌐 UI:       ${CYAN}http://localhost:5173${NC}"
echo -e "    🔌 API:      ${CYAN}http://localhost:8080${NC}"
echo -e "    🛡️ Socket:   ${CYAN}$KARI_SOCKET_PATH${NC}"
echo -e "${GREEN}======================================================${NC}"
echo -e "Streaming logs... (Press ${RED}Ctrl+C${NC} to stop all services)\n"

sleep 2
for pid in "${PIDS[@]}"; do
    if ! kill -0 "$pid" 2>/dev/null; then
        echo -e "${RED}❌ A service crashed immediately. Check .local-dev/logs/${NC}"
        cleanup
    fi
done

tail -f "$DEV_DIR/logs/agent.log" "$DEV_DIR/logs/api.log" "$DEV_DIR/logs/frontend.log"
