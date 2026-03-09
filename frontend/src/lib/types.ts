// API types matching Go backend models

export interface Show {
  name: string
  path: string
  seasons: Season[]
  episodes: number
}

export interface Season {
  number: number
  path: string
  episodes: number
}

export interface Movie {
  name: string
  path: string
  files: number
}

export interface DownloadItem {
  name: string
  path: string
  isDir: boolean
  videoCount: number
  size: number
}

export interface LinkRequest {
  source: string
  type: 'series' | 'movie'
  name: string
  season: number
  dryRun: boolean
}

export interface LinkResult {
  linked: number
  skipped: number
  failed: number
  size: number
  destDir: string
  files: string[]
}

export interface HistoryEntry {
  id: number
  timestamp: string
  mediaType: string
  showName: string
  season?: number
  fileCount: number
  totalSize: number
  destPath: string
  source: string
}

export interface ParseResult {
  name: string
  season: number | null
}

export interface LibraryStats {
  shows: number
  seasons: number
  episodes: number
  movies: number
  size: number
}

export interface Settings {
  qbitUrl: string
  qbitUser: string
  qbitPass: string
  qbitCategory: string
  shokoUrl: string
  shokoApiKey: string
  notifyUrl: string
  downloadDir: string
  mediaDir: string
  moviesDir: string
}

export interface TorrentStatus {
  name: string
  hash: string
  state: string
  progress: number
  dlSpeed: number
  ulSpeed: number
  size: number
  eta: number
  ratio: number
}

export interface NyaaResult {
  title: string
  magnet: string
  size: string
  seeders: number
  leechers: number
}

export interface WSMessage {
  type: string
  data?: unknown
}

export interface LinkProgress {
  file: string
  status: 'linked' | 'skipped' | 'failed'
  current: number
  total: number
}

export interface RSSRule {
  id: number
  name: string
  query: string
  showName: string
  season: number
  mediaType: 'series' | 'movie'
  minSeeders: number
  resolution: string
  enabled: boolean
  lastCheck?: string
  createdAt: string
  matchCount: number
}

export interface RSSMatch {
  id: number
  ruleId: number
  title: string
  hash: string
  matched: string
  status: 'downloaded' | 'linked' | 'failed' | 'pending'
  ruleName: string
}

export interface FileSafetyInfo {
  path: string
  nlink: number
  safe: boolean
}

export interface UnlinkPreview {
  safeFiles: FileSafetyInfo[] | null
  unsafeFiles: FileSafetyInfo[] | null
  totalFiles: number
}

export interface TorrentProgress {
  torrents: TorrentStatus[]
  completed?: TorrentStatus[]
}
