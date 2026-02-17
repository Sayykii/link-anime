#!/bin/bash
# deploy.sh — Push code and rebuild on the server
# Usage: ./deploy.sh [--no-cache]
set -euo pipefail

SERVER="root@media"
REPO_NAME="link-anime"
REMOTE_REPO_DIR="/opt/stacks/${REPO_NAME}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[deploy]${NC} $*"; }
warn() { echo -e "${YELLOW}[deploy]${NC} $*"; }
err() { echo -e "${RED}[deploy]${NC} $*" >&2; }

# Parse args
NO_CACHE=""
if [[ "${1:-}" == "--no-cache" ]]; then
    NO_CACHE="--no-cache"
    warn "Full rebuild requested (--no-cache)"
fi

# Step 1: Push to GitHub
log "Pushing to GitHub..."
git push origin main 2>&1 || true

# Step 2: Ensure repo exists on server
log "Ensuring repo exists on server..."
ssh "$SERVER" bash -s -- "$REPO_NAME" "$REMOTE_REPO_DIR" <<'ENSURE_SCRIPT'
set -euo pipefail
REPO_NAME="$1"
REPO_DIR="$2"

if [[ -d "$REPO_DIR/.git" ]]; then
    echo "[deploy] Repo already exists at ${REPO_DIR}"
else
    echo "[deploy] Cloning repo to ${REPO_DIR}..."
    mkdir -p "$(dirname "$REPO_DIR")"
    git clone "https://github.com/Sayykii/${REPO_NAME}.git" "$REPO_DIR"
    echo "[deploy] Cloned successfully"
fi
ENSURE_SCRIPT

# Step 3: Ensure .env exists on server (copy local one if missing)
HAS_ENV=$(ssh "$SERVER" "[ -f '${REMOTE_REPO_DIR}/.env' ] && echo yes || echo no")
if [[ "$HAS_ENV" == "no" ]]; then
    if [[ -f "${SCRIPT_DIR}/.env" ]]; then
        log "Copying .env to server..."
        scp "${SCRIPT_DIR}/.env" "${SERVER}:${REMOTE_REPO_DIR}/.env"
        log ".env copied. Edit it on the server if needed: ${REMOTE_REPO_DIR}/.env"
    else
        err "No .env found locally or on server! Create one at ${REMOTE_REPO_DIR}/.env"
        exit 1
    fi
else
    log ".env already exists on server"
fi

# Step 4: Pull latest code, rebuild, and restart
log "Deploying..."
ssh "$SERVER" bash -s -- "$REMOTE_REPO_DIR" "$NO_CACHE" <<'DEPLOY_SCRIPT'
set -euo pipefail
REPO_DIR="$1"
NO_CACHE="${2:-}"

cd "$REPO_DIR"

echo "[deploy] Pulling latest code..."
git pull origin main

echo "[deploy] Rebuilding container..."
if [[ -n "$NO_CACHE" ]]; then
    docker compose build --no-cache
    docker compose up -d --force-recreate
else
    docker compose up -d --build --force-recreate
fi

echo ""
echo "[deploy] Container status:"
docker compose ps

echo ""
echo "[deploy] Recent logs:"
docker compose logs --tail=15 --no-log-prefix
DEPLOY_SCRIPT

log "Deploy complete!"
