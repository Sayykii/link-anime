package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"link-anime/internal/database"
	"link-anime/internal/models"

	"github.com/go-chi/chi/v5"
)

type createJobRequest struct {
	InputPath string `json:"inputPath"`
	Preset    string `json:"preset"`
}

// handleListUpscaleJobs returns all upscale jobs.
func (s *Server) handleListUpscaleJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := database.ListJobs()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if jobs == nil {
		jobs = []models.UpscaleJob{}
	}
	jsonOK(w, jobs)
}

// handleCreateUpscaleJob creates a new upscale job.
func (s *Server) handleCreateUpscaleJob(w http.ResponseWriter, r *http.Request) {
	var req createJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input path exists
	if req.InputPath == "" {
		jsonError(w, "inputPath is required", http.StatusBadRequest)
		return
	}
	if _, err := os.Stat(req.InputPath); os.IsNotExist(err) {
		jsonError(w, "input file does not exist", http.StatusBadRequest)
		return
	}

	// Validate preset
	validPresets := map[string]bool{"fast": true, "balanced": true, "quality": true}
	if !validPresets[req.Preset] {
		jsonError(w, "preset must be fast, balanced, or quality", http.StatusBadRequest)
		return
	}

	// Generate output path: replace extension with _4k.mkv
	ext := filepath.Ext(req.InputPath)
	outputPath := strings.TrimSuffix(req.InputPath, ext) + "_4k.mkv"

	job, err := database.CreateJob(req.InputPath, outputPath, req.Preset)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, job)
}

// handleGetUpscaleJob returns a single upscale job by ID.
func (s *Server) handleGetUpscaleJob(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid job ID", http.StatusBadRequest)
		return
	}

	job, err := database.GetJob(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if job == nil {
		jsonError(w, "job not found", http.StatusNotFound)
		return
	}

	jsonOK(w, job)
}

// handleDeleteUpscaleJob removes an upscale job by ID.
func (s *Server) handleDeleteUpscaleJob(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid job ID", http.StatusBadRequest)
		return
	}

	// Check job exists and status
	job, err := database.GetJob(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if job == nil {
		jsonError(w, "job not found", http.StatusNotFound)
		return
	}
	if job.Status == models.UpscaleStatusRunning {
		jsonError(w, "cannot delete running job", http.StatusBadRequest)
		return
	}

	if err := database.DeleteJob(id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]bool{"deleted": true})
}
