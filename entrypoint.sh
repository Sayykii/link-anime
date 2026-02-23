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

# Add render group for Vulkan GPU access if /dev/dri exists
if [ -d /dev/dri ]; then
    # Get the GID of the render device (renderD128)
    RENDER_GID=$(stat -c '%g' /dev/dri/renderD128 2>/dev/null || true)
    if [ -n "$RENDER_GID" ] && [ "$RENDER_GID" != "0" ]; then
        # Create render group if it doesn't exist
        if ! getent group render >/dev/null 2>&1; then
            addgroup -g "$RENDER_GID" render 2>/dev/null || true
        fi
        # Add user to render group
        addgroup linkanime render 2>/dev/null || true
    fi
    # Also add to video group if it exists
    addgroup linkanime video 2>/dev/null || true
fi

# Ensure correct ownership of app data
chown -R "$PUID":"$PGID" /app/data

# Run as the created user
exec su-exec "$PUID":"$PGID" "$@"
