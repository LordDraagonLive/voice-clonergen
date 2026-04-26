package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"personal-voice-cloner/apps/api/internal/config"
	"personal-voice-cloner/apps/api/internal/db"
	"personal-voice-cloner/apps/api/internal/generation"
	httpapi "personal-voice-cloner/apps/api/internal/http"
	"personal-voice-cloner/apps/api/internal/queue"
	"personal-voice-cloner/apps/api/internal/storage"
	"personal-voice-cloner/apps/api/internal/voices"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Error("parse redis url", "error", err)
		os.Exit(1)
	}
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()

	var store storage.Store = storage.NewLocalStore(cfg.StorageLocalPath)
	q := queue.NewRedisQueue(redisClient, "voice-cloner-jobs", log)

	voiceRepo := voices.NewRepository(pool)
	genRepo := generation.NewRepository(pool)
	voiceSvc := voices.NewService(voiceRepo, store, q, cfg.DefaultModel)
	genSvc := generation.NewService(genRepo, q, store, cfg.DefaultModel)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewRouter(cfg, voiceSvc, genSvc),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Info("api listening", "addr", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("api stopped", "error", err)
		os.Exit(1)
	}
}
