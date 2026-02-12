package api

import (
	"net/http"

	"link-anime/internal/scanner"
)

func (s *Server) handleGetShows(w http.ResponseWriter, r *http.Request) {
	mediaDir := s.getMediaDir()
	shows, err := scanner.ScanLibrary(mediaDir)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, shows)
}

func (s *Server) handleGetMovies(w http.ResponseWriter, r *http.Request) {
	moviesDir := s.getMoviesDir()
	movies, err := scanner.ScanMovies(moviesDir)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, movies)
}

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	mediaDir := s.getMediaDir()
	moviesDir := s.getMoviesDir()

	shows, _ := scanner.ScanLibrary(mediaDir)
	movies, _ := scanner.ScanMovies(moviesDir)
	size := scanner.LibrarySize(mediaDir, moviesDir)

	totalSeasons := 0
	totalEpisodes := 0
	for _, show := range shows {
		totalSeasons += len(show.Seasons)
		totalEpisodes += show.Episodes
	}

	jsonOK(w, map[string]interface{}{
		"shows":    len(shows),
		"seasons":  totalSeasons,
		"episodes": totalEpisodes,
		"movies":   len(movies),
		"size":     size,
	})
}
