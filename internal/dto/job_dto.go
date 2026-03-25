package dto

import "time"

type CreateJobRequest struct {
	CompanyName    string    `json:"company_name"`
	RoleTitle      string    `json:"role_title"`
	Location       string    `json:"location"`
	ApplyLink      string    `json:"apply_link"`
	LinkedInJobURL string    `json:"linkedin_job_url"`
	ResumeLink     string    `json:"resume_link"`
	Status         string    `json:"status"`
	DiscardReason  string    `json:"discard_reason"`
	SalaryText     string    `json:"salary_text"`
	IsEasyApply    string    `json:"is_easy_apply"`
	AppliedAt      time.Time `json:"applied_at"`
}

type UpdateJobRequest struct {
	CompanyName    *string    `json:"company_name"`
	RoleTitle      *string    `json:"role_title"`
	Location       *string    `json:"location"`
	ApplyLink      *string    `json:"apply_link"`
	LinkedInJobURL *string    `json:"linkedin_job_url"`
	ResumeLink     *string    `json:"resume_link"`
	Status         *string    `json:"status"`
	DiscardReason  *string    `json:"discard_reason"`
	SalaryText     *string    `json:"salary_text"`
	IsEasyApply    *bool      `json:"is_easy_apply"`
	AppliedAt      *time.Time `json:"applied_at"`
}

type BulkDeleteJobsRequest struct {
	IDs []string `json:"ids"`
}

type JobResponse struct {
	ID             string     `json:"id"`
	CompanyName    string     `json:"company_name"`
	RoleTitle      string     `json:"role_title"`
	Location       string     `json:"location"`
	ApplyLink      string     `json:"apply_link"`
	LinkedInJobURL string     `json:"linkedin_job_url"`
	ResumeLink     string     `json:"resume_link"`
	Status         string     `json:"status"`
	DiscardReason  *string    `json:"discard_reason,omitempty"`
	SalaryText     string     `json:"salary_text"`
	IsEasyApply    bool       `json:"is_easy_apply"`
	AppliedAt      time.Time  `json:"applied_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

type ListJobsResponse struct {
	Data  []JobResponse `json:"data"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
	Total int64         `json:"total"`
}

type ExistsApplyLinkResponse struct {
	Exists bool `json:"exists"`
}

type BulkDeleteJobsResponse struct {
	DeletedCount int `json:"deleted_count"`
}
