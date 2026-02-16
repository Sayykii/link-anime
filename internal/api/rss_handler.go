package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"link-anime/internal/models"
	"link-anime/internal/rss"
)

// handleListRSSRules returns all RSS rules.
func (s *Server) handleListRSSRules(w http.ResponseWriter, r *http.Request) {
	rules, err := rss.ListRules()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rules == nil {
		jsonOK(w, []interface{}{})
		return
	}
	jsonOK(w, rules)
}

// handleGetRSSRule returns a single RSS rule.
func (s *Server) handleGetRSSRule(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	rule, err := rss.GetRule(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rule == nil {
		jsonError(w, "rule not found", http.StatusNotFound)
		return
	}

	jsonOK(w, rule)
}

// handleCreateRSSRule creates a new RSS rule.
func (s *Server) handleCreateRSSRule(w http.ResponseWriter, r *http.Request) {
	var rule models.RSSRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if rule.Name == "" || rule.Query == "" || rule.ShowName == "" {
		jsonError(w, "name, query, and showName are required", http.StatusBadRequest)
		return
	}

	if rule.MediaType == "" {
		rule.MediaType = "series"
	}
	if rule.Season == 0 && rule.MediaType == "series" {
		rule.Season = 1
	}

	if err := rss.CreateRule(&rule); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, rule)
}

// handleUpdateRSSRule updates an existing RSS rule.
func (s *Server) handleUpdateRSSRule(w http.ResponseWriter, r *http.Request) {
	var rule models.RSSRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if rule.ID == 0 {
		jsonError(w, "id is required", http.StatusBadRequest)
		return
	}

	if err := rss.UpdateRule(&rule); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, rule)
}

// handleDeleteRSSRule deletes an RSS rule.
func (s *Server) handleDeleteRSSRule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.ID == 0 {
		jsonError(w, "id is required", http.StatusBadRequest)
		return
	}

	if err := rss.DeleteRule(req.ID); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]bool{"ok": true})
}

// handleToggleRSSRule enables or disables an RSS rule.
func (s *Server) handleToggleRSSRule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID      int64 `json:"id"`
		Enabled bool  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.ID == 0 {
		jsonError(w, "id is required", http.StatusBadRequest)
		return
	}

	if err := rss.ToggleRule(req.ID, req.Enabled); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]bool{"ok": true})
}

// handleListRSSMatches returns matches, optionally filtered by rule ID.
func (s *Server) handleListRSSMatches(w http.ResponseWriter, r *http.Request) {
	ruleIDStr := r.URL.Query().Get("ruleId")
	limitStr := r.URL.Query().Get("limit")

	var ruleID int64
	if ruleIDStr != "" {
		var err error
		ruleID, err = strconv.ParseInt(ruleIDStr, 10, 64)
		if err != nil {
			jsonError(w, "invalid ruleId", http.StatusBadRequest)
			return
		}
	}

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	matches, err := rss.ListMatches(ruleID, limit)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if matches == nil {
		jsonOK(w, []interface{}{})
		return
	}

	jsonOK(w, matches)
}

// handleClearRSSMatches deletes all matches for a rule.
func (s *Server) handleClearRSSMatches(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RuleID int64 `json:"ruleId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.RuleID == 0 {
		jsonError(w, "ruleId is required", http.StatusBadRequest)
		return
	}

	if err := rss.ClearMatches(req.RuleID); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]bool{"ok": true})
}

// handleRSSPollNow triggers an immediate RSS poll.
func (s *Server) handleRSSPollNow(w http.ResponseWriter, r *http.Request) {
	if s.Poller == nil {
		jsonError(w, "RSS poller not initialized", http.StatusBadRequest)
		return
	}

	s.Poller.PollNow()
	jsonOK(w, map[string]bool{"ok": true})
}
