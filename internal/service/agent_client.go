package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"job_tracker_be/internal/globals"
)

// agentHTTPClient is the shared typed-HTTP-proxy plumbing for every service
// that talks to the Python AI agent service (currently just
// ApplicationService) — request building, response decoding, and mapping the
// agent's HTTP status codes onto Go's existing globals.Err* sentinels so
// every caller gets the same writeServiceError behavior.
type agentHTTPClient struct {
	baseURL string
	client  *http.Client
}

func newAgentHTTPClient(baseURL string) agentHTTPClient {
	return agentHTTPClient{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c agentHTTPClient) doJSON(ctx context.Context, method, path string, body any, out any) error {
	if c.baseURL == "" {
		return fmt.Errorf("python agent service not configured: %w", globals.ErrInternal)
	}

	var reqBody bytes.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		reqBody = *bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, &reqBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("agent service unreachable: %w: %w", globals.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		if out == nil {
			return nil
		}
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode agent response: %w", err)
		}
		return nil
	}

	detail := decodeErrorDetail(resp)
	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("%s: %w", detail, globals.ErrNotFound)
	case http.StatusConflict:
		return fmt.Errorf("%s: %w", detail, globals.ErrConflict)
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return fmt.Errorf("%s: %w", detail, globals.ErrBadRequest)
	default:
		return fmt.Errorf("agent service error (%d): %s: %w", resp.StatusCode, detail, globals.ErrUpstream)
	}
}

func decodeErrorDetail(resp *http.Response) string {
	var payload struct {
		Detail string `json:"detail"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil || payload.Detail == "" {
		return resp.Status
	}
	return payload.Detail
}
