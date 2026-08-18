package dto

type CreateApplicationRequest struct {
	JobID string `json:"job_id"`
	Mode  string `json:"mode"`
}

type BulkCreateApplicationsRequest struct {
	JobIDs []string `json:"job_ids"`
	Mode   string   `json:"mode"`
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

type BulkCreateApplicationsResponse struct {
	Created        int      `json:"created"`
	ApplicationIDs []string `json:"application_ids"`
}

type ApplicationEventResponse struct {
	EventType string `json:"event_type"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}
