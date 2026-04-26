package models

import "time"

const (
	ModelQwen3TTS = "qwen3-tts"
	ModelXTTSv2   = "xtts-v2"
)

type VoiceProfile struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	ModelDefault     string    `json:"modelDefault"`
	ConsentConfirmed bool      `json:"consentConfirmed"`
	ConsentText      string    `json:"consentText"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type VoiceSample struct {
	ID                 string    `json:"id"`
	VoiceProfileID     string    `json:"voiceProfileId"`
	OriginalFilePath   string    `json:"originalFilePath"`
	CleanedFilePath    string    `json:"cleanedFilePath"`
	Status             string    `json:"status"`
	ErrorMessage       string    `json:"errorMessage"`
	DurationSeconds    float64   `json:"durationSeconds"`
	SampleRate         int       `json:"sampleRate"`
	QualityScore       float64   `json:"qualityScore"`
	TranscriptOptional string    `json:"transcriptOptional"`
	CreatedAt          time.Time `json:"createdAt"`
}

type GenerationJob struct {
	ID              string     `json:"id"`
	VoiceProfileID  string     `json:"voiceProfileId"`
	ModelName       string     `json:"modelName"`
	InputText       string     `json:"inputText"`
	Status          string     `json:"status"`
	ErrorMessage    string     `json:"errorMessage"`
	ProgressMessage string     `json:"progressMessage"`
	OutputFilePath  string     `json:"outputFilePath"`
	LatencyMS       int64      `json:"latencyMs"`
	RealtimeFactor  float64    `json:"realtimeFactor"`
	CreatedAt       time.Time  `json:"createdAt"`
	CompletedAt     *time.Time `json:"completedAt,omitempty"`
}

type BenchmarkRun struct {
	ID             string    `json:"id"`
	VoiceProfileID string    `json:"voiceProfileId"`
	InputText      string    `json:"inputText"`
	ModelsTested   []string  `json:"modelsTested"`
	ResultsJSON    string    `json:"resultsJson"`
	CreatedAt      time.Time `json:"createdAt"`
}
