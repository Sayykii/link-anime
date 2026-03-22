package api

import (
	"net/http"
	"strconv"

	"link-anime/internal/scanner"
	"link-anime/internal/shoko"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleShokoSeries(w http.ResponseWriter, r *http.Request) {
	if s.Shoko == nil || !s.Shoko.IsConfigured() {
		jsonError(w, "Shoko not configured", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	series, total, err := s.Shoko.GetSeries(page, pageSize)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"series": series,
		"total":  total,
	})
}

func (s *Server) handleShokoSeriesSearch(w http.ResponseWriter, r *http.Request) {
	if s.Shoko == nil || !s.Shoko.IsConfigured() {
		jsonError(w, "Shoko not configured", http.StatusBadRequest)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		jsonError(w, "q parameter required", http.StatusBadRequest)
		return
	}

	results, err := s.Shoko.SearchSeries(query)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if results == nil {
		jsonOK(w, []interface{}{})
		return
	}

	jsonOK(w, results)
}

func (s *Server) handleShokoSeriesDetail(w http.ResponseWriter, r *http.Request) {
	if s.Shoko == nil || !s.Shoko.IsConfigured() {
		jsonError(w, "Shoko not configured", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid series ID", http.StatusBadRequest)
		return
	}

	series, err := s.Shoko.GetSeriesDetails(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, series)
}

func (s *Server) handleShokoSeriesEpisodes(w http.ResponseWriter, r *http.Request) {
	if s.Shoko == nil || !s.Shoko.IsConfigured() {
		jsonError(w, "Shoko not configured", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid series ID", http.StatusBadRequest)
		return
	}

	includeMissing := r.URL.Query().Get("includeMissing") == "true"

	episodes, err := s.Shoko.GetSeriesEpisodes(id, includeMissing)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if episodes == nil {
		jsonOK(w, []interface{}{})
		return
	}

	jsonOK(w, episodes)
}

func (s *Server) handleShokoDashboard(w http.ResponseWriter, r *http.Request) {
	if s.Shoko == nil || !s.Shoko.IsConfigured() {
		jsonError(w, "Shoko not configured", http.StatusBadRequest)
		return
	}

	// Fetch all dashboard data concurrently
	type result struct {
		key string
		val interface{}
		err error
	}

	ch := make(chan result, 4)

	go func() {
		stats, err := s.Shoko.GetDashboardStats()
		ch <- result{"stats", stats, err}
	}()
	go func() {
		recent, err := s.Shoko.GetRecentlyAddedEpisodes(10)
		ch <- result{"recentlyAdded", recent, err}
	}()
	go func() {
		cw, err := s.Shoko.GetContinueWatching(10)
		ch <- result{"continueWatching", cw, err}
	}()
	go func() {
		cal, err := s.Shoko.GetCalendar(7)
		ch <- result{"calendar", cal, err}
	}()

	data := map[string]interface{}{}
	for i := 0; i < 4; i++ {
		r := <-ch
		if r.err != nil {
			// Don't fail the whole response if one section errors
			data[r.key] = nil
		} else {
			data[r.key] = r.val
		}
	}

	jsonOK(w, data)
}

func (s *Server) handleShokoFolderMap(w http.ResponseWriter, r *http.Request) {
	if s.Shoko == nil || !s.Shoko.IsConfigured() {
		jsonError(w, "Shoko not configured", http.StatusBadRequest)
		return
	}

	mediaDir := s.getMediaDir()
	shows, err := scanner.ScanLibrary(mediaDir)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	folderNames := make([]string, len(shows))
	for i, show := range shows {
		folderNames[i] = show.Name
	}

	seriesMap, err := s.Shoko.GetFolderSeriesMap(folderNames)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, shoko.BuildFolderMappingResponse(seriesMap))
}

func (s *Server) handleShokoImage(w http.ResponseWriter, r *http.Request) {
	if s.Shoko == nil || !s.Shoko.IsConfigured() {
		http.Error(w, "Shoko not configured", http.StatusBadRequest)
		return
	}

	source := chi.URLParam(r, "source")
	imageType := chi.URLParam(r, "type")
	id := chi.URLParam(r, "id")

	data, contentType, err := s.Shoko.GetImage(source, imageType, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400") // cache 24h
	w.Write(data)
}
