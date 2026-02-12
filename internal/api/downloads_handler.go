package api

import (
	"net/http"

	"link-anime/internal/parser"
	"link-anime/internal/scanner"
)

func (s *Server) handleGetDownloads(w http.ResponseWriter, r *http.Request) {
	downloadDir := s.getDownloadDir()
	items, err := scanner.ScanDownloads(downloadDir)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if items == nil {
		jsonOK(w, []interface{}{})
		return
	}

	jsonOK(w, items)
}

func (s *Server) handleParseRelease(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		jsonError(w, "name parameter required", http.StatusBadRequest)
		return
	}

	result := parser.ParseReleaseName(name)
	jsonOK(w, map[string]interface{}{
		"name":   result.Name,
		"season": result.Season,
	})
}
