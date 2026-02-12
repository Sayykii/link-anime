package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// Result holds the parsed release name and optional season.
type Result struct {
	Name   string
	Season *int
}

var (
	// Leading [Group] tags
	reGroupTag = regexp.MustCompile(`^\[([^\]]*)\]\s*`)

	// Season patterns (order matters — "nth Season" must be checked before "Season N")
	reSCode       = regexp.MustCompile(`(?i)\bS(\d+)(?:[Ee]\d+(?:-[Ee]?\d+)?)?(?:\b|$)`)
	reNthSeason   = regexp.MustCompile(`(?i)\b([2-9])(nd|rd|th)\s+Season`)
	reSeasonN     = regexp.MustCompile(`(?i)\bSeason\s*(\d+)`)
	rePartNum     = regexp.MustCompile(`(?i)\bPart\s+(\d+)`)
	rePartRoman   = regexp.MustCompile(`(?i)\bPart\s+(II|III|IV|V|VI|VII|VIII|IX)\b`)
	reCourNum     = regexp.MustCompile(`(?i)\bCour\s+(\d+)`)

	// Stripping patterns
	reStripNthSeason  = regexp.MustCompile(`(?i)\s+\d+(nd|rd|th)\s+Season`)
	reStripSCode      = regexp.MustCompile(`(?i)\s+-?\s*S\d+(?:[Ee]\d+(?:-[Ee]?\d+)?)?`)
	reStripSeasonN    = regexp.MustCompile(`(?i)\s+-?\s*Season\s*\d+`)
	reStripPartNum    = regexp.MustCompile(`(?i)\s+-?\s*Part\s+(?:II|III|IV|V|VI|VII|VIII|IX|\d+)`)
	reStripCour       = regexp.MustCompile(`(?i)\s+-?\s*Cour\s+\d+`)

	// Episode ranges: (01-24), E01-E24, EP01-EP11, " - 00-23"
	reEpRange1 = regexp.MustCompile(`(?i)\s*\(?[Ee][Pp]?\d+(?:-[Ee]?[Pp]?\d+)?\)?`)
	reEpRange2 = regexp.MustCompile(`\s*\(\d+-\d+\)`)
	reEpRange3 = regexp.MustCompile(`\s+-\s*\d+-\d+`)

	// Trailing bracketed/parenthesized tags: [1080p], (BD FLAC), etc.
	reBracketTrail = regexp.MustCompile(`\s*[\[\(][^\]\)]*[\]\)]\s*$`)

	// Technical keywords that always trail the show name
	reTechKeywords = regexp.MustCompile(`(?i)\s+(1080p|720p|480p|2160p|4[Kk]|[Bb][Dd]|[Bb]lu-?[Rr]ay|[Ww][Ee][Bb]-?[Dd][Ll]|[Ww][Ee][Bb]-?[Rr][Ii][Pp]|[Hh][Ee][Vv][Cc]|x26[45]|[Ff][Ll][Aa][Cc]|[Aa][Aa][Cc]|[Dd][Uu][Aa][Ll]|[Mm][Uu][Ll][Tt][Ii]|[Bb]atch|[Cc]omplete|[Rr][Ee][Mm][Uu][Xx])(\s.*|$)`)

	// Trailing hyphens/dashes
	reTrailDash = regexp.MustCompile(`\s*-\s*$`)

	// Multiple spaces
	reMultiSpace = regexp.MustCompile(`\s+`)
)

var romanMap = map[string]int{
	"II": 2, "III": 3, "IV": 4, "V": 5,
	"VI": 6, "VII": 7, "VIII": 8, "IX": 9,
}

// ParseReleaseName extracts a clean show/movie name and optional season
// from a typical anime release folder name.
func ParseReleaseName(input string) Result {
	name := input
	var season *int

	// Strip leading [Group] tags (one or more)
	for reGroupTag.MatchString(name) {
		name = reGroupTag.ReplaceAllString(name, "")
	}

	// Detect dot-style naming (more dots than spaces) and convert
	dotCount := strings.Count(name, ".")
	spaceCount := strings.Count(name, " ")
	if dotCount > spaceCount && dotCount >= 2 {
		name = strings.ReplaceAll(name, ".", " ")
	}

	// Replace underscores with spaces
	name = strings.ReplaceAll(name, "_", " ")

	// Extract season number (try each pattern in priority order)
	// "nth Season" must be before "Season N" to avoid partial matches
	if m := reSCode.FindStringSubmatch(name); m != nil {
		season = parseSeasonNum(m[1])
	} else if m := reNthSeason.FindStringSubmatch(name); m != nil {
		season = parseSeasonNum(m[1])
	} else if m := reSeasonN.FindStringSubmatch(name); m != nil {
		season = parseSeasonNum(m[1])
	} else if m := rePartNum.FindStringSubmatch(name); m != nil {
		season = parseSeasonNum(m[1])
	} else if m := rePartRoman.FindStringSubmatch(name); m != nil {
		if v, ok := romanMap[m[1]]; ok {
			season = &v
		}
	} else if m := reCourNum.FindStringSubmatch(name); m != nil {
		season = parseSeasonNum(m[1])
	}

	// Strip season indicators (order: nth Season, S-code, Season N, Part, Cour)
	name = reStripNthSeason.ReplaceAllString(name, "")
	name = reStripSCode.ReplaceAllString(name, "")
	name = reStripSeasonN.ReplaceAllString(name, "")
	name = reStripPartNum.ReplaceAllString(name, "")
	name = reStripCour.ReplaceAllString(name, "")

	// Strip episode ranges
	name = reEpRange1.ReplaceAllString(name, "")
	name = reEpRange2.ReplaceAllString(name, "")
	name = reEpRange3.ReplaceAllString(name, "")

	// Strip trailing bracketed/parenthesized tags (repeatedly)
	for reBracketTrail.MatchString(name) {
		name = reBracketTrail.ReplaceAllString(name, "")
	}

	// Strip from first unbracketed technical keyword onward
	name = reTechKeywords.ReplaceAllString(name, "")

	// Strip trailing hyphens
	name = reTrailDash.ReplaceAllString(name, "")

	// Collapse multiple spaces and trim
	name = reMultiSpace.ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)

	return Result{Name: name, Season: season}
}

func parseSeasonNum(s string) *int {
	// Strip leading zeros
	s = strings.TrimLeft(s, "0")
	if s == "" {
		s = "0"
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &n
}
