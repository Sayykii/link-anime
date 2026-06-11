package qbit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr bool
	}{
		{"legacy success 200 Ok.", 200, "Ok.", false},
		{"qbit 5.2+ success 204 no content", 204, "", false},
		{"legacy bad creds 200 Fails.", 200, "Fails.", true},
		{"qbit 5.2+ bad creds 401", 401, "Unauthorized", true},
		{"banned 403", 403, "Forbidden", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v2/auth/login" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			c := New(srv.URL, "admin", "pass")
			err := c.Login()
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
