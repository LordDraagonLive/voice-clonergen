package httpapi

import (
	"encoding/json"
	"net/http"
	"slices"

	"personal-voice-cloner/apps/api/internal/voices"

	"github.com/go-chi/chi/v5"
)

func RegisterVoiceRoutes(r chi.Router, svc *voices.Service, maxUploadMB int64, allowedAudioTypes []string) {
	r.Post("/voice-profiles", func(w http.ResponseWriter, r *http.Request) {
		var input voices.CreateProfileInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		profile, err := svc.CreateProfile(r.Context(), input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, profile)
	})
	r.Get("/voice-profiles", func(w http.ResponseWriter, r *http.Request) {
		profiles, err := svc.ListProfiles(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, profiles)
	})
	r.Get("/voice-profiles/{id}", func(w http.ResponseWriter, r *http.Request) {
		profile, err := svc.GetProfile(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, profile)
	})
	r.Get("/voice-profiles/{id}/samples", func(w http.ResponseWriter, r *http.Request) {
		samples, err := svc.ListSamples(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, samples)
	})
	r.Post("/voice-profiles/{id}/samples", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadMB<<20)
		if err := r.ParseMultipartForm(maxUploadMB << 20); err != nil {
			writeError(w, http.StatusBadRequest, "invalid multipart form")
			return
		}
		file, header, err := r.FormFile("audio")
		if err != nil {
			writeError(w, http.StatusBadRequest, "audio file is required")
			return
		}
		defer file.Close()
		if contentType := header.Header.Get("Content-Type"); contentType != "" && !slices.Contains(allowedAudioTypes, contentType) {
			writeError(w, http.StatusBadRequest, "unsupported audio content type")
			return
		}
		sample, err := svc.AddSample(r.Context(), chi.URLParam(r, "id"), header.Filename, r.FormValue("transcript"), file)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, sample)
	})
}
