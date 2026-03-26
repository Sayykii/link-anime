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
  files: string[]
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
  linked: boolean
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
  filter: string    // "", "noremakes", "trusted"
  category: string  // nyaa category: "1_2", "1_0", etc.
  groups: string    // comma-separated release group allowlist
  autoLink: boolean
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
  torrentName: string
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

// --- Shoko types ---

export interface ShokoImage {
  ID: number
  Type: string
  Source: string
  Preferred: boolean
  Width?: number
  Height?: number
}

export interface ShokoEpisodeTypeCounts {
  Unknown: number
  Episodes: number
  Specials: number
  Credits: number
  Trailers: number
  Parodies: number
  Others: number
}

export interface ShokoSizes {
  Local: ShokoEpisodeTypeCounts
  Watched: ShokoEpisodeTypeCounts
  Total: ShokoEpisodeTypeCounts
  Missing: { Episodes: number; Specials: number }
}

export interface ShokoAniDB {
  ID: number
  Title: string
  Description: string
  Type: string
  EpisodeCount: number
  AirDate?: string
  EndDate?: string
  Rating?: { Value: number; MaxValue: number; Votes: number; Source: string }
}

export interface ShokoSeries {
  Name: string
  IDs: { ID: number; AniDB: number; MAL?: number[]; ParentGroup: number }
  Images: { Posters?: ShokoImage[]; Backdrops?: ShokoImage[]; Banners?: ShokoImage[] }
  Sizes: ShokoSizes
  Size: number
  AirsOn?: Record<string, boolean>
  Created: string
  Updated: string
  AniDB?: ShokoAniDB
}

export interface ShokoEpisode {
  Name: string
  IDs: { ID: number; ParentSeries: number; AniDB: number }
  Images: { Posters?: ShokoImage[]; Backdrops?: ShokoImage[] }
  Duration: string
  Watched: string | null
  WatchCount: number
  Size: number // 0 = missing
  Created: string
  AniDB?: {
    ID: number
    Type: string
    EpisodeNumber: number
    Title: string
    AirDate?: string
    Summary?: string
  }
}

export interface ShokoDashboardStats {
  FileCount: number
  FileSize: number
  SeriesCount: number
  GroupCount: number
  FinishedSeries: number
  WatchedEpisodes: number
  WatchedHours: number
  MissingEpisodes: number
  MissingEpisodesCollecting: number
  UnrecognizedFiles: number
}

export interface ShokoDashboardEpisode {
  Title: string
  Number: number
  Type: string
  AirDate?: string
  SeriesTitle: string
  SeriesPoster?: ShokoImage
  IDs: { ID: number; Series: number; ShokoSeries: number }
}

export interface ShokoDashboard {
  stats: ShokoDashboardStats | null
  recentlyAdded: ShokoDashboardEpisode[] | null
  continueWatching: ShokoDashboardEpisode[] | null
  calendar: ShokoDashboardEpisode[] | null
}
