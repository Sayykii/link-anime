package api

import (
	"net/http"
)

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	s.Hub.HandleWS(w, r)
}
