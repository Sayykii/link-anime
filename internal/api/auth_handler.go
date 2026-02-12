package api

import (
	"encoding/json"
	"net/http"

	"link-anime/internal/auth"
)

type loginRequest struct {
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if !auth.CheckPassword(req.Password) {
		jsonError(w, "invalid password", http.StatusUnauthorized)
		return
	}

	token, err := auth.CreateSession()
	if err != nil {
		jsonError(w, "session error", http.StatusInternalServerError)
		return
	}

	auth.SetSessionCookie(w, token)
	jsonOK(w, map[string]bool{"ok": true})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("link-anime-session")
	if err == nil {
		auth.DestroySession(cookie.Value)
	}
	auth.ClearSessionCookie(w)
	jsonOK(w, map[string]bool{"ok": true})
}

func (s *Server) handleAuthCheck(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("link-anime-session")
	if err != nil || !auth.ValidateSession(cookie.Value) {
		jsonOK(w, map[string]bool{"authenticated": false})
		return
	}
	jsonOK(w, map[string]bool{"authenticated": true})
}

// --- JSON helpers ---

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
