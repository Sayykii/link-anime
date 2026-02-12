package api

import (
	"net/http"
	"strconv"

	"link-anime/internal/linker"
	"link-anime/internal/models"
)

func (s *Server) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	entries, err := linker.GetHistory(limit)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entries == nil {
		entries = []models.HistoryEntry{}
	}

	jsonOK(w, entries)
}
