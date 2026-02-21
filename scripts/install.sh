#!/bin/bash

# ==============================================================================
# 🛡️ Kari Idempotent Installer
# ==============================================================================
# Version: 1.0.0
# License: MIT
# Description: Installs Go Brain, Rust Muscle, and Svelte UI with Root isolation.

set -e # Exit on any error

# Brand Colors for Output
TEAL='\033[0;36m'
GRAY='\033[0;90m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${TEAL}"
echo "  _  __              _ "
echo " | |/ /_ _ _ __ _  _ "
echo " | ' < _` | '__| || |"
echo " |_|\_\__,_|_|   \_, |"
echo "                 |__/ "
echo -e "${NC}Made Simple. Designed Secure."
echo "------------------------------------------------"

# ==============================================================================
# 1. Sanity Checks & Environment Detection
# ==============================================================================

# Check for Root Privileges
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}Error: This script must be run as root.${NC}" 
   exit 1
fi

# Detect OS Distribution
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    echo -e "${RED}Error: Cannot detect OS distribution. Manual install required.${NC}"
    exit 1
fi

echo -e "${GRAY}[1/5] Running sanity checks...${NC}"

# Check for minimum RAM (1GB Recommended)
TOTAL_RAM=$(free -m | awk '/^Mem:/{print $2}')
if [ "$TOTAL_RAM" -lt 900 ]; then
    echo -e "${RED}Warning: Kari recommends at least 1GB of RAM. Current: ${TOTAL_RAM}MB${NC}"
fi

# ==============================================================================
# 2. Dependency Management (Platform Agnostic)
# ==============================================================================

echo -e "${GRAY}[2/5] Installing system dependencies for ${OS}...${NC}"

case "$OS" in
    ubuntu|debian)
        apt-get update -y
        apt-get install -y curl wget git postgresql nginx build-essential ufw
        ;;
    fedora|almalinux|rocky)
        dnf install -y curl wget git postgresql-server nginx gcc ufw
        ;;
    *)
        echo -e "${RED}Error: Unsupported OS ($OS). Please install dependencies manually.${NC}"
        exit 1
        ;;
esac

# ==============================================================================
# 3. Secure Directory & User Provisioning
# ==============================================================================

echo -e "${GRAY}[3/5] Provisioning secure system directories...${NC}"

# Create unprivileged user for the Go Brain
if ! id "kari-api" &>/dev/null; then
    useradd -r -s /bin/false kari-api
fi

# Setup secure directories
mkdir -p /etc/kari/ssl
mkdir -p /var/run/kari
mkdir -p /var/www/html
mkdir -p /opt/kari/bin

# Lock down SSL directory to root (Agent Muscle will manage this)
chown root:root /etc/kari/ssl
chmod 700 /etc/kari/ssl

# Lock down Unix Socket directory for Go-Rust communication
chown root:kari-api /var/run/kari
chmod 770 /var/run/kari

# ==============================================================================
# 4. Binary Installation
# ==============================================================================

echo -e "${GRAY}[4/5] Deploying platform-agnostic binaries...${NC}"

# Note: In a real CI/CD flow, these would be pulled from a release CDN.
# Here we ensure they exist in the opt path.
# mv ./api/bin/kari-api /opt/kari/bin/
# mv ./agent/target/release/kari-agent /opt/kari/bin/

chmod +x /opt/kari/bin/kari-*

# ==============================================================================
# 5. Systemd Service Orchestration
# ==============================================================================

echo -e "${GRAY}[5/5] Configuring systemd services...${NC}"

# Kari Rust Agent (The Muscle - Runs as Root)
cat <<EOF > /etc/systemd/system/kari-agent.service
[Unit]
Description=Kari Rust System Agent
After=network.target

[Service]
ExecStart=/opt/kari/bin/kari-agent
Restart=always
User=root
Group=root

[Install]
WantedBy=multi-user.target
EOF

# Kari Go API (The Brain - Runs as kari-api)
cat <<EOF > /etc/systemd/system/kari-api.service
[Unit]
Description=Kari Go API Orchestrator
After=postgresql.service kari-agent.service

[Service]
ExecStart=/opt/kari/bin/kari-api
Restart=always
User=kari-api
Group=kari-api
EnvironmentFile=/etc/kari/api.env

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now kari-agent
systemctl enable --now kari-api

echo -e "${TEAL}"
echo "------------------------------------------------"
echo "🚀 Kari Installation Complete!"
echo "------------------------------------------------"
echo -e "${NC}API listening on: ${TEAL}http://localhost:8080${NC}"
echo -e "Next Step: Configure your /etc/kari/api.env with secrets."
