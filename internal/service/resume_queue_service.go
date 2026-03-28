package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"job_tracker_be/internal/dao"
	"job_tracker_be/internal/globals"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	ResumeQueueStatusAdded = "added"
)

type ResumeQueueService struct {
	jobDAO     dao.JobDAO
	queueDAO   dao.ResumeQueueDAO
	webhookURL string
	client     *http.Client
}

func NewResumeQueueService(jobDAO dao.JobDAO, queueDAO dao.ResumeQueueDAO, webhookURL string) *ResumeQueueService {
	return &ResumeQueueService{
		jobDAO:     jobDAO,
		queueDAO:   queueDAO,
		webhookURL: strings.TrimSpace(webhookURL),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *ResumeQueueService) EnqueueJob(ctx context.Context, rawJobID string) (string, error) {
	jobID, err := uuid.Parse(strings.TrimSpace(rawJobID))
	if err != nil {
		return "", fmt.Errorf("invalid id: %w", globals.ErrBadRequest)
	}

	job, err := s.jobDAO.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", globals.ErrNotFound
		}
		return "", err
	}

	if strings.TrimSpace(job.JobDescription) == "" {
		return "", fmt.Errorf("job_description is required before generating resume: %w", globals.ErrBadRequest)
	}
	if strings.TrimSpace(job.ResumeLink) != "" {
		return "", fmt.Errorf("resume_link already exists for this job: %w", globals.ErrConflict)
	}

	if err := s.queueDAO.Enqueue(ctx, jobID, job.ApplyLink, ResumeQueueStatusAdded); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "apply_link already exists") {
			return "", fmt.Errorf("job with this apply_link is already queued: %w", globals.ErrConflict)
		}
		return "", err
	}

	go s.callWebhook(jobID)
	return "queued", nil
}

func (s *ResumeQueueService) ListByStatus(ctx context.Context, status string, limit int) ([]dao.ResumeQueueItem, error) {
	normalizedStatus := strings.ToLower(strings.TrimSpace(status))
	if normalizedStatus == "" {
		normalizedStatus = ResumeQueueStatusAdded
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	return s.queueDAO.ListByStatus(ctx, normalizedStatus, limit)
}

func (s *ResumeQueueService) DeleteByJobID(ctx context.Context, rawJobID string) error {
	jobID, err := uuid.Parse(strings.TrimSpace(rawJobID))
	if err != nil {
		return fmt.Errorf("invalid job_id: %w", globals.ErrBadRequest)
	}

	if err := s.queueDAO.DeleteByJobID(ctx, jobID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return globals.ErrNotFound
		}
		return err
	}
	return nil
}

func (s *ResumeQueueService) callWebhook(jobID uuid.UUID) {
	if s.webhookURL == "" {
		return
	}

	webhookURL, err := url.Parse(s.webhookURL)
	if err != nil {
		log.Printf("resume webhook parse error: %v", err)
		return
	}

	query := webhookURL.Query()
	query.Set("job_id", jobID.String())
	webhookURL.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, webhookURL.String(), nil)
	if err != nil {
		log.Printf("resume webhook request build error: %v", err)
		return
	}

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("resume webhook call failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		log.Printf("resume webhook non-success status: %d", resp.StatusCode)
	}
}
