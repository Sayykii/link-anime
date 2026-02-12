#!/bin/sh
set -e

PUID=${PUID:-977}
PGID=${PGID:-988}

echo "Starting link-anime with UID=$PUID, GID=$PGID"

# Create group if it doesn't exist
if ! getent group linkanimegrp >/dev/null 2>&1; then
    addgroup -g "$PGID" linkanimegrp
fi

# Create user if it doesn't exist
if ! getent passwd linkanime >/dev/null 2>&1; then
    adduser -D -u "$PUID" -G linkanimegrp -h /app -s /sbin/nologin linkanime
fi

# Ensure correct ownership of app data
chown -R "$PUID":"$PGID" /app/data

# Run as the created user
exec su-exec "$PUID":"$PGID" "$@"
