package database

import (
	"database/sql"
	"fmt"

	"link-anime/internal/models"
)

// GetNextPendingJob returns the oldest pending upscale job, or nil if none exists.
func GetNextPendingJob() (*models.UpscaleJob, error) {
	row := DB.QueryRow(`
		SELECT id, input_path, output_path, preset, status, error, created_at, started_at, completed_at
		FROM upscale_jobs
		WHERE status = ?
		ORDER BY created_at ASC
		LIMIT 1
	`, models.UpscaleStatusPending)

	var job models.UpscaleJob
	var errMsg sql.NullString
	var startedAt, completedAt sql.NullTime

	err := row.Scan(
		&job.ID,
		&job.InputPath,
		&job.OutputPath,
		&job.Preset,
		&job.Status,
		&errMsg,
		&job.CreatedAt,
		&startedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan job: %w", err)
	}

	if errMsg.Valid {
		job.Error = &errMsg.String
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

// UpdateJobStatus updates a job's status and sets appropriate timestamps.
func UpdateJobStatus(id int64, status string, errMsg *string) error {
	var query string
	var args []interface{}

	switch status {
	case models.UpscaleStatusRunning:
		query = `UPDATE upscale_jobs SET status = ?, started_at = CURRENT_TIMESTAMP WHERE id = ?`
		args = []interface{}{status, id}
	case models.UpscaleStatusCompleted:
		query = `UPDATE upscale_jobs SET status = ?, completed_at = CURRENT_TIMESTAMP WHERE id = ?`
		args = []interface{}{status, id}
	case models.UpscaleStatusFailed:
		query = `UPDATE upscale_jobs SET status = ?, error = ?, completed_at = CURRENT_TIMESTAMP WHERE id = ?`
		args = []interface{}{status, errMsg, id}
	case models.UpscaleStatusPending:
		query = `UPDATE upscale_jobs SET status = ?, started_at = NULL WHERE id = ?`
		args = []interface{}{status, id}
	default:
		query = `UPDATE upscale_jobs SET status = ? WHERE id = ?`
		args = []interface{}{status, id}
	}

	_, err := DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("update job status: %w", err)
	}
	return nil
}

// ResetRunningJob atomically resets a running job back to pending.
// Returns nil if the job was already completed/failed (no reset needed).
func ResetRunningJob(id int64) error {
	result, err := DB.Exec(`
		UPDATE upscale_jobs 
		SET status = ?, started_at = NULL 
		WHERE id = ? AND status = ?
	`, models.UpscaleStatusPending, id, models.UpscaleStatusRunning)

	if err != nil {
		return fmt.Errorf("reset job: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	// 0 rows means job already completed/failed - no reset needed
	_ = rows
	return nil
}

// ListJobs returns all upscale jobs sorted by created_at DESC (newest first).
func ListJobs() ([]models.UpscaleJob, error) {
	rows, err := DB.Query(`
		SELECT id, input_path, output_path, preset, status, error, created_at, started_at, completed_at
		FROM upscale_jobs
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()

	var jobs []models.UpscaleJob
	for rows.Next() {
		var job models.UpscaleJob
		var errMsg sql.NullString
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&job.ID,
			&job.InputPath,
			&job.OutputPath,
			&job.Preset,
			&job.Status,
			&errMsg,
			&job.CreatedAt,
			&startedAt,
			&completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}

		if errMsg.Valid {
			job.Error = &errMsg.String
		}
		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate jobs: %w", err)
	}

	return jobs, nil
}

// GetJob returns a single upscale job by ID, or nil if not found.
func GetJob(id int64) (*models.UpscaleJob, error) {
	row := DB.QueryRow(`
		SELECT id, input_path, output_path, preset, status, error, created_at, started_at, completed_at
		FROM upscale_jobs
		WHERE id = ?
	`, id)

	var job models.UpscaleJob
	var errMsg sql.NullString
	var startedAt, completedAt sql.NullTime

	err := row.Scan(
		&job.ID,
		&job.InputPath,
		&job.OutputPath,
		&job.Preset,
		&job.Status,
		&errMsg,
		&job.CreatedAt,
		&startedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get job: %w", err)
	}

	if errMsg.Valid {
		job.Error = &errMsg.String
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

// CreateJob inserts a new pending upscale job and returns the created job.
func CreateJob(inputPath, outputPath, preset string) (*models.UpscaleJob, error) {
	result, err := DB.Exec(`
		INSERT INTO upscale_jobs (input_path, output_path, preset, status, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, inputPath, outputPath, preset, models.UpscaleStatusPending)

	if err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	return GetJob(id)
}

// DeleteJob removes an upscale job by ID.
func DeleteJob(id int64) error {
	_, err := DB.Exec(`DELETE FROM upscale_jobs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete job: %w", err)
	}
	return nil
}
