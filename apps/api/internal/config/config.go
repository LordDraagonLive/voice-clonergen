package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr          string
	DatabaseURL       string
	RedisURL          string
	StorageProvider   string
	StorageLocalPath  string
	S3Endpoint        string
	S3AccessKey       string
	S3SecretKey       string
	S3Bucket          string
	InferenceRPCURL   string
	DefaultModel      string
	MaxUploadMB       int64
	AllowedAudioTypes []string
}

func Load() Config {
	return Config{
		HTTPAddr:          env("HTTP_ADDR", ":8080"),
		DatabaseURL:       env("DATABASE_URL", "postgres://voice:voice@localhost:5432/voice_cloner?sslmode=disable"),
		RedisURL:          env("REDIS_URL", "redis://localhost:6379/0"),
		StorageProvider:   env("STORAGE_PROVIDER", "local"),
		StorageLocalPath:  env("STORAGE_LOCAL_PATH", "./data/audio"),
		S3Endpoint:        os.Getenv("S3_ENDPOINT"),
		S3AccessKey:       os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:       os.Getenv("S3_SECRET_KEY"),
		S3Bucket:          os.Getenv("S3_BUCKET"),
		InferenceRPCURL:   env("INFERENCE_RPC_URL", "http://localhost:50051"),
		DefaultModel:      env("DEFAULT_MODEL", "qwen3-tts"),
		MaxUploadMB:       envInt("MAX_UPLOAD_MB", 50),
		AllowedAudioTypes: split(env("ALLOWED_AUDIO_TYPES", "audio/wav,audio/mpeg,audio/mp4,audio/flac,audio/x-m4a")),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func split(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
