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
		r.Get("/", jobController.ListJobs)
		r.Get("/exists", jobController.ExistsByApplyLink)
		r.Get("/{id}", jobController.GetJob)
		r.Put("/{id}", jobController.UpdateJob)
		r.Delete("/{id}", jobController.DeleteJob)
	})

	return r
}
