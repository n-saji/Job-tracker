package service

import (
	"context"
	"net/http"

	"job_tracker_be/internal/dto"
)

// ApplicationService proxies application lifecycle operations to the Python
// AI agent service, which owns the applications/application_events tables
// and the Redis queue + worker pool. Go never talks to Playwright directly
// or writes to those tables — it's a thin, typed HTTP client.
type ApplicationService struct {
	agentHTTPClient
}

func NewApplicationService(baseURL string) *ApplicationService {
	return &ApplicationService{agentHTTPClient: newAgentHTTPClient(baseURL)}
}

func (s *ApplicationService) CreateApplication(ctx context.Context, jobID, mode string) (dto.ApplicationResponse, error) {
	var resp dto.ApplicationResponse
	err := s.doJSON(ctx, http.MethodPost, "/api/v1/applications", dto.CreateApplicationRequest{JobID: jobID, Mode: mode}, &resp)
	return resp, err
}

func (s *ApplicationService) BulkCreateApplications(ctx context.Context, jobIDs []string, mode string) (dto.BulkCreateApplicationsResponse, error) {
	var resp dto.BulkCreateApplicationsResponse
	err := s.doJSON(ctx, http.MethodPost, "/api/v1/applications/bulk", dto.BulkCreateApplicationsRequest{JobIDs: jobIDs, Mode: mode}, &resp)
	return resp, err
}

func (s *ApplicationService) ListApplications(ctx context.Context) ([]dto.ApplicationResponse, error) {
	var resp []dto.ApplicationResponse
	err := s.doJSON(ctx, http.MethodGet, "/api/v1/applications", nil, &resp)
	return resp, err
}

func (s *ApplicationService) GetApplication(ctx context.Context, id string) (dto.ApplicationResponse, error) {
	var resp dto.ApplicationResponse
	err := s.doJSON(ctx, http.MethodGet, "/api/v1/applications/"+id, nil, &resp)
	return resp, err
}

func (s *ApplicationService) CancelApplication(ctx context.Context, id string) (dto.ApplicationResponse, error) {
	var resp dto.ApplicationResponse
	err := s.doJSON(ctx, http.MethodPost, "/api/v1/applications/"+id+"/cancel", nil, &resp)
	return resp, err
}

func (s *ApplicationService) GetApplicationEvents(ctx context.Context, id string) ([]dto.ApplicationEventResponse, error) {
	var resp []dto.ApplicationEventResponse
	err := s.doJSON(ctx, http.MethodGet, "/api/v1/applications/"+id+"/events", nil, &resp)
	return resp, err
}
