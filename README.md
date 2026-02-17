# link-anime

A web app for hardlinking anime from a download directory into a Jellyfin-style media library. Replaces a 2800-line bash script with a proper UI.

Built with Go + Vue 3, Dockerized, designed for a single-user homelab setup. Integrates with qBittorrent, Shoko Server, and Nyaa RSS for a complete anime download-to-library workflow.

> This project was written almost entirely by AI (Claude) and is mostly for personal use. Your mileage may vary.

## What it does

- **Hardlinks** anime files from a downloads directory into organized show/season folders (same filesystem, no extra disk usage)
- **Filename parsing** extracts show name, season, episode number, quality, codec, etc.
- **Link wizard** — step-by-step UI: pick source files, choose type (series/movie), set show name and season, preview, confirm
- **Library browser** — view linked shows/movies, unlink individual seasons or entire shows
- **Hardlink safety** — warns before removing files that are the last remaining copy (nlink=1)
- **Undo** — revert the last link operation with one click
- **qBittorrent integration** — view active torrents with live progress updates via WebSocket, add torrents by magnet link, search Nyaa directly from the UI
- **RSS watch rules** — auto-download new episodes from Nyaa RSS based on configurable rules
- **Shoko Server integration** — trigger library scans after linking
- **Notifications** — Discord webhooks, ntfy, or generic webhook on link/download events
- **Download monitor** — polls qBit every 5s, broadcasts live progress via WebSocket, notifies on completion
- **Dark mode** with a Steel Ball Run (JoJo) inspired theme

## Stack

| Layer | Tech |
|-------|------|
| Backend | Go, chi router, gorilla/websocket, modernc.org/sqlite |
| Frontend | Vue 3, Vite, Tailwind CSS v4, shadcn-vue, Pinia, Vue Router |
| Auth | Single-user password with bcrypt + session cookies |
| Real-time | WebSocket for live torrent progress and link updates |
| Database | SQLite (settings, history, linked files, RSS rules) |
| Deploy | Docker multi-stage build (bun + go + alpine) |

## Setup

### Prerequisites

- Docker and Docker Compose
- A storage volume where both your downloads and media library live (for hardlinks to work, they must be on the same filesystem)

### Quick start

```bash
git clone https://github.com/Sayykii/link-anime.git
cd link-anime
cp .env.example .env
# Edit .env with your settings
docker compose up -d --build
```

The app will be available at `http://localhost:8787`.

### Configuration

Copy `.env.example` to `.env` and edit:

```env
# Container user/group IDs (match your media permissions)
PUID=977
PGID=988
TZ=Europe/Sofia

# App password
LA_PASSWORD=changeme

# Paths inside the container (relative to the /data volume mount)
LA_DOWNLOAD_DIR=/data/downloads/complete/anime
LA_MEDIA_DIR=/data/media/anime
LA_MOVIES_DIR=/data/media/anime-movies

# qBittorrent (optional)
LA_QBIT_URL=http://host.docker.internal:8080
LA_QBIT_USER=admin
LA_QBIT_PASS=yourpassword
LA_QBIT_CATEGORY=anime

# Shoko Server (optional)
LA_SHOKO_URL=http://host.docker.internal:8111
LA_SHOKO_APIKEY=your-api-key

# Notifications (optional)
LA_NOTIFY_URL=https://discord.com/api/webhooks/...
```

The volume mount in `compose.yaml` maps `/mnt/storage:/data` — adjust this to match your storage path. The key requirement is that downloads and media directories are on the **same filesystem** so hardlinks work.

If qBittorrent or Shoko Server run in separate Docker containers on the same host, use `host.docker.internal` to reach them (already configured via `extra_hosts` in compose).

### Settings

Most settings can also be configured from the Settings page in the UI after first login. Settings saved in the UI are stored in SQLite and override environment variables.

## License

MIT
