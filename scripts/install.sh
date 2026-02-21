#!/bin/bash
# ==============================================================================
# 🛡️ Kari Hardened Idempotent Installer
# ==============================================================================
set -euo pipefail

# --- Color formatting ---
TEAL='\033[0;36m'
GRAY='\033[1;30m'
GREEN='\033[0;32m'
NC='\033[0m'

# 1. 🛡️ Identity & Directory Provisioning
echo -e "${GRAY}[1/5] Provisioning secure users and paths...${NC}"

# Create the Brain's restricted identity
if ! id "kari-api" &>/dev/null; then
    useradd -r -s /bin/false kari-api
fi

# Ensure new files are never world-readable
umask 027

# Setup secure directories
mkdir -p /etc/kari/ssl
mkdir -p /var/run/kari
mkdir -p /var/www/kari
mkdir -p /opt/kari/bin

# 2. 🛡️ Permission Hardening
echo -e "${GRAY}[2/5] Enforcing Zero-Trust SLA boundaries...${NC}"

# Muscle (Rust) strictly owns the SSL storage. Go Brain cannot read this.
chown root:root /etc/kari/ssl
chmod 700 /etc/kari/ssl

# 🛡️ SLA Fix: The Socket Directory
# Root (Muscle) owns the directory to create the socket.
# The kari-api group (Brain) is granted read/write access to communicate.
chown root:kari-api /var/run/kari
chmod 750 /var/run/kari

# 3. 🛡️ Hardened Systemd Units
echo -e "${GRAY}[3/5] Deploying hardened service units...${NC}"

# Brain (Go API) Service - Maximum Restriction
cat <<EOF > /etc/systemd/system/kari-api.service
[Unit]
Description=Kari Go API Orchestrator
After=postgresql.service kari-agent.service

[Service]
ExecStart=/opt/kari/bin/kari-api
Restart=always
User=kari-api
Group=kari-api
EnvironmentFile=-/etc/kari/api.env

# 🛡️ Kari Zero-Trust Sandbox
# Make the entire OS read-only except the socket directory
ProtectSystem=strict
ReadWritePaths=/var/run/kari
ProtectHome=true
PrivateTmp=true
PrivateDevices=true
ProtectKernelTunables=true
ProtectControlGroups=true
RestrictSUIDSGID=true
NoNewPrivileges=true
# Drop ALL capabilities, the Brain needs none to route HTTP/gRPC
CapabilityBoundingSet=

[Install]
WantedBy=multi-user.target
EOF

# Muscle (Rust Agent) Service - Elevated but Scoped
cat <<EOF > /etc/systemd/system/kari-agent.service
[Unit]
Description=Kari Rust System Agent
After=network.target

[Service]
ExecStart=/opt/kari/bin/kari-agent
Restart=always
User=root
Group=root
EnvironmentFile=-/etc/kari/agent.env

# 🛡️ Muscle Sandbox
# Needs root, but does not need to mess with kernel modules or hostnames
ProtectKernelModules=true
ProtectHostname=true
RestrictRealtime=true

[Install]
WantedBy=multi-user.target
EOF

# 4. 🛡️ Binary Integrity
echo -e "${GRAY}[4/5] Securing executables...${NC}"
chown -R root:root /opt/kari/bin
chmod 755 /opt/kari/bin
chmod 700 /opt/kari/bin/kari-agent # Only root can execute
chmod 755 /opt/kari/bin/kari-api   # Brain needs rx

# 5. 🛡️ Daemon Reload & Enable
echo -e "${GRAY}[5/5] Reloading systemd daemon...${NC}"
systemctl daemon-reload
# systemctl enable kari-agent kari-api # Uncomment to enable on boot
# systemctl restart kari-agent kari-api # Uncomment to start immediately

echo -e "${TEAL}------------------------------------------------${NC}"
echo -e "${GREEN}✅ Kari Hardened Installation Complete!${NC}"
echo -e "${TEAL}------------------------------------------------${NC}"
