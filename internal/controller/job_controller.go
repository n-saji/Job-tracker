package controller

import (
	"encoding/json"
	"errors"
	"fmt"
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
	resumeQueueSvc *service.ResumeQueueService
	broker         *JobEventBroker
	requestTimeout time.Duration
}

func NewJobController(svc *service.JobService, resumeQueueSvc *service.ResumeQueueService, broker *JobEventBroker, requestTimeout time.Duration) *JobController {
	return &JobController{service: svc, resumeQueueSvc: resumeQueueSvc, broker: broker, requestTimeout: requestTimeout}
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

	if c.broker != nil {
		c.broker.Publish(jobCreatedEvent{Job: mapJob(job)})
	}

	writeJSON(w, http.StatusCreated, mapJob(job))
}

type jobCreatedEvent struct {
	Job dto.JobResponse `json:"job"`
}

func (c *JobController) StreamCreatedJobs(w http.ResponseWriter, r *http.Request) {
	if c.broker == nil {
		writeError(w, http.StatusInternalServerError, globals.CodeInternal, "job event stream unavailable")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, globals.CodeInternal, "streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	events, unsubscribe := c.broker.Subscribe()
	defer unsubscribe()

	_, _ = w.Write([]byte(": connected\n\n"))
	flusher.Flush()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = w.Write([]byte(": ping\n\n"))
			flusher.Flush()
		case payload := <-events:
			raw, err := json.Marshal(payload)
			if err != nil {
				continue
			}

			if _, err := fmt.Fprintf(w, "event: job_created\ndata: %s\n\n", raw); err != nil {
				return
			}
			flusher.Flush()
		}
	}
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
	discardReason := r.URL.Query().Get("discard_reason")
	includeDiscarded := parseBoolQuery(r, "include_discarded", false)
	company := r.URL.Query().Get("company")
	location := r.URL.Query().Get("location")
	minMatchRating, err := parseOptionalFloatQuery(r, "min_match_rating")
	if err != nil {
		writeError(w, http.StatusBadRequest, globals.CodeBadRequest, "min_match_rating must be a valid number")
		return
	}
	maxMatchRating, err := parseOptionalFloatQuery(r, "max_match_rating")
	if err != nil {
		writeError(w, http.StatusBadRequest, globals.CodeBadRequest, "max_match_rating must be a valid number")
		return
	}
	sortMatch := r.URL.Query().Get("sort_match")

	jobs, total, normalizedPage, normalizedLimit, err := c.service.List(
		ctx,
		page,
		limit,
		status,
		discardReason,
		includeDiscarded,
		company,
		location,
		minMatchRating,
		maxMatchRating,
		sortMatch,
	)
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

func (c *JobController) BulkDeleteJobs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	var req dto.BulkDeleteJobsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, globals.CodeBadRequest, "invalid request payload")
		return
	}

	deletedCount, err := c.service.DeleteMany(ctx, req.IDs)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.BulkDeleteJobsResponse{DeletedCount: deletedCount})
}

func (c *JobController) BulkUpdateJobsStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	var req dto.BulkUpdateJobsStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, globals.CodeBadRequest, "invalid request payload")
		return
	}

	updatedCount, err := c.service.BulkUpdateStatus(ctx, req.IDs, req.Status, req.DiscardReason)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.BulkUpdateJobsStatusResponse{UpdatedCount: updatedCount})
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

func (c *JobController) TriggerResumeGenerate(w http.ResponseWriter, r *http.Request) {
	if c.resumeQueueSvc == nil {
		writeError(w, http.StatusInternalServerError, globals.CodeInternal, "resume queue service unavailable")
		return
	}

	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	id := chi.URLParam(r, "id")
	status, err := c.resumeQueueSvc.EnqueueJob(ctx, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusAccepted, dto.ResumeGenerateTriggerResponse{
		Status:  status,
		Message: "resume generation queued",
	})
}

func (c *JobController) ListResumeQueue(w http.ResponseWriter, r *http.Request) {
	if c.resumeQueueSvc == nil {
		writeError(w, http.StatusInternalServerError, globals.CodeInternal, "resume queue service unavailable")
		return
	}

	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	status := r.URL.Query().Get("status")
	limit := parseIntQuery(r, "limit", 100)

	items, err := c.resumeQueueSvc.ListByStatus(ctx, status, limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	response := make([]dto.ResumeQueueItemResponse, 0, len(items))
	for i := range items {
		item := items[i]
		response = append(response, dto.ResumeQueueItemResponse{
			JobID:     item.JobID.String(),
			ApplyLink: item.ApplyLink,
			Status:    item.Status,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, dto.ResumeQueueListResponse{Data: response})
}

func (c *JobController) UpdateResumeLink(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	id := chi.URLParam(r, "id")

	var req dto.UpdateResumeLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, globals.CodeBadRequest, "invalid request payload")
		return
	}

	resumeLink := req.ResumeLink
	job, err := c.service.Update(ctx, id, dto.UpdateJobRequest{ResumeLink: &resumeLink})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapJob(job))
}

func (c *JobController) DeleteResumeQueueItem(w http.ResponseWriter, r *http.Request) {
	if c.resumeQueueSvc == nil {
		writeError(w, http.StatusInternalServerError, globals.CodeInternal, "resume queue service unavailable")
		return
	}

	ctx, cancel := service.WithTimeout(r.Context(), c.requestTimeout)
	defer cancel()

	jobID := chi.URLParam(r, "job_id")
	if err := c.resumeQueueSvc.DeleteByJobID(ctx, jobID); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func mapJob(job *dao.Job) dto.JobResponse {
	return dto.JobResponse{
		ID:             job.ID.String(),
		CompanyName:    job.CompanyName,
		RoleTitle:      job.RoleTitle,
		Location:       job.Location,
		JobDescription: job.JobDescription,
		ApplyLink:      job.ApplyLink,
		LinkedInJobURL: job.LinkedInJobURL,
		ResumeLink:     job.ResumeLink,
		Status:         job.Status,
		DiscardReason:  job.DiscardReason,
		SalaryText:     job.SalaryText,
		IsEasyApply:    job.IsEasyApply,
		MatchRating:    job.MatchRating,
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

func parseBoolQuery(r *http.Request, key string, fallback bool) bool {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func parseOptionalFloatQuery(r *http.Request, key string) (*float64, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
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
