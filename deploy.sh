#!/bin/bash
# deploy.sh — Push code and rebuild on the server
# Usage: ./deploy.sh [--no-cache]
set -euo pipefail

SERVER="root@media"
REPO_NAME="link-anime"

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
git push origin main 2>&1 || {
    warn "Nothing to push (already up to date)"
}

# Step 2: Find the repo on the server, pull, and rebuild
log "Connecting to ${SERVER}..."
ssh "$SERVER" bash -s -- "$REPO_NAME" "$NO_CACHE" <<'REMOTE_SCRIPT'
set -euo pipefail
REPO_NAME="$1"
NO_CACHE="${2:-}"

# Find the repo directory
REPO_DIR=""
for dir in \
    "/opt/dockhand/${REPO_NAME}" \
    "/opt/stacks/${REPO_NAME}" \
    "/root/${REPO_NAME}" \
    "/srv/${REPO_NAME}" \
    "/home/*/${REPO_NAME}"; do
    # Expand globs
    for expanded in $dir; do
        if [[ -d "$expanded" && -f "$expanded/compose.yaml" ]]; then
            REPO_DIR="$expanded"
            break 2
        fi
    done
done

# If not found in common locations, search for it
if [[ -z "$REPO_DIR" ]]; then
    echo "[deploy] Searching for ${REPO_NAME} repo..."
    REPO_DIR=$(find / -maxdepth 4 -name "compose.yaml" -path "*${REPO_NAME}*" -exec dirname {} \; 2>/dev/null | head -1)
fi

if [[ -z "$REPO_DIR" ]]; then
    echo "[deploy] ERROR: Could not find ${REPO_NAME} directory on server" >&2
    exit 1
fi

echo "[deploy] Found repo at: ${REPO_DIR}"
cd "$REPO_DIR"

# Pull latest code
echo "[deploy] Pulling latest code..."
git pull origin main

# Rebuild and restart
echo "[deploy] Rebuilding container..."
if [[ -n "$NO_CACHE" ]]; then
    docker compose build --no-cache
    docker compose up -d --force-recreate
else
    docker compose up -d --build --force-recreate
fi

# Show status
echo "[deploy] Container status:"
docker compose ps

# Show last few log lines
echo ""
echo "[deploy] Recent logs:"
docker compose logs --tail=10 --no-log-prefix
REMOTE_SCRIPT

log "Deploy complete!"
