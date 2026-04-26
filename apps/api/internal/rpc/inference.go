package rpc

import "context"

type GenerateRequest struct {
	ModelName          string
	Text               string
	ReferenceAudioPath string
	ReferenceText      string
	OutputPath         string
	Language           string
	StylePrompt        string
}

type GenerateResult struct {
	OutputPath      string  `json:"outputPath"`
	LatencyMS       int64   `json:"latencyMs"`
	RealtimeFactor  float64 `json:"realtimeFactor"`
	ModelName       string  `json:"modelName"`
	SampleRate      int     `json:"sampleRate"`
	DurationSeconds float64 `json:"durationSeconds"`
}

type InferenceClient interface {
	Generate(ctx context.Context, req GenerateRequest) (GenerateResult, error)
}

type PlaceholderClient struct {
	URL string
}

func (c PlaceholderClient) Generate(ctx context.Context, req GenerateRequest) (GenerateResult, error) {
	return GenerateResult{}, ctx.Err()
}
