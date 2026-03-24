package rss

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"link-anime/internal/database"
	"link-anime/internal/models"
	"link-anime/internal/nyaa"
	"link-anime/internal/qbit"
	"link-anime/internal/scanner"
	"link-anime/internal/ws"
)

// QbitGetter is a function that returns the current qBittorrent client.
// This allows the poller to pick up client changes from reinitClients.
type QbitGetter func() *qbit.Client

// Poller periodically checks Nyaa RSS feeds for new matches.
type Poller struct {
	hub      *ws.Hub
	getQbit  QbitGetter
	interval time.Duration
	stopCh   chan struct{}
	mu       sync.Mutex
	running  bool
	MediaDir func() string
	MoviesDir func() string
}

// NewPoller creates a new RSS poller.
func NewPoller(hub *ws.Hub, getQbit QbitGetter, interval time.Duration) *Poller {
	if interval < 5*time.Minute {
		interval = 15 * time.Minute
	}
	return &Poller{
		hub:      hub,
		getQbit:  getQbit,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the polling loop in a goroutine.
func (p *Poller) Start() {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return
	}
	p.running = true
	p.mu.Unlock()

	log.Printf("RSS poller started (interval: %s)", p.interval)

	go func() {
		// Run immediately on start
		p.poll()

		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.poll()
			case <-p.stopCh:
				log.Println("RSS poller stopped")
				return
			}
		}
	}()
}

// Stop stops the polling loop.
func (p *Poller) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.running {
		close(p.stopCh)
		p.running = false
	}
}

// PollNow triggers an immediate poll cycle (for manual refresh).
func (p *Poller) PollNow() {
	go p.poll()
}

// poll runs one cycle: fetch each enabled rule, check for new matches.
func (p *Poller) poll() {
	rules, err := ListRules()
	if err != nil {
		log.Printf("RSS poll: failed to list rules: %v", err)
		return
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		p.checkRule(rule)
	}
}

// FilterResults applies a rule's filters (seeders, resolution, groups) to search results.
// Exported so the test-query endpoint can reuse the same logic.
func FilterResults(results []models.NyaaResult, rule models.RSSRule) []models.NyaaResult {
	var filtered []models.NyaaResult

	// Parse groups allowlist once
	var groups []string
	if rule.Groups != "" {
		for _, g := range strings.Split(rule.Groups, ",") {
			g = strings.TrimSpace(g)
			if g != "" {
				groups = append(groups, strings.ToLower(g))
			}
		}
	}

	for _, result := range results {
		// Min seeders filter
		if rule.MinSeeders > 0 && result.Seeders < rule.MinSeeders {
			continue
		}

		// Resolution filter
		if rule.Resolution != "" && !strings.Contains(strings.ToLower(result.Title), strings.ToLower(rule.Resolution)) {
			continue
		}

		// Release group filter
		if len(groups) > 0 {
			titleLower := strings.ToLower(result.Title)
			matched := false
			for _, g := range groups {
				if strings.Contains(titleLower, g) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		filtered = append(filtered, result)
	}

	return filtered
}

// checkRule fetches the Nyaa RSS feed for a rule and processes matches.
func (p *Poller) checkRule(rule models.RSSRule) {
	// Use rule's filter, default to "noremakes" if empty
	filter := rule.Filter
	if filter == "" {
		filter = "noremakes"
	}

	results, err := nyaa.Search(rule.Query, filter, rule.Category)
	if err != nil {
		log.Printf("RSS poll [%s]: search failed: %v", rule.Name, err)
		return
	}

	// Apply filters (seeders, resolution, groups)
	results = FilterResults(results, rule)

	for _, result := range results {
		// Generate a hash for deduplication (use title hash since Nyaa RSS doesn't give info hashes directly)
		hash := hashTitle(result.Title)

		// Check if we already matched this
		if isAlreadyMatched(rule.ID, hash) {
			continue
		}

		// Check if the episode already exists in the library
		epNum := parseEpisodeNum(result.Title)
		if epNum > 0 && p.episodeExistsInLibrary(rule, epNum) {
			log.Printf("RSS skip [%s]: episode %d already in library, skipping %s", rule.Name, epNum, result.Title)
			// Record as skipped so we don't check again
			if err := InsertMatch(rule.ID, result.Title, hash, "skipped", result.Title); err != nil {
				log.Printf("RSS poll [%s]: failed to record skipped match: %v", rule.Name, err)
			}
			continue
		}

		// New match found
		log.Printf("RSS match [%s]: %s", rule.Name, result.Title)

		// Try to add to qBittorrent if configured
		status := "downloaded"
		qbitClient := p.getQbit()
		if qbitClient != nil && qbitClient.IsConfigured() {
			if err := qbitClient.AddMagnet(result.Magnet, "", ""); err != nil {
				log.Printf("RSS poll [%s]: failed to add torrent: %v", rule.Name, err)
				status = "failed"
			}
		} else {
			log.Printf("RSS poll [%s]: qBittorrent not configured, recording match only", rule.Name)
			status = "pending"
		}

		// Record the match (store title as torrent_name for auto-link matching)
		if err := InsertMatch(rule.ID, result.Title, hash, status, result.Title); err != nil {
			log.Printf("RSS poll [%s]: failed to record match: %v", rule.Name, err)
		}

		// Broadcast via WebSocket
		p.hub.Broadcast(models.WSMessage{
			Type: "rss_match",
			Data: map[string]interface{}{
				"ruleName": rule.Name,
				"title":    result.Title,
				"status":   status,
			},
		})
	}

	// Update last_check timestamp
	updateLastCheck(rule.ID)
}

// reEpisodeNum extracts episode numbers from torrent titles like S01E05, E05, - 05, etc.
var reEpisodeNum = regexp.MustCompile(`(?i)(?:S\d+\s*)?E(\d+)|(?:^|[\s\-])\s*(\d{2,3})\s*(?:[\s\-\[]|$)`)

// parseEpisodeNum extracts the episode number from a torrent title.
// Returns -1 if no episode number found (e.g., batch releases).
func parseEpisodeNum(title string) int {
	m := reEpisodeNum.FindStringSubmatch(title)
	if m == nil {
		return -1
	}
	numStr := m[1]
	if numStr == "" {
		numStr = m[2]
	}
	// Avoid matching years like 2024, 1080, etc.
	if len(numStr) == 4 {
		return -1
	}
	n := 0
	for _, c := range numStr {
		n = n*10 + int(c-'0')
	}
	if n == 0 || n > 999 {
		return -1
	}
	return n
}

// episodeExistsInLibrary checks if an episode file matching the given episode number
// already exists in the library destination for the rule's show/season.
func (p *Poller) episodeExistsInLibrary(rule models.RSSRule, episodeNum int) bool {
	if p.MediaDir == nil && p.MoviesDir == nil {
		return false
	}

	var destDir string
	if rule.MediaType == "movie" {
		if p.MoviesDir == nil {
			return false
		}
		destDir = filepath.Join(p.MoviesDir(), rule.ShowName)
	} else {
		if p.MediaDir == nil {
			return false
		}
		destDir = filepath.Join(p.MediaDir(), rule.ShowName, fmt.Sprintf("Season %d", rule.Season))
	}

	entries, err := os.ReadDir(destDir)
	if err != nil {
		return false // dir doesn't exist = episode doesn't exist
	}

	// Build pattern to match episode number in filenames
	// Matches: E05, E5, S01E05, - 05, etc.
	epPattern := regexp.MustCompile(fmt.Sprintf(`(?i)(?:E0*%d|(?:^|[\s\-])0*%d(?:[\s\-\.\[]|$))`, episodeNum, episodeNum))

	for _, e := range entries {
		if !e.IsDir() && scanner.IsVideo(e.Name()) {
			if epPattern.MatchString(e.Name()) {
				return true
			}
		}
	}

	return false
}

// --- Database helpers ---

// ListRules returns all RSS rules.
func ListRules() ([]models.RSSRule, error) {
	rows, err := database.DB.Query(`
		SELECT r.id, r.name, r.query, r.show_name, r.season, r.media_type,
		       r.min_seeders, r.resolution, r.filter, r.category, r.groups,
		       r.auto_link, r.enabled, r.last_check, r.created_at,
		       (SELECT COUNT(*) FROM rss_matches WHERE rule_id = r.id) as match_count
		FROM rss_rules r
		ORDER BY r.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list rules: %w", err)
	}
	defer rows.Close()

	var rules []models.RSSRule
	for rows.Next() {
		var r models.RSSRule
		var lastCheck sql.NullTime
		if err := rows.Scan(&r.ID, &r.Name, &r.Query, &r.ShowName, &r.Season,
			&r.MediaType, &r.MinSeeders, &r.Resolution, &r.Filter, &r.Category,
			&r.Groups, &r.AutoLink, &r.Enabled,
			&lastCheck, &r.CreatedAt, &r.MatchCount); err != nil {
			return nil, fmt.Errorf("scan rule: %w", err)
		}
		if lastCheck.Valid {
			r.LastCheck = &lastCheck.Time
		}
		rules = append(rules, r)
	}

	return rules, nil
}

// GetRule returns a single RSS rule by ID.
func GetRule(id int64) (*models.RSSRule, error) {
	var r models.RSSRule
	var lastCheck sql.NullTime
	err := database.DB.QueryRow(`
		SELECT r.id, r.name, r.query, r.show_name, r.season, r.media_type,
		       r.min_seeders, r.resolution, r.filter, r.category, r.groups,
		       r.auto_link, r.enabled, r.last_check, r.created_at,
		       (SELECT COUNT(*) FROM rss_matches WHERE rule_id = r.id) as match_count
		FROM rss_rules r WHERE r.id = ?
	`, id).Scan(&r.ID, &r.Name, &r.Query, &r.ShowName, &r.Season,
		&r.MediaType, &r.MinSeeders, &r.Resolution, &r.Filter, &r.Category,
		&r.Groups, &r.AutoLink, &r.Enabled,
		&lastCheck, &r.CreatedAt, &r.MatchCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get rule: %w", err)
	}
	if lastCheck.Valid {
		r.LastCheck = &lastCheck.Time
	}
	return &r, nil
}

// CreateRule inserts a new RSS rule.
func CreateRule(r *models.RSSRule) error {
	result, err := database.DB.Exec(`
		INSERT INTO rss_rules (name, query, show_name, season, media_type, min_seeders, resolution, filter, category, groups, auto_link, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, r.Name, r.Query, r.ShowName, r.Season, r.MediaType, r.MinSeeders, r.Resolution, r.Filter, r.Category, r.Groups, r.AutoLink, r.Enabled)
	if err != nil {
		return fmt.Errorf("create rule: %w", err)
	}
	r.ID, _ = result.LastInsertId()
	return nil
}

// UpdateRule updates an existing RSS rule.
func UpdateRule(r *models.RSSRule) error {
	_, err := database.DB.Exec(`
		UPDATE rss_rules SET name = ?, query = ?, show_name = ?, season = ?,
		       media_type = ?, min_seeders = ?, resolution = ?, filter = ?,
		       category = ?, groups = ?, auto_link = ?, enabled = ?
		WHERE id = ?
	`, r.Name, r.Query, r.ShowName, r.Season, r.MediaType, r.MinSeeders, r.Resolution,
		r.Filter, r.Category, r.Groups, r.AutoLink, r.Enabled, r.ID)
	if err != nil {
		return fmt.Errorf("update rule: %w", err)
	}
	return nil
}

// DeleteRule deletes an RSS rule and its matches.
func DeleteRule(id int64) error {
	_, err := database.DB.Exec("DELETE FROM rss_rules WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete rule: %w", err)
	}
	return nil
}

// ToggleRule enables or disables an RSS rule.
func ToggleRule(id int64, enabled bool) error {
	_, err := database.DB.Exec("UPDATE rss_rules SET enabled = ? WHERE id = ?", enabled, id)
	if err != nil {
		return fmt.Errorf("toggle rule: %w", err)
	}
	return nil
}

// ListMatches returns matches, optionally filtered by rule ID.
func ListMatches(ruleID int64, limit int) ([]models.RSSMatch, error) {
	query := `
		SELECT m.id, m.rule_id, m.title, m.hash, m.torrent_name, m.matched, m.status, r.name
		FROM rss_matches m
		JOIN rss_rules r ON r.id = m.rule_id
	`
	args := []interface{}{}

	if ruleID > 0 {
		query += " WHERE m.rule_id = ?"
		args = append(args, ruleID)
	}

	query += " ORDER BY m.matched DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list matches: %w", err)
	}
	defer rows.Close()

	var matches []models.RSSMatch
	for rows.Next() {
		var m models.RSSMatch
		if err := rows.Scan(&m.ID, &m.RuleID, &m.Title, &m.Hash, &m.TorrentName, &m.Matched, &m.Status, &m.RuleName); err != nil {
			return nil, fmt.Errorf("scan match: %w", err)
		}
		matches = append(matches, m)
	}

	return matches, nil
}

// InsertMatch records a new RSS match.
func InsertMatch(ruleID int64, title, hash, status, torrentName string) error {
	_, err := database.DB.Exec(`
		INSERT OR IGNORE INTO rss_matches (rule_id, title, hash, status, torrent_name) VALUES (?, ?, ?, ?, ?)
	`, ruleID, title, hash, status, torrentName)
	return err
}

// ClearMatches deletes all matches for a rule.
func ClearMatches(ruleID int64) error {
	_, err := database.DB.Exec("DELETE FROM rss_matches WHERE rule_id = ?", ruleID)
	return err
}

// FindMatchByTorrentName looks up a "downloaded" RSS match by torrent name.
// Returns the match and its associated rule, or nil if not found.
func FindMatchByTorrentName(name string) (*models.RSSMatch, *models.RSSRule) {
	var m models.RSSMatch
	var r models.RSSRule
	var lastCheck sql.NullTime

	err := database.DB.QueryRow(`
		SELECT m.id, m.rule_id, m.title, m.hash, m.torrent_name, m.matched, m.status,
		       r.id, r.name, r.show_name, r.season, r.media_type, r.auto_link,
		       r.last_check
		FROM rss_matches m
		JOIN rss_rules r ON r.id = m.rule_id
		WHERE m.torrent_name = ? AND m.status = 'downloaded'
		LIMIT 1
	`, name).Scan(
		&m.ID, &m.RuleID, &m.Title, &m.Hash, &m.TorrentName, &m.Matched, &m.Status,
		&r.ID, &r.Name, &r.ShowName, &r.Season, &r.MediaType, &r.AutoLink,
		&lastCheck,
	)
	if err != nil {
		return nil, nil
	}
	if lastCheck.Valid {
		r.LastCheck = &lastCheck.Time
	}
	return &m, &r
}

// UpdateMatchStatus updates the status of a match.
func UpdateMatchStatus(id int64, status string) error {
	_, err := database.DB.Exec("UPDATE rss_matches SET status = ? WHERE id = ?", status, id)
	return err
}

// --- Helpers ---

func isAlreadyMatched(ruleID int64, hash string) bool {
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM rss_matches WHERE rule_id = ? AND hash = ?", ruleID, hash).Scan(&count)
	return count > 0
}

func updateLastCheck(ruleID int64) {
	database.DB.Exec("UPDATE rss_rules SET last_check = CURRENT_TIMESTAMP WHERE id = ?", ruleID)
}

func hashTitle(title string) string {
	h := sha256.Sum256([]byte(title))
	return hex.EncodeToString(h[:16]) // 32-char hex string
}
