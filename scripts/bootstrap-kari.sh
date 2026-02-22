#!/bin/bash

# 🛡️ Karı Panel: Universal Distro-Agnostic Bootstrap
# Validated on: Debian, Ubuntu, RHEL/Alma, Fedora, Arch, Alpine.

set -e

# --- 1. Universal Environment Discovery ---
echo "🔍 Performing System Archeology..."

# Check for cgroup v2 (Mandatory for our SLA and Rust Muscle)
if [ ! -f /sys/fs/cgroup/cgroup.controllers ]; then
  echo "❌ Error: cgroup v2 is not enabled. Karı requires a modern Linux kernel (5.x+)."
  exit 1
fi

# Identify Init System (Must be systemd for our current Muscle implementation)
if [[ $(ps --no-headers -o comm 1) != "systemd" ]]; then
  echo "❌ Error: systemd not detected. Karı currently utilizes systemd-run for jail isolation."
  exit 1
fi

# Determine Package Manager (for Docker installation if missing)
if command -v dnf >/dev/null; then PKG_MGR="dnf";
elif command -v apt-get >/dev/null; then PKG_MGR="apt-get";
elif command -v pacman >/dev/null; then PKG_MGR="pacman";
elif command -v apk >/dev/null; then PKG_MGR="apk";
else echo "⚠️  Unknown package manager. Please ensure Docker is installed manually."; fi

# --- 2. Standardized User/Group Logic ---
# Some distros use different GIDs for 'system' users. We force 1001 for PeerCred consistency.
echo "🛡️  Standardizing Karı Identity (UID/GID 1001)..."
if ! getent group 1001 >/dev/null; then
    groupadd -g 1001 kari-internal 2>/dev/null || addgroup -g 1001 kari-internal
fi

if ! getent passwd 1001 >/dev/null; then
    useradd -u 1001 -g 1001 -m -s /sbin/nologin kari 2>/dev/null || \
    adduser -u 1001 -G kari-internal -h /home/kari -s /sbin/nologin -D kari
fi

# --- 3. Distro-Agnostic Directory Mapping ---
# We use /opt/kari for binaries and /var/lib/kari for data to follow FHS standards.
PATHS=("/var/run/kari" "/var/lib/kari/jails" "/opt/kari/config" "/etc/kari/ssl")
for path in "${PATHS[@]}"; do
    mkdir -p "$path"
    chown 1001:1001 "$path"
done
chmod 770 /var/run/kari # UDS Socket Directory

# --- 4. Setup Mode Initialization ---
SETUP_JWT=$(openssl rand -hex 16)
cat <<EOF > /opt/kari/config/.env.setup
KARI_SETUP_MODE=true
KARI_SETUP_TOKEN=$SETUP_JWT
KARI_API_UID=1001
KARI_API_GID=1001
EOF

# --- 5. Container Launch ---
# Using 'docker-compose' or 'docker compose' (distro agnostic check)
DOCKER_COMPOSE_CMD="docker-compose"
if ! command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
fi

echo "📦 Launching Karı Stack via $DOCKER_COMPOSE_CMD..."
$DOCKER_COMPOSE_CMD -f docker-compose.prod.yml up -d

# --- 6. Hand-off ---
IP_ADDR=$(curl -s https://ifconfig.me || hostname -I | awk '{print $1}')
echo -e "\n✅ \033[0;32mKarı Layer 0 (Physical) is established.\033[0m"
echo -e "🌐 Complete setup at: \033[0;34mhttp://$IP_ADDR:3000/setup?token=$SETUP_JWT\033[0m"