package generation

import (
	"context"

	"personal-voice-cloner/apps/api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateJob(ctx context.Context, job models.GenerationJob) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO generation_jobs (id, voice_profile_id, model_name, input_text, status, error_message, progress_message, output_file_path, latency_ms, realtime_factor, created_at, completed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`, job.ID, job.VoiceProfileID, job.ModelName, job.InputText, job.Status, job.ErrorMessage, job.ProgressMessage, job.OutputFilePath, job.LatencyMS, job.RealtimeFactor, job.CreatedAt, job.CompletedAt)
	return err
}

func (r *Repository) GetJob(ctx context.Context, id string) (models.GenerationJob, error) {
	var job models.GenerationJob
	err := r.db.QueryRow(ctx, `
		SELECT id::text, voice_profile_id::text, model_name, input_text, status, error_message, progress_message, output_file_path, latency_ms, realtime_factor, created_at, completed_at
		FROM generation_jobs WHERE id = $1
	`, id).Scan(&job.ID, &job.VoiceProfileID, &job.ModelName, &job.InputText, &job.Status, &job.ErrorMessage, &job.ProgressMessage, &job.OutputFilePath, &job.LatencyMS, &job.RealtimeFactor, &job.CreatedAt, &job.CompletedAt)
	return job, err
}

func (r *Repository) ListJobs(ctx context.Context) ([]models.GenerationJob, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, voice_profile_id::text, model_name, input_text, status, error_message, progress_message, output_file_path, latency_ms, realtime_factor, created_at, completed_at
		FROM generation_jobs ORDER BY created_at DESC LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := make([]models.GenerationJob, 0)
	for rows.Next() {
		var job models.GenerationJob
		if err := rows.Scan(&job.ID, &job.VoiceProfileID, &job.ModelName, &job.InputText, &job.Status, &job.ErrorMessage, &job.ProgressMessage, &job.OutputFilePath, &job.LatencyMS, &job.RealtimeFactor, &job.CreatedAt, &job.CompletedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (r *Repository) CreateBenchmark(ctx context.Context, b models.BenchmarkRun) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO benchmark_runs (id, voice_profile_id, input_text, models_tested, results_json, created_at)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6)
	`, b.ID, b.VoiceProfileID, b.InputText, b.ModelsTested, b.ResultsJSON, b.CreatedAt)
	return err
}

func isNoRows(err error) bool {
	return err == pgx.ErrNoRows
}
