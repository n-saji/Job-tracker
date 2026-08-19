package service

import (
	"context"
	"net/http"

	"job_tracker_be/internal/dto"
)

// ApplicationService proxies application lifecycle operations to the Python
// AI agent service, which owns the applications/application_events/
// agent_sessions tables. Go never talks to a browser directly or writes to
// those tables — it's a thin, typed HTTP client.
//
// Callers: job_tracker_extension's popup calls CreateAgentSession directly
// (via this same Go endpoint, not job_tracker_fe) after the user picks a job
// to apply to — Go stays the single source of truth for job data either way,
// so the extension doesn't duplicate job listing/filtering logic. Once a
// session exists, the extension's content script calls the Python agent's
// resolve/fields/outcome endpoints directly from the browser, bypassing Go
// entirely — those are chatty, DOM-round-trip-heavy calls with no reason to
// hop through an extra service. job_tracker_fe still uses ListApplications/
// GetApplication/CancelApplication/GetApplicationEvents for its read-only
// Applications status page.
type ApplicationService struct {
	agentHTTPClient
}

func NewApplicationService(baseURL string) *ApplicationService {
	return &ApplicationService{agentHTTPClient: newAgentHTTPClient(baseURL)}
}

func (s *ApplicationService) CreateAgentSession(ctx context.Context, jobID string) (dto.CreateAgentSessionResponse, error) {
	var resp dto.CreateAgentSessionResponse
	err := s.doJSON(ctx, http.MethodPost, "/api/v1/agent-sessions", dto.CreateAgentSessionRequest{JobID: jobID}, &resp)
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
