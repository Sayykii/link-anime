import type { LinkRequest, LinkResult, LibraryStats, Show, Movie, DownloadItem, HistoryEntry, ParseResult, Settings, TorrentStatus, NyaaResult } from '@/lib/types'

class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const opts: RequestInit = {
    method,
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
  }
  if (body !== undefined) {
    opts.body = JSON.stringify(body)
  }

  const resp = await fetch(`/api${path}`, opts)

  if (resp.status === 401) {
    // Redirect to login if unauthorized
    if (window.location.pathname !== '/login') {
      window.location.href = '/login'
    }
    throw new ApiError('Unauthorized', 401)
  }

  const data = await resp.json()

  if (!resp.ok) {
    throw new ApiError(data.error || 'Request failed', resp.status)
  }

  return data as T
}

export function useApi() {
  return {
    // Auth
    login: (password: string) => request<{ ok: boolean }>('POST', '/auth/login', { password }),
    logout: () => request<{ ok: boolean }>('POST', '/auth/logout'),
    checkAuth: () => request<{ authenticated: boolean }>('GET', '/auth/check'),

    // Library
    getShows: () => request<Show[]>('GET', '/library/shows'),
    getMovies: () => request<Movie[]>('GET', '/library/movies'),
    getStats: () => request<LibraryStats>('GET', '/library/stats'),

    // Downloads
    getDownloads: () => request<DownloadItem[]>('GET', '/downloads'),
    parseRelease: (name: string) => request<ParseResult>('GET', `/downloads/parse?name=${encodeURIComponent(name)}`),

    // Link operations
    link: (req: LinkRequest) => request<LinkResult>('POST', '/link', req),
    linkPreview: (req: LinkRequest) => request<LinkResult>('POST', '/link/preview', req),
    unlink: (path: string) => request<LinkResult>('DELETE', '/link/unlink', { path }),
    undo: () => request<{ result: LinkResult; entry: HistoryEntry }>('POST', '/link/undo'),

    // History
    getHistory: (limit = 50) => request<HistoryEntry[]>('GET', `/history?limit=${limit}`),

    // Settings
    getSettings: () => request<Settings>('GET', '/settings'),
    updateSettings: (settings: Settings) => request<{ ok: boolean }>('PUT', '/settings', settings),
    changePassword: (current: string, newPass: string) => request<{ ok: boolean }>('POST', '/settings/password', { current, new: newPass }),

    // qBittorrent
    getQbitTorrents: (category?: string) => request<TorrentStatus[]>('GET', `/qbit/torrents${category ? `?category=${encodeURIComponent(category)}` : ''}`),
    addQbitTorrent: (magnet: string, category?: string) => request<{ ok: boolean }>('POST', '/qbit/add', { magnet, category }),
    deleteQbitTorrent: (hash: string, deleteFiles = false) => request<{ ok: boolean }>('DELETE', '/qbit/delete', { hash, deleteFiles }),
    testQbit: () => request<{ ok: boolean }>('GET', '/qbit/test'),

    // Nyaa
    searchNyaa: (q: string, filter?: string) => request<NyaaResult[]>('GET', `/nyaa/search?q=${encodeURIComponent(q)}${filter ? `&filter=${filter}` : ''}`),

    // Shoko
    shokoScan: () => request<{ ok: boolean }>('POST', '/shoko/scan'),
    testShoko: () => request<{ ok: boolean }>('GET', '/shoko/test'),
  }
}
