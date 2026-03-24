package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"job_tracker_be/internal/dao"
	"job_tracker_be/internal/dto"
	"job_tracker_be/internal/globals"
	"job_tracker_be/internal/service"

	"github.com/go-chi/chi/v5"
)

type JobController struct {
	service        *service.JobService
	requestTimeout time.Duration
}

func NewJobController(svc *service.JobService, requestTimeout time.Duration) *JobController {
	return &JobController{service: svc, requestTimeout: requestTimeout}
}

func (c *JobController) CreateJob(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	var req dto.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, globals.CodeBadRequest, "invalid request payload")
		return
	}

	job, err := c.service.Create(ctx, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, mapJob(job))
}

func (c *JobController) GetJob(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	id := chi.URLParam(r, "id")
	job, err := c.service.GetByID(ctx, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapJob(job))
}

func (c *JobController) ListJobs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	page := parseIntQuery(r, "page", globals.DefaultPage)
	limit := parseIntQuery(r, "limit", globals.DefaultLimit)
	status := r.URL.Query().Get("status")
	company := r.URL.Query().Get("company")

	jobs, total, normalizedPage, normalizedLimit, err := c.service.List(ctx, page, limit, status, company)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	responses := make([]dto.JobResponse, 0, len(jobs))
	for i := range jobs {
		job := jobs[i]
		responses = append(responses, mapJob(&job))
	}

	writeJSON(w, http.StatusOK, dto.ListJobsResponse{
		Data:  responses,
		Page:  normalizedPage,
		Limit: normalizedLimit,
		Total: total,
	})
}

func (c *JobController) UpdateJob(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	id := chi.URLParam(r, "id")

	var req dto.UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, globals.CodeBadRequest, "invalid request payload")
		return
	}

	job, err := c.service.Update(ctx, id, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapJob(job))
}

func (c *JobController) DeleteJob(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	id := chi.URLParam(r, "id")
	if err := c.service.Delete(ctx, id); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *JobController) ExistsByApplyLink(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	applyLink := r.URL.Query().Get("apply_link")
	exists, err := c.service.ExistsByApplyLink(ctx, applyLink)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ExistsApplyLinkResponse{Exists: exists})
}

func mapJob(job *dao.Job) dto.JobResponse {
	return dto.JobResponse{
		ID:             job.ID.String(),
		CompanyName:    job.CompanyName,
		RoleTitle:      job.RoleTitle,
		Location:       job.Location,
		ApplyLink:      job.ApplyLink,
		LinkedInJobURL: job.LinkedInJobURL,
		ResumeLink:     job.ResumeLink,
		Status:         job.Status,
		SalaryText:     job.SalaryText,
		IsEasyApply:    job.IsEasyApply,
		AppliedAt:      job.AppliedAt,
		CreatedAt:      job.CreatedAt,
		UpdatedAt:      job.UpdatedAt,
		DeletedAt:      job.DeletedAt,
	}
}

func parseIntQuery(r *http.Request, key string, fallback int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, globals.ErrBadRequest):
		writeError(w, http.StatusBadRequest, globals.CodeBadRequest, err.Error())
	case errors.Is(err, globals.ErrNotFound):
		writeError(w, http.StatusNotFound, globals.CodeNotFound, err.Error())
	case errors.Is(err, globals.ErrConflict):
		writeError(w, http.StatusConflict, globals.CodeConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, globals.CodeInternal, "internal server error")
	}
}
