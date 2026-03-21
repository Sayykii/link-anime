package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"link-anime/internal/database"

	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName = "link-anime-session"
	sessionDuration   = 24 * time.Hour
)

type session struct {
	expiresAt time.Time
}

var (
	sessions = make(map[string]*session)
	mu       sync.RWMutex
)

// InitPassword hashes and stores the initial password if none exists.
func InitPassword(password string) error {
	existing, err := database.GetSetting("password_hash")
	if err != nil {
		return err
	}
	if existing != "" {
		return nil // already set
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return database.SetSetting("password_hash", string(hash))
}

// CheckPassword verifies a password against the stored hash.
func CheckPassword(password string) bool {
	hash, err := database.GetSetting("password_hash")
	if err != nil || hash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ChangePassword updates the stored password hash.
func ChangePassword(newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return database.SetSetting("password_hash", string(hash))
}

// CreateSession creates a new session and returns the token.
func CreateSession() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	mu.Lock()
	sessions[token] = &session{expiresAt: time.Now().Add(sessionDuration)}
	mu.Unlock()

	return token, nil
}

// ValidateSession checks if a session token is valid.
func ValidateSession(token string) bool {
	mu.RLock()
	s, exists := sessions[token]
	mu.RUnlock()

	if !exists {
		return false
	}
	if time.Now().After(s.expiresAt) {
		mu.Lock()
		delete(sessions, token)
		mu.Unlock()
		return false
	}
	return true
}

// DestroySession removes a session.
func DestroySession(token string) {
	mu.Lock()
	delete(sessions, token)
	mu.Unlock()
}

// GenerateAPIKey creates a new random API key and stores it in the database.
func GenerateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	key := "la_" + hex.EncodeToString(b)
	if err := database.SetSetting("api_key", key); err != nil {
		return "", err
	}
	return key, nil
}

// GetAPIKey returns the current API key, or empty if none exists.
func GetAPIKey() string {
	key, err := database.GetSetting("api_key")
	if err != nil {
		return ""
	}
	return key
}

// DeleteAPIKey removes the stored API key.
func DeleteAPIKey() error {
	return database.SetSetting("api_key", "")
}

// validateAPIKey checks if the provided key matches the stored API key.
func validateAPIKey(key string) bool {
	if key == "" {
		return false
	}
	stored := GetAPIKey()
	return stored != "" && stored == key
}

// Middleware protects routes requiring authentication.
// Accepts either a session cookie or an X-API-Key header.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check API key first (for external tools like userscripts)
		if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
			if validateAPIKey(apiKey) {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, `{"error":"invalid api key"}`, http.StatusUnauthorized)
			return
		}

		// Fall back to session cookie
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || !ValidateSession(cookie.Value) {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SetSessionCookie sets the session cookie on the response.
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionDuration.Seconds()),
	})
}

// ClearSessionCookie removes the session cookie.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

// CleanupExpiredSessions removes expired sessions periodically.
func CleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for token, s := range sessions {
				if now.After(s.expiresAt) {
					delete(sessions, token)
				}
			}
			mu.Unlock()
		}
	}()
}
