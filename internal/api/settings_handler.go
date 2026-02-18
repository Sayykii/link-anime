package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"link-anime/internal/auth"
	"link-anime/internal/database"
	"link-anime/internal/models"
)

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings := models.Settings{
		QbitURL:      settingOr("qbit_url", s.Config.QbitURL),
		QbitUser:     settingOr("qbit_user", s.Config.QbitUser),
		QbitPass:     settingOr("qbit_pass", s.Config.QbitPass),
		QbitCategory: settingOr("qbit_category", s.Config.QbitCategory),
		ShokoURL:     settingOr("shoko_url", s.Config.ShokoURL),
		ShokoAPIKey:  settingOr("shoko_apikey", s.Config.ShokoAPIKey),
		NotifyURL:    settingOr("notify_url", s.Config.NotifyURL),
		DownloadDir:  s.getDownloadDir(),
		MediaDir:     s.getMediaDir(),
		MoviesDir:    s.getMoviesDir(),
	}

	// Mask password
	if settings.QbitPass != "" {
		settings.QbitPass = "********"
	}

	jsonOK(w, settings)
}

func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req models.Settings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	pairs := map[string]string{
		"qbit_url":      req.QbitURL,
		"qbit_user":     req.QbitUser,
		"qbit_category": req.QbitCategory,
		"shoko_url":     req.ShokoURL,
		"shoko_apikey":  req.ShokoAPIKey,
		"notify_url":    req.NotifyURL,
		"download_dir":  req.DownloadDir,
		"media_dir":     req.MediaDir,
		"movies_dir":    req.MoviesDir,
	}

	// Only update qbit password if it's not the masked value
	if req.QbitPass != "" && req.QbitPass != "********" {
		pairs["qbit_pass"] = req.QbitPass
	}

	for key, value := range pairs {
		if err := database.SetSetting(key, value); err != nil {
			jsonError(w, fmt.Sprintf("failed to save %s: %v", key, err), http.StatusInternalServerError)
			return
		}
	}

	// Reinitialize clients with new settings
	s.reinitClients()

	jsonOK(w, map[string]bool{"ok": true})
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Current string `json:"current"`
		New     string `json:"new"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if !auth.CheckPassword(req.Current) {
		jsonError(w, "current password is incorrect", http.StatusUnauthorized)
		return
	}

	if len(req.New) < 4 {
		jsonError(w, "new password must be at least 4 characters", http.StatusBadRequest)
		return
	}

	if err := auth.ChangePassword(req.New); err != nil {
		jsonError(w, "failed to change password", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]bool{"ok": true})
}

// reinitClients updates qbit, shoko, and notifier with current DB settings.
func (s *Server) reinitClients() {
	qbitURL := settingOr("qbit_url", s.Config.QbitURL)
	qbitUser := settingOr("qbit_user", s.Config.QbitUser)
	qbitPass := settingOr("qbit_pass", s.Config.QbitPass)

	if qbitURL != "" {
		s.Qbit = newQbitClient(qbitURL, qbitUser, qbitPass)
		log.Printf("[settings] qBit client re-initialized: %s", qbitURL)
	}

	shokoURL := settingOr("shoko_url", s.Config.ShokoURL)
	shokoKey := settingOr("shoko_apikey", s.Config.ShokoAPIKey)
	if shokoURL != "" {
		s.Shoko = newShokoClient(shokoURL, shokoKey)
		log.Printf("[settings] Shoko client re-initialized: url=%s apikey=%v", shokoURL, shokoKey != "")
	} else {
		log.Printf("[settings] Shoko not configured (no URL)")
	}

	notifyURL := settingOr("notify_url", s.Config.NotifyURL)
	s.Notifier = newNotifier(notifyURL)
}

func settingOr(key, fallback string) string {
	if v, err := database.GetSetting(key); err == nil && v != "" {
		return v
	}
	return fallback
}

// Port returns the server port.
func (s *Server) Port() string {
	return ":" + strconv.Itoa(s.Config.Port)
}
