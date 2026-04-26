package voices

import (
	"context"

	"personal-voice-cloner/apps/api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateProfile(ctx context.Context, p models.VoiceProfile) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO voice_profiles (id, name, description, model_default, consent_confirmed, consent_text, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, p.ID, p.Name, p.Description, p.ModelDefault, p.ConsentConfirmed, p.ConsentText, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *Repository) ListProfiles(ctx context.Context) ([]models.VoiceProfile, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, name, description, model_default, consent_confirmed, consent_text, created_at, updated_at
		FROM voice_profiles
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make([]models.VoiceProfile, 0)
	for rows.Next() {
		var p models.VoiceProfile
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.ModelDefault, &p.ConsentConfirmed, &p.ConsentText, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

func (r *Repository) GetProfile(ctx context.Context, id string) (models.VoiceProfile, error) {
	var p models.VoiceProfile
	err := r.db.QueryRow(ctx, `
		SELECT id::text, name, description, model_default, consent_confirmed, consent_text, created_at, updated_at
		FROM voice_profiles
		WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.ModelDefault, &p.ConsentConfirmed, &p.ConsentText, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *Repository) CreateSample(ctx context.Context, s models.VoiceSample) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO voice_samples (id, voice_profile_id, original_file_path, cleaned_file_path, status, error_message, duration_seconds, sample_rate, quality_score, transcript_optional, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, s.ID, s.VoiceProfileID, s.OriginalFilePath, s.CleanedFilePath, s.Status, s.ErrorMessage, s.DurationSeconds, s.SampleRate, s.QualityScore, s.TranscriptOptional, s.CreatedAt)
	return err
}

func (r *Repository) ListSamples(ctx context.Context, profileID string) ([]models.VoiceSample, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, voice_profile_id::text, original_file_path, cleaned_file_path, status, error_message, duration_seconds, sample_rate, quality_score, transcript_optional, created_at
		FROM voice_samples
		WHERE voice_profile_id = $1
		ORDER BY created_at DESC
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	samples := make([]models.VoiceSample, 0)
	for rows.Next() {
		var s models.VoiceSample
		if err := rows.Scan(&s.ID, &s.VoiceProfileID, &s.OriginalFilePath, &s.CleanedFilePath, &s.Status, &s.ErrorMessage, &s.DurationSeconds, &s.SampleRate, &s.QualityScore, &s.TranscriptOptional, &s.CreatedAt); err != nil {
			return nil, err
		}
		samples = append(samples, s)
	}
	return samples, rows.Err()
}
