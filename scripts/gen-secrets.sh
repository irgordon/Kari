#!/bin/bash

# Karı Orchestration Engine - Secure Secret Generator
# 🛡️ SLA: Generate 2026-grade entropy for Zero-Trust boundaries

set -e

ENV_FILE=".env"

# 1. 🛡️ Guard against accidental overwrite of production secrets
if [ -f "$ENV_FILE" ]; then
    printf "⚠️  $ENV_FILE already exists. Overwrite? (y/N): "
    read -r response
    if [[ ! "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        echo "Aborting to protect existing secrets."
        exit 1
    fi
fi

echo "🔐 Karı Panel: Harvesting entropy for fresh secrets..."

# 2. 🛡️ AES-256-GCM Master Key
# Generates 32 bytes (256 bits) of random data -> 64 hex characters.
# Used by the Go Brain to wrap SSH keys and repository secrets.
ENC_KEY=$(openssl rand -hex 32)

# 3. 🛡️ JWT Signing Secret
# Generates a high-entropy string, cleaned of URL-unsafe characters.
JWT_SEC=$(openssl rand -base64 64 | tr -dc 'a-zA-Z0-9' | fold -w 64 | head -n 1)

# 4. 🛡️ PostgreSQL Root Password
# 24-character random alphanumeric string for the internal backplane.
DB_PASS=$(openssl rand -base64 32 | tr -dc 'a-zA-Z0-9' | fold -w 24 | head -n 1)

# 5. 🛡️ Write to .env with Strict Formatting
cat <<EOF > $ENV_FILE
# ==============================================================================
# 🛡️ KARİ PANEL AUTO-GENERATED SECRETS (2026-GRADE)
# ==============================================================================
DB_PASSWORD=$DB_PASS
ENCRYPTION_KEY=$ENC_KEY
JWT_SECRET=$JWT_SEC

# ==============================================================================
# 🧠 BRAIN CONFIGURATION
# ==============================================================================
PORT=8080
AGENT_SOCKET=/var/run/kari/agent.sock

# 🛡️ Peer Credential Guard: Must match the 'user' in docker-compose
KARI_EXPECTED_API_UID=1001

# ==============================================================================
# ⚙️ MUSCLE AGENT CONFIGURATION
# ==============================================================================
KARI_WEB_ROOT=/var/www/kari
KARI_SYSTEMD_DIR=/etc/systemd/system
RUST_LOG=info

# ==============================================================================
# 💻 UI / FRONTEND CONFIGURATION
# ==============================================================================
INTERNAL_API_URL=http://api:8080
PUBLIC_API_URL=http://localhost:8080
EOF

# 6. 🛡️ Hardening the file permissions
# 600 ensures only the current user (the one running Docker) can read the file.
chmod 600 $ENV_FILE

echo "✅ $ENV_FILE generated successfully."
echo "🛡️  Permissions set to 600 (Owner Read/Write Only)."
echo "--------------------------------------------------"
echo "🚀 NEXT STEP: Run 'make deploy' to launch the Karı Panel."
