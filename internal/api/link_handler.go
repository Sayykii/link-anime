package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"link-anime/internal/linker"
	"link-anime/internal/models"
	"link-anime/internal/notify"
)

func (s *Server) handleLink(w http.ResponseWriter, r *http.Request) {
	var req models.LinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Source == "" || req.Name == "" || req.Type == "" {
		jsonError(w, "source, name, and type are required", http.StatusBadRequest)
		return
	}

	if req.Type != "series" && req.Type != "movie" {
		jsonError(w, "type must be 'series' or 'movie'", http.StatusBadRequest)
		return
	}

	downloadDir := s.getDownloadDir()
	mediaDir := s.getMediaDir()
	moviesDir := s.getMoviesDir()

	result, err := linker.Link(req, downloadDir, mediaDir, moviesDir, s.Hub)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send notification
	if s.Notifier != nil && result.Linked > 0 && !req.DryRun {
		title := "Linked: " + req.Name
		msg := ""
		if req.Type == "series" {
			msg = fmt.Sprintf("Linked to Season %d", req.Season)
		} else {
			msg = "Linked movie"
		}
		s.Notifier.Send(title, msg, []notify.Field{
			{Name: "Files", Value: fmt.Sprintf("%d", result.Linked)},
			{Name: "Size", Value: notify.FormatSize(result.Size)},
		}, "green")
	}

	// Trigger Shoko scan if configured
	if s.Shoko != nil && s.Shoko.IsConfigured() && result.Linked > 0 && !req.DryRun {
		go s.Shoko.ScanAllImportFolders()
	}

	jsonOK(w, result)
}

func (s *Server) handleLinkPreview(w http.ResponseWriter, r *http.Request) {
	var req models.LinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	req.DryRun = true

	downloadDir := s.getDownloadDir()
	mediaDir := s.getMediaDir()
	moviesDir := s.getMoviesDir()

	result, err := linker.Link(req, downloadDir, mediaDir, moviesDir, nil)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, result)
}

func (s *Server) handleUnlinkPreview(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		jsonError(w, "path query parameter is required", http.StatusBadRequest)
		return
	}

	preview, err := linker.UnlinkPreview(path)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, preview)
}

func (s *Server) handleUnlink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path  string `json:"path"`
		Force bool   `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		jsonError(w, "path is required", http.StatusBadRequest)
		return
	}

	result, err := linker.Unlink(req.Path, req.Force)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, result)
}

func (s *Server) handleUndoPreview(w http.ResponseWriter, r *http.Request) {
	preview, entry, err := linker.UndoPreview()
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonOK(w, map[string]interface{}{
		"preview": preview,
		"entry":   entry,
	})
}

func (s *Server) handleUndo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Force bool `json:"force"`
	}
	// Body is optional — if empty or invalid, force defaults to false
	json.NewDecoder(r.Body).Decode(&req)

	result, entry, err := linker.Undo(req.Force)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonOK(w, map[string]interface{}{
		"result": result,
		"entry":  entry,
	})
}
