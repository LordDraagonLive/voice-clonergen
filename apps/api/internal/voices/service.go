package voices

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"personal-voice-cloner/apps/api/internal/models"
	"personal-voice-cloner/apps/api/internal/queue"
	"personal-voice-cloner/apps/api/internal/storage"

	"github.com/google/uuid"
)

var ErrConsentRequired = errors.New("voice profile consent confirmation is required")

type Service struct {
	repo         *Repository
	store        storage.Store
	queue        queue.Queue
	defaultModel string
}

func NewService(repo *Repository, store storage.Store, queue queue.Queue, defaultModel string) *Service {
	return &Service{repo: repo, store: store, queue: queue, defaultModel: defaultModel}
}

type CreateProfileInput struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	ModelDefault     string `json:"modelDefault"`
	ConsentConfirmed bool   `json:"consentConfirmed"`
	ConsentText      string `json:"consentText"`
}

func (s *Service) CreateProfile(ctx context.Context, in CreateProfileInput) (models.VoiceProfile, error) {
	if strings.TrimSpace(in.Name) == "" {
		return models.VoiceProfile{}, errors.New("name is required")
	}
	if !in.ConsentConfirmed {
		return models.VoiceProfile{}, ErrConsentRequired
	}
	if strings.TrimSpace(in.ConsentText) == "" {
		return models.VoiceProfile{}, errors.New("consent text is required")
	}
	model := in.ModelDefault
	if model == "" {
		model = s.defaultModel
	}
	now := time.Now().UTC()
	profile := models.VoiceProfile{
		ID:               uuid.NewString(),
		Name:             strings.TrimSpace(in.Name),
		Description:      strings.TrimSpace(in.Description),
		ModelDefault:     model,
		ConsentConfirmed: true,
		ConsentText:      strings.TrimSpace(in.ConsentText),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	return profile, s.repo.CreateProfile(ctx, profile)
}

func (s *Service) ListProfiles(ctx context.Context) ([]models.VoiceProfile, error) {
	return s.repo.ListProfiles(ctx)
}

func (s *Service) GetProfile(ctx context.Context, id string) (models.VoiceProfile, error) {
	return s.repo.GetProfile(ctx, id)
}

func (s *Service) AddSample(ctx context.Context, profileID, filename, transcript string, r io.Reader) (models.VoiceSample, error) {
	if _, err := s.repo.GetProfile(ctx, profileID); err != nil {
		return models.VoiceSample{}, err
	}
	id := uuid.NewString()
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".wav"
	}
	key := fmt.Sprintf("voices/%s/samples/%s/original%s", profileID, id, ext)
	path, err := s.store.Save(ctx, key, r)
	if err != nil {
		return models.VoiceSample{}, err
	}
	sample := models.VoiceSample{
		ID:                 id,
		VoiceProfileID:     profileID,
		OriginalFilePath:   path,
		Status:             "pending",
		TranscriptOptional: strings.TrimSpace(transcript),
		CreatedAt:          time.Now().UTC(),
	}
	if err := s.repo.CreateSample(ctx, sample); err != nil {
		return models.VoiceSample{}, err
	}
	_ = s.queue.Enqueue(ctx, queue.Job{Type: "preprocess_sample", ID: sample.ID, Data: map[string]any{
		"voiceProfileId":    profileID,
		"originalFilePath":  sample.OriginalFilePath,
		"transcript":        sample.TranscriptOptional,
		"cleanedOutputPath": fmt.Sprintf("voices/%s/samples/%s/cleaned.wav", profileID, id),
	}})
	return sample, nil
}

func (s *Service) ListSamples(ctx context.Context, profileID string) ([]models.VoiceSample, error) {
	if _, err := s.repo.GetProfile(ctx, profileID); err != nil {
		return nil, err
	}
	return s.repo.ListSamples(ctx, profileID)
}
