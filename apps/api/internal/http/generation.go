package httpapi

import (
	"encoding/json"
	"net/http"

	"personal-voice-cloner/apps/api/internal/generation"

	"github.com/go-chi/chi/v5"
)

func RegisterGenerationRoutes(r chi.Router, svc *generation.Service) {
	r.Post("/generations", func(w http.ResponseWriter, r *http.Request) {
		var input generation.CreateGenerationInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		job, err := svc.CreateGeneration(r.Context(), input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, job)
	})
	r.Get("/generations", func(w http.ResponseWriter, r *http.Request) {
		jobs, err := svc.ListJobs(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, jobs)
	})
	r.Get("/generations/{id}", func(w http.ResponseWriter, r *http.Request) {
		job, err := svc.GetJob(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, job)
	})
	r.Get("/generations/{id}/download", func(w http.ResponseWriter, r *http.Request) {
		path, err := svc.OpenOutput(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		http.ServeFile(w, r, path)
	})
	r.Post("/benchmarks", func(w http.ResponseWriter, r *http.Request) {
		var input generation.BenchmarkInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		run, err := svc.CreateBenchmark(r.Context(), input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, run)
	})
}
