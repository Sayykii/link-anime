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
	// Video file extensions — strip before any other processing
	reVideoExt = regexp.MustCompile(`(?i)\.(mkv|mp4|avi|m4v|ts|webm)$`)

	// Leading [Group] tags
	reGroupTag = regexp.MustCompile(`^\[([^\]]*)\]\s*`)

	// SxxExx pattern — when present, everything from this token onward is episode info
	// We capture season and truncate the name to the prefix before this token.
	// Handles: S01E05, S02E05, S01E01-E24, S01E01-12
	reSxxExx = regexp.MustCompile(`(?i)\bS(\d+)\s*E\d+`)

	// Season-only patterns (no episode number)
	reSCode     = regexp.MustCompile(`(?i)\bS(\d+)(?:\b|$)`)
	reNthSeason = regexp.MustCompile(`(?i)\b([2-9])(nd|rd|th)\s+Season`)
	reSeasonN   = regexp.MustCompile(`(?i)\bSeason\s*(\d+)`)
	rePartNum   = regexp.MustCompile(`(?i)\bPart\s+(\d+)`)
	rePartRoman = regexp.MustCompile(`(?i)\bPart\s+(II|III|IV|V|VI|VII|VIII|IX)\b`)
	reCourNum   = regexp.MustCompile(`(?i)\bCour\s+(\d+)`)

	// Year detection: 4-digit year not followed by 'p' (to avoid 1080p/2160p)
	reYear = regexp.MustCompile(`\b((?:19|20)\d{2})(?:[^p\d]|$)`)

	// Release revision tags: v2, v3, etc.
	reRevision = regexp.MustCompile(`(?i)\s+v\d+\b`)

	// Stripping patterns
	reStripNthSeason = regexp.MustCompile(`(?i)\s+\d+(nd|rd|th)\s+Season`)
	reStripSCode     = regexp.MustCompile(`(?i)\s*-?\s*\bS\d+(?:\s*[Ee]\d+(?:-[Ee]?\d+)?)?\b`)
	reStripSeasonN   = regexp.MustCompile(`(?i)\s+-?\s*Season\s*\d+`)
	reStripPartNum   = regexp.MustCompile(`(?i)\s+-?\s*Part\s+(?:II|III|IV|V|VI|VII|VIII|IX|\d+)`)
	reStripCour      = regexp.MustCompile(`(?i)\s+-?\s*Cour\s+\d+`)

	// Episode ranges: (01-24), E01-E24, EP01-EP11, " - 00-23"
	reEpRange1 = regexp.MustCompile(`(?i)\s*\(?[Ee][Pp]?\d+(?:-[Ee]?[Pp]?\d+)?\)?`)
	reEpRange2 = regexp.MustCompile(`\s*\(\d+-\d+\)`)
	reEpRange3 = regexp.MustCompile(`\s+-\s*\d+-\d+`)

	// Trailing bracketed/parenthesized tags: [1080p], (BD FLAC), etc.
	// But NOT (YYYY) year tags — those are preserved separately.
	reBracketTrail = regexp.MustCompile(`\s*[\[\(][^\]\)]*[\]\)]\s*$`)

	// Technical keywords that always trail the show name
	reTechKeywords = regexp.MustCompile(`(?i)\s+(1080p|720p|480p|2160p|4[Kk]|[Bb][Dd]|[Bb]lu-?[Rr]ay|[Ww][Ee][Bb]-?[Dd][Ll]|[Ww][Ee][Bb]-?[Rr][Ii][Pp]|[Hh][Ee][Vv][Cc]|x26[45]|[Ff][Ll][Aa][Cc]|[Aa][Aa][Cc]|[Dd][Uu][Aa][Ll]|[Mm][Uu][Ll][Tt][Ii]|[Bb]atch|[Cc]omplete|[Rr][Ee][Mm][Uu][Xx]|[Dd][Dd][Pp]?\d|[Hh]\.?26[45]|[Aa][Vv]1|[Mm][Ss][Uu][Bb][Ss]?|NF|AMZN|CR)(\s.*|$)`)

	// Trailing hyphens/dashes
	reTrailDash = regexp.MustCompile(`\s*-\s*$`)

	// Leading hyphens/dashes
	reLeadDash = regexp.MustCompile(`^\s*-\s*`)

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
	var detectedYear string

	// === Phase 1: Pre-normalization cleanup ===

	// Strip video file extension FIRST (before dot conversion)
	name = reVideoExt.ReplaceAllString(name, "")

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

	// === Phase 2: Extract year before stripping ===
	// Detect year (19xx/20xx) not followed by 'p' (avoids 1080p/2160p)
	if m := reYear.FindStringSubmatch(name); m != nil {
		yr, _ := strconv.Atoi(m[1])
		if yr >= 1970 && yr <= 2030 {
			detectedYear = m[1]
		}
	}

	// === Phase 3: Season extraction with SxxExx truncation ===
	// If SxxExx is present, extract season and TRUNCATE name to prefix before it.
	// This drops episode titles like "Logistics in the Northern Plateau"
	if loc := reSxxExx.FindStringIndex(name); loc != nil {
		m := reSxxExx.FindStringSubmatch(name)
		if m != nil {
			season = parseSeasonNum(m[1])
		}
		// Keep only the part before SxxExx
		name = name[:loc[0]]
	} else if m := reSCode.FindStringSubmatch(name); m != nil {
		// S-code without episode: S01, S2, etc.
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

	// === Phase 4: Strip season indicators, revisions, episode ranges ===

	// Strip revision tags (v2, v3, etc.)
	name = reRevision.ReplaceAllString(name, "")

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

	// === Phase 5: Strip trailing noise ===

	// Strip trailing bracketed/parenthesized tags (repeatedly)
	for reBracketTrail.MatchString(name) {
		name = reBracketTrail.ReplaceAllString(name, "")
	}

	// Strip from first unbracketed technical keyword onward
	name = reTechKeywords.ReplaceAllString(name, "")

	// Strip trailing hyphens
	name = reTrailDash.ReplaceAllString(name, "")

	// Strip leading hyphens (can happen after group tag + season removal)
	name = reLeadDash.ReplaceAllString(name, "")

	// Strip the detected year from the name (we'll re-append it formatted)
	if detectedYear != "" {
		// Remove year in parens: (2025)
		name = strings.ReplaceAll(name, "("+detectedYear+")", "")
		// Remove bare year as standalone token
		reYearBare := regexp.MustCompile(`\b` + detectedYear + `\b`)
		name = reYearBare.ReplaceAllString(name, "")
	}

	// Collapse multiple spaces and trim
	name = reMultiSpace.ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)

	// Strip trailing hyphens again after cleanup
	name = reTrailDash.ReplaceAllString(name, "")
	name = strings.TrimSpace(name)

	// === Phase 6: Re-append year if detected ===
	if detectedYear != "" && name != "" {
		name = name + " (" + detectedYear + ")"
	}

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
