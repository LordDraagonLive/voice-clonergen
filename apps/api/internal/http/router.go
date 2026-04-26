package httpapi

import (
	"net/http"

	"personal-voice-cloner/apps/api/internal/auth"
	"personal-voice-cloner/apps/api/internal/config"
	"personal-voice-cloner/apps/api/internal/generation"
	"personal-voice-cloner/apps/api/internal/voices"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func NewRouter(cfg config.Config, voiceSvc *voices.Service, genSvc *generation.Service) http.Handler {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Route("/api", func(r chi.Router) {
		r.Use(auth.Middleware)
		RegisterVoiceRoutes(r, voiceSvc, cfg.MaxUploadMB, cfg.AllowedAudioTypes)
		RegisterGenerationRoutes(r, genSvc)
	})
	return r
}
