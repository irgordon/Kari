#!/usr/bin/env bash
# Karı - Local Development Bootstrapper
# Refactored for Platform Agnosticism, SOLID, and SLA Compliance.

set -euo pipefail

# --- Color formatting ---
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# -----------------------------------------------------------------------------
# 1. Platform-Agnostic Setup
# -----------------------------------------------------------------------------
# Determine the absolute path of the project root regardless of OS
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
DEV_DIR="$ROOT_DIR/.local-dev"

# Detect Docker Compose command (support newer 'docker compose' and older v1)
DOCKER_COMPOSE_CMD="docker compose"
if ! docker compose version &>/dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
fi

echo -e "${CYAN}======================================================${NC}"
echo -e "${CYAN}🚀 Starting Karı Local Development Environment...${NC}"
echo -e "${CYAN}======================================================${NC}"

# -----------------------------------------------------------------------------
# 2. Pre-Flight Dependency Checks
# -----------------------------------------------------------------------------
check_cmd() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}❌ Missing dependency: $1 is required to run Karı locally.${NC}"
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
mkdir -p "$DEV_DIR/tmp" "$DEV_DIR/logs"

# Ensure frontend dependencies are installed
if [ ! -d "$ROOT_DIR/frontend/node_modules" ]; then
    echo -e "${YELLOW}📦 Installing SvelteKit dependencies...${NC}"
    (cd "$ROOT_DIR/frontend" && npm install)
fi

# Clean up stale socket file to prevent Rust Agent bind errors
export KARI_SOCKET_PATH="$DEV_DIR/tmp/agent.sock"
rm -f "$KARI_SOCKET_PATH"

# -----------------------------------------------------------------------------
# 4. Database Bootstrapping (Docker Compose)
# -----------------------------------------------------------------------------
echo -e "${YELLOW}🗄️ Starting Infrastructure via $DOCKER_COMPOSE_CMD...${NC}"
$DOCKER_COMPOSE_CMD up -d db # Targeted start of the 'db' service

# Wait for Postgres using a more platform-agnostic check
echo -e "${YELLOW}⏳ Waiting for database to accept connections...${NC}"
MAX_RETRIES=30
COUNT=0
until docker exec kari-db pg_isready -U kari_admin &>/dev/null || [ $COUNT -eq $MAX_RETRIES ]; do
    sleep 1
    ((COUNT++))
done

if [ $COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}❌ Database failed to start in time.${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Database is ready.${NC}"

# -----------------------------------------------------------------------------
# 5. Process Management & Cleanup (The Trap)
# -----------------------------------------------------------------------------
# Use a more robust PID management system
declare -a PIDS=()

cleanup() {
    echo -e "\n${RED}🛑 Shutting down Karı Development Environment...${NC}"
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            # Send SIGTERM for graceful shutdown
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
(cd "$ROOT_DIR/agent" && cargo run > "$DEV_DIR/logs/agent.log" 2>&1) &
PIDS+=($!)

# B. Start the Go API (The Brain)
# Using 'go run' directly for local dev to avoid path issues
echo -e "${YELLOW}🧠 Starting Go API...${NC}"
(cd "$ROOT_DIR/api" && \
 DATABASE_URL="postgres://kari_admin:dev_password_only@localhost:5432/kari?sslmode=disable" \
 JWT_SECRET="dev_secret_for_testing_only" \
 PORT="8080" \
 go run ./cmd/kari-api/main.go > "$DEV_DIR/logs/api.log" 2>&1) &
PIDS+=($!)

# C. Start the SvelteKit Frontend (The UI)
echo -e "${YELLOW}🎨 Starting SvelteKit UI...${NC}"
(cd "$ROOT_DIR/frontend" && npm run dev > "$DEV_DIR/logs/frontend.log" 2>&1) &
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
echo -e "Streaming logs... (Press ${RED}Ctrl+C${NC} to stop all services)"
echo ""

# SLA: Check if processes actually stayed alive
sleep 2
for pid in "${PIDS[@]}"; do
    if ! kill -0 "$pid" 2>/dev/null; then
        echo -e "${RED}❌ One of the services failed to start. Check .local-dev/logs/${NC}"
        cleanup
    fi
done

tail -f "$DEV_DIR/logs/agent.log" "$DEV_DIR/logs/api.log" "$DEV_DIR/logs/frontend.log"
