package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"job_tracker_be/internal/dto"
	"job_tracker_be/internal/globals"
	"job_tracker_be/internal/service"

	"github.com/go-chi/chi/v5"
)

type ApplicationController struct {
	service        *service.ApplicationService
	requestTimeout time.Duration
}

func NewApplicationController(svc *service.ApplicationService, requestTimeout time.Duration) *ApplicationController {
	return &ApplicationController{service: svc, requestTimeout: requestTimeout}
}

func (c *ApplicationController) CreateAgentSession(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	var req dto.CreateAgentSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, globals.CodeBadRequest, "invalid request payload")
		return
	}

	resp, err := c.service.CreateAgentSession(ctx, req.JobID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (c *ApplicationController) ListApplications(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	resp, err := c.service.ListApplications(ctx)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (c *ApplicationController) GetApplication(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	id := chi.URLParam(r, "id")
	resp, err := c.service.GetApplication(ctx, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (c *ApplicationController) CancelApplication(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	id := chi.URLParam(r, "id")
	resp, err := c.service.CancelApplication(ctx, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (c *ApplicationController) GetApplicationEvents(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	id := chi.URLParam(r, "id")
	resp, err := c.service.GetApplicationEvents(ctx, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
