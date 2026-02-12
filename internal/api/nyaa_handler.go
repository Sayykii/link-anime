package api

import (
	"net/http"

	"link-anime/internal/nyaa"
)

func (s *Server) handleNyaaSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		jsonError(w, "q parameter required", http.StatusBadRequest)
		return
	}

	filter := r.URL.Query().Get("filter")

	results, err := nyaa.Search(query, filter)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if results == nil {
		jsonOK(w, []interface{}{})
		return
	}

	jsonOK(w, results)
}

func (s *Server) handleShokoScan(w http.ResponseWriter, r *http.Request) {
	if s.Shoko == nil || !s.Shoko.IsConfigured() {
		jsonError(w, "Shoko not configured", http.StatusBadRequest)
		return
	}

	if err := s.Shoko.ScanAllImportFolders(); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]bool{"ok": true})
}

func (s *Server) handleShokoTest(w http.ResponseWriter, r *http.Request) {
	if s.Shoko == nil || !s.Shoko.IsConfigured() {
		jsonError(w, "Shoko not configured", http.StatusBadRequest)
		return
	}

	if err := s.Shoko.TestConnection(); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]bool{"ok": true})
}
