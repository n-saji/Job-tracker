package service

import (
	"context"
	"net/http"

	"job_tracker_be/internal/dto"
)

// AuthService proxies job-board authentication *status* to the Python AI
// agent service — same ownership split as ApplicationService: Go never
// touches session files, credentials, or the job_board_accounts table
// directly. Actually authenticating a job board is a local, human-assisted
// CLI action (`python -m agent auth <board> --headed --url ...`) — there's
// no way to pop a real browser window through a web request, so this is
// read-only by design.
type AuthService struct {
	agentHTTPClient
}

func NewAuthService(baseURL string) *AuthService {
	return &AuthService{agentHTTPClient: newAgentHTTPClient(baseURL)}
}

func (s *AuthService) GetStatus(ctx context.Context) ([]dto.JobBoardAccountResponse, error) {
	var resp []dto.JobBoardAccountResponse
	err := s.doJSON(ctx, http.MethodGet, "/api/v1/auth/status", nil, &resp)
	return resp, err
}

func (s *AuthService) Preflight(ctx context.Context, jobIDs []string) (dto.AuthPreflightResponse, error) {
	var resp dto.AuthPreflightResponse
	err := s.doJSON(ctx, http.MethodPost, "/api/v1/auth/preflight", dto.AuthPreflightRequest{JobIDs: jobIDs}, &resp)
	return resp, err
}
