package parser

import (
	"testing"
)

func intPtr(n int) *int {
	return &n
}

func TestParseReleaseName(t *testing.T) {
	tests := []struct {
		input          string
		expectedName   string
		expectedSeason *int
	}{
		// Basic group tag stripping
		{
			"[SubsPlease] Frieren S01 (1080p) [HEVC]",
			"Frieren",
			intPtr(1),
		},
		// Multiple group tags
		{
			"[Group1] [Group2] My Show S02 [1080p]",
			"My Show",
			intPtr(2),
		},
		// Dot-style naming
		{
			"Show.Name.S01.1080p.BluRay.FLAC",
			"Show Name",
			intPtr(1),
		},
		// Underscore naming
		{
			"[Group] Show_Name_S03 [720p]",
			"Show Name",
			intPtr(3),
		},
		// "2nd Season" pattern
		{
			"[SubsPlease] Mushoku Tensei 2nd Season [BD 1080p FLAC]",
			"Mushoku Tensei",
			intPtr(2),
		},
		// "3rd Season"
		{
			"[Group] Some Anime 3rd Season [1080p]",
			"Some Anime",
			intPtr(3),
		},
		// "Season 2"
		{
			"[sam] Dr. STONE Season 2 [BD 1080p FLAC]",
			"Dr. STONE",
			intPtr(2),
		},
		// "Season 01" with leading zero
		{
			"[Group] Show Name Season 01 [1080p]",
			"Show Name",
			intPtr(1),
		},
		// "Part 2"
		{
			"[Group] Attack on Titan Part 2 [1080p]",
			"Attack on Titan",
			intPtr(2),
		},
		// "Part III" roman numeral
		{
			"[Group] JoJo Part III [BD 1080p]",
			"JoJo",
			intPtr(3),
		},
		// "Part IV"
		{
			"[Group] JoJo Part IV [BD]",
			"JoJo",
			intPtr(4),
		},
		// "Cour 2"
		{
			"[Group] 86 Eighty-Six Cour 2 [1080p]",
			"86 Eighty-Six",
			intPtr(2),
		},
		// Episode range stripping
		{
			"[SubsPlease] Dr. STONE (01-24) (1080p)",
			"Dr. STONE",
			nil,
		},
		// S01E01-E24
		{
			"[Group] Show S01E01-E24 [1080p]",
			"Show",
			intPtr(1),
		},
		// Technical keyword stripping
		{
			"[Group] Vinland Saga 1080p x265 FLAC Batch",
			"Vinland Saga",
			nil,
		},
		// Complex real-world example
		{
			"[Erai-raws] Sousou no Frieren S01 (01-28) [1080p][Multiple Subtitle] [Batch]",
			"Sousou no Frieren",
			intPtr(1),
		},
		// No season info
		{
			"[SubsPlease] Akira (1988) [1080p]",
			"Akira (1988)",
			nil,
		},
		// Dot-style with season
		{
			"Bocchi.the.Rock.S01.1080p.WEB-DL.AAC",
			"Bocchi the Rock",
			intPtr(1),
		},
		// Season with dash separator
		{
			"[Group] Show Name - S2 [1080p]",
			"Show Name",
			intPtr(2),
		},
		// EP01-EP11 style
		{
			"[Group] Show EP01-EP11 [1080p]",
			"Show",
			nil,
		},
		// Batch tag with complete
		{
			"[Group] Spy x Family Complete [1080p] [Batch]",
			"Spy x Family",
			nil,
		},

		// === Real-world test cases ===

		// Dot-style movie with year, NF, DDP, H.264, release group after dash
		{
			"100.Meters.2025.1080p.NF.WEB-DL.MULTi.DDP5.1.H.264-VARYG.mkv",
			"100 Meters (2025)",
			nil,
		},
		// Bracketed group, subtitle in name, S-code with revision, BD AV1, dual audio
		{
			"[Breeze] Dr. STONE - New World - S03 v3 [1080p BD AV1][Dual Audio]",
			"Dr. STONE - New World",
			intPtr(3),
		},
		// Bracketed group, SxxExx pattern (name truncated before SxxExx)
		{
			"[KawaSubs] Journal with Witch - S01E01 [WEB 1080p AVC EAC3] [Eng-Sub].mkv",
			"Journal with Witch",
			intPtr(1),
		},
		// Bracketed group, S-code in title without dash separator, BD FLAC x265
		{
			"[KWTR] Takopis Original Sin S01 (BD 1080p FLAC x265)",
			"Takopis Original Sin",
			intPtr(1),
		},
		// Dot-style SxxExx with episode title — name truncated before SxxExx
		{
			"Frieren.Beyond.Journeys.End.S02E05.Logistics.in.the.Northern.Plateau.1080p.NF.WEB-DL.JPN.AAC2.0.H.264.MSubs-ToonsHub.mkv",
			"Frieren Beyond Journeys End",
			intPtr(2),
		},
		// Dot-style movie name with year and AMZN source
		{
			"One.Piece.Film.Red.2022.1080p.AMZN.WEB-DL.DDP5.1.H.264-VARYG.mkv",
			"One Piece Film Red (2022)",
			nil,
		},
		// Simple bracketed batch with episode range in parens
		{
			"[SubsPlease] Oshi no Ko (01-11) (1080p) [Batch]",
			"Oshi no Ko",
			nil,
		},
		// Dot-style with CR (Crunchyroll) tag
		{
			"Dandadan.S01.1080p.CR.WEB-DL.AAC.x265-Group.mkv",
			"Dandadan",
			intPtr(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseReleaseName(tt.input)

			if result.Name != tt.expectedName {
				t.Errorf("Name: got %q, want %q", result.Name, tt.expectedName)
			}

			if tt.expectedSeason == nil && result.Season != nil {
				t.Errorf("Season: got %d, want nil", *result.Season)
			} else if tt.expectedSeason != nil && result.Season == nil {
				t.Errorf("Season: got nil, want %d", *tt.expectedSeason)
			} else if tt.expectedSeason != nil && result.Season != nil && *result.Season != *tt.expectedSeason {
				t.Errorf("Season: got %d, want %d", *result.Season, *tt.expectedSeason)
			}
		})
	}
}
