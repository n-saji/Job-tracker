package dto

type CreateAgentSessionRequest struct {
	JobID string `json:"job_id"`
}

type CreateAgentSessionResponse struct {
	ApplicationID string `json:"application_id"`
	SessionToken  string `json:"session_token"`
	ExpiresAt     string `json:"expires_at"`
}

type ApplicationResponse struct {
	ID             string  `json:"id"`
	JobID          string  `json:"job_id"`
	Status         string  `json:"status"`
	Mode           string  `json:"mode"`
	Error          *string `json:"error"`
	ScreenshotPath *string `json:"screenshot_path"`
	CompanyName    string  `json:"company_name"`
	RoleTitle      string  `json:"role_title"`
}

type ApplicationEventResponse struct {
	EventType string `json:"event_type"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}
