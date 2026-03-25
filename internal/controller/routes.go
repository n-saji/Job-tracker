package controller

import (
	"net/http"
	"time"

	"job_tracker_be/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(jobService *service.JobService, requestTimeout time.Duration) http.Handler {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			origin := req.Header.Get("Origin")
			if origin == "http://localhost:3000" || origin == "http://localhost:5678" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}


			if req.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, req)
		})
	})
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(requestTimeout))

	jobController := NewJobController(jobService, requestTimeout)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/jobs", func(r chi.Router) {
		r.Post("/", jobController.CreateJob)
		r.Post("/bulk-delete", jobController.BulkDeleteJobs)
		r.Get("/", jobController.ListJobs)
		r.Get("/exists", jobController.ExistsByApplyLink)
		r.Get("/{id}", jobController.GetJob)
		r.Put("/{id}", jobController.UpdateJob)
		r.Delete("/{id}", jobController.DeleteJob)
	})

	return r
}
