package generation

import (
	"context"
	"errors"
	"strings"
	"time"

	"personal-voice-cloner/apps/api/internal/models"
	"personal-voice-cloner/apps/api/internal/queue"
	"personal-voice-cloner/apps/api/internal/storage"

	"github.com/google/uuid"
)

type Service struct {
	repo         *Repository
	queue        queue.Queue
	store        storage.Store
	defaultModel string
}

func NewService(repo *Repository, q queue.Queue, store storage.Store, defaultModel string) *Service {
	return &Service{repo: repo, queue: q, store: store, defaultModel: defaultModel}
}

type CreateGenerationInput struct {
	VoiceProfileID string `json:"voiceProfileId"`
	ModelName      string `json:"modelName"`
	Text           string `json:"text"`
	Format         string `json:"format"`
}

func (s *Service) CreateGeneration(ctx context.Context, in CreateGenerationInput) (models.GenerationJob, error) {
	if strings.TrimSpace(in.VoiceProfileID) == "" {
		return models.GenerationJob{}, errors.New("voiceProfileId is required")
	}
	if strings.TrimSpace(in.Text) == "" {
		return models.GenerationJob{}, errors.New("text is required")
	}
	model := in.ModelName
	if model == "" {
		model = s.defaultModel
	}
	job := models.GenerationJob{
		ID:              uuid.NewString(),
		VoiceProfileID:  in.VoiceProfileID,
		ModelName:       model,
		InputText:       strings.TrimSpace(in.Text),
		Status:          "pending",
		ProgressMessage: "Waiting for inference worker",
		CreatedAt:       time.Now().UTC(),
	}
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return models.GenerationJob{}, err
	}
	return job, s.queue.Enqueue(ctx, queue.Job{Type: "generate_speech", ID: job.ID})
}

func (s *Service) GetJob(ctx context.Context, id string) (models.GenerationJob, error) {
	return s.repo.GetJob(ctx, id)
}

func (s *Service) ListJobs(ctx context.Context) ([]models.GenerationJob, error) {
	return s.repo.ListJobs(ctx)
}

func (s *Service) OpenOutput(ctx context.Context, id string) (string, error) {
	job, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return "", err
	}
	if job.OutputFilePath == "" {
		return "", errors.New("generation output is not ready")
	}
	return job.OutputFilePath, nil
}

type BenchmarkInput struct {
	VoiceProfileID string   `json:"voiceProfileId"`
	Text           string   `json:"text"`
	Models         []string `json:"models"`
}

func (s *Service) CreateBenchmark(ctx context.Context, in BenchmarkInput) (models.BenchmarkRun, error) {
	if strings.TrimSpace(in.VoiceProfileID) == "" || strings.TrimSpace(in.Text) == "" {
		return models.BenchmarkRun{}, errors.New("voiceProfileId and text are required")
	}
	if len(in.Models) == 0 {
		in.Models = []string{models.ModelQwen3TTS, models.ModelXTTSv2}
	}
	run := models.BenchmarkRun{
		ID:             uuid.NewString(),
		VoiceProfileID: in.VoiceProfileID,
		InputText:      strings.TrimSpace(in.Text),
		ModelsTested:   in.Models,
		ResultsJSON:    `{"status":"queued"}`,
		CreatedAt:      time.Now().UTC(),
	}
	if err := s.repo.CreateBenchmark(ctx, run); err != nil {
		return models.BenchmarkRun{}, err
	}
	return run, s.queue.Enqueue(ctx, queue.Job{Type: "benchmark_models", ID: run.ID})
}
