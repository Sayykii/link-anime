package shoko

// Paginated response wrapper from Shoko API.
type listResult[T any] struct {
	Total int `json:"Total"`
	List  []T `json:"List"`
}

// ShokoSeries represents a series from the Shoko API.
type ShokoSeries struct {
	Name   string        `json:"Name"`
	IDs    ShokoIDs      `json:"IDs"`
	Images ShokoImages   `json:"Images"`
	Sizes  ShokoSizes    `json:"Sizes"`
	Size   int           `json:"Size"` // total local file count
	AirsOn []string      `json:"AirsOn,omitempty"`
	Created string       `json:"Created"`
	Updated string       `json:"Updated"`

	// Included when ?includeDataFrom=AniDB
	AniDB *ShokoAniDB `json:"AniDB,omitempty"`
}

// ShokoIDs contains cross-reference IDs for a series.
type ShokoIDs struct {
	ID          int   `json:"ID"`
	AniDB       int   `json:"AniDB"`
	MAL         []int `json:"MAL,omitempty"`
	ParentGroup int   `json:"ParentGroup"`
}

// ShokoImages contains image collections for a series.
type ShokoImages struct {
	Posters   []ShokoImage `json:"Posters,omitempty"`
	Backdrops []ShokoImage `json:"Backdrops,omitempty"`
	Banners   []ShokoImage `json:"Banners,omitempty"`
}

// ShokoImage represents a single image reference.
type ShokoImage struct {
	ID       int    `json:"ID"`
	Type     string `json:"Type"`     // Poster, Backdrop, Banner, etc.
	Source   string `json:"Source"`   // AniDB, TMDB, Shoko
	Preferred bool  `json:"Preferred"`
	Width    int    `json:"Width,omitempty"`
	Height   int    `json:"Height,omitempty"`
}

// ShokoSizes contains episode count breakdowns by type.
type ShokoSizes struct {
	Local   ShokoEpisodeTypeCounts  `json:"Local"`
	Watched ShokoEpisodeTypeCounts  `json:"Watched"`
	Total   ShokoEpisodeTypeCounts  `json:"Total"`
	Missing ShokoMissingCounts      `json:"Missing"`
}

// ShokoEpisodeTypeCounts has counts per episode type.
type ShokoEpisodeTypeCounts struct {
	Unknown  int `json:"Unknown"`
	Episodes int `json:"Episodes"`
	Specials int `json:"Specials"`
	Credits  int `json:"Credits"`
	Trailers int `json:"Trailers"`
	Parodies int `json:"Parodies"`
	Others   int `json:"Others"`
}

// ShokoMissingCounts has missing counts (only episodes and specials).
type ShokoMissingCounts struct {
	Episodes int `json:"Episodes"`
	Specials int `json:"Specials"`
}

// ShokoAniDB contains AniDB-specific metadata.
type ShokoAniDB struct {
	ID          int    `json:"ID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
	Type        string `json:"Type"` // TVSeries, Movie, OVA, etc.
	EpisodeCount int   `json:"EpisodeCount"`
	AirDate     string `json:"AirDate,omitempty"`
	EndDate     string `json:"EndDate,omitempty"`
	Rating      *ShokoRating `json:"Rating,omitempty"`
}

// ShokoRating contains rating info.
type ShokoRating struct {
	Value    float64 `json:"Value"`
	MaxValue float64 `json:"MaxValue"`
	Votes    int     `json:"Votes"`
	Source   string  `json:"Source"`
}

// ShokoEpisode represents an episode from the Shoko API.
type ShokoEpisode struct {
	Name       string      `json:"Name"`
	IDs        ShokoEpIDs  `json:"IDs"`
	Images     ShokoImages `json:"Images"`
	Duration   string      `json:"Duration"` // TimeSpan format
	Watched    *string     `json:"Watched"`  // null = unwatched
	WatchCount int         `json:"WatchCount"`
	Size       int         `json:"Size"` // file count (0 = missing)
	Created    string      `json:"Created"`

	// Included when ?includeDataFrom=AniDB
	AniDB *ShokoAniDBEpisode `json:"AniDB,omitempty"`
}

// ShokoEpIDs contains episode cross-reference IDs.
type ShokoEpIDs struct {
	ID           int `json:"ID"`
	ParentSeries int `json:"ParentSeries"`
	AniDB        int `json:"AniDB"`
}

// ShokoAniDBEpisode contains AniDB episode metadata.
type ShokoAniDBEpisode struct {
	ID       int    `json:"ID"`
	Type     string `json:"Type"` // Episode, Special, etc.
	Number   int    `json:"EpisodeNumber"`
	Title    string `json:"Title"`
	AirDate  string `json:"AirDate,omitempty"`
	Summary  string `json:"Summary,omitempty"`
	Rating   *ShokoRating `json:"Rating,omitempty"`
}

// ShokoDashboardStats contains collection statistics.
type ShokoDashboardStats struct {
	FileCount                  int     `json:"FileCount"`
	FileSize                   int64   `json:"FileSize"`
	SeriesCount                int     `json:"SeriesCount"`
	GroupCount                 int     `json:"GroupCount"`
	FinishedSeries             int     `json:"FinishedSeries"`
	WatchedEpisodes            int     `json:"WatchedEpisodes"`
	WatchedHours               float64 `json:"WatchedHours"`
	MissingEpisodes            int     `json:"MissingEpisodes"`
	MissingEpisodesCollecting  int     `json:"MissingEpisodesCollecting"`
	UnrecognizedFiles          int     `json:"UnrecognizedFiles"`
}

// ShokoDashboardEpisode represents an episode in dashboard endpoints.
type ShokoDashboardEpisode struct {
	Title        string     `json:"Title"`
	Number       int        `json:"Number"`
	Type         string     `json:"Type"`
	AirDate      string     `json:"AirDate,omitempty"`
	SeriesTitle  string     `json:"SeriesTitle"`
	SeriesPoster *ShokoImage `json:"SeriesPoster,omitempty"`
	IDs          struct {
		ID          int `json:"ID"`
		Series      int `json:"Series"`
		ShokoSeries int `json:"ShokoSeries"`
	} `json:"IDs"`
}
