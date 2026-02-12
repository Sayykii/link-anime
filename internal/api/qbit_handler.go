package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleQbitTorrents(w http.ResponseWriter, r *http.Request) {
	if s.Qbit == nil || !s.Qbit.IsConfigured() {
		jsonError(w, "qBittorrent not configured", http.StatusBadRequest)
		return
	}

	category := r.URL.Query().Get("category")
	if category == "" {
		category = settingOr("qbit_category", s.Config.QbitCategory)
	}

	torrents, err := s.Qbit.ListTorrents(category)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if torrents == nil {
		jsonOK(w, []interface{}{})
		return
	}

	jsonOK(w, torrents)
}

func (s *Server) handleQbitAdd(w http.ResponseWriter, r *http.Request) {
	if s.Qbit == nil || !s.Qbit.IsConfigured() {
		jsonError(w, "qBittorrent not configured", http.StatusBadRequest)
		return
	}

	var req struct {
		Magnet   string `json:"magnet"`
		Category string `json:"category"`
		SavePath string `json:"savePath"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Magnet == "" {
		jsonError(w, "magnet is required", http.StatusBadRequest)
		return
	}

	if req.Category == "" {
		req.Category = settingOr("qbit_category", s.Config.QbitCategory)
	}

	if err := s.Qbit.AddMagnet(req.Magnet, req.Category, req.SavePath); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]bool{"ok": true})
}

func (s *Server) handleQbitDelete(w http.ResponseWriter, r *http.Request) {
	if s.Qbit == nil || !s.Qbit.IsConfigured() {
		jsonError(w, "qBittorrent not configured", http.StatusBadRequest)
		return
	}

	var req struct {
		Hash        string `json:"hash"`
		DeleteFiles bool   `json:"deleteFiles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Hash == "" {
		jsonError(w, "hash is required", http.StatusBadRequest)
		return
	}

	if err := s.Qbit.DeleteTorrent(req.Hash, req.DeleteFiles); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]bool{"ok": true})
}

func (s *Server) handleQbitTest(w http.ResponseWriter, r *http.Request) {
	if s.Qbit == nil || !s.Qbit.IsConfigured() {
		jsonError(w, "qBittorrent not configured", http.StatusBadRequest)
		return
	}

	if err := s.Qbit.Login(); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]bool{"ok": true})
}
