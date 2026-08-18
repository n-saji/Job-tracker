package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"job_tracker_be/internal/dto"
	"job_tracker_be/internal/globals"
	"job_tracker_be/internal/service"
)

type AuthController struct {
	service        *service.AuthService
	requestTimeout time.Duration
}

func NewAuthController(svc *service.AuthService, requestTimeout time.Duration) *AuthController {
	return &AuthController{service: svc, requestTimeout: requestTimeout}
}

func (c *AuthController) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	resp, err := c.service.GetStatus(ctx)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (c *AuthController) Preflight(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	var req dto.AuthPreflightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, globals.CodeBadRequest, "invalid request payload")
		return
	}

	resp, err := c.service.Preflight(ctx, req.JobIDs)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
