package dto

import (
	"bytes"
	"encoding/json"
	"time"
)

type OptionalFloat64 struct {
	Set   bool
	Value *float64
}

func (o *OptionalFloat64) UnmarshalJSON(data []byte) error {
	o.Set = true
	if bytes.Equal(data, []byte("null")) {
		o.Value = nil
		return nil
	}

	var value float64
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	o.Value = &value
	return nil
}

type CreateJobRequest struct {
	CompanyName    string    `json:"company_name"`
	RoleTitle      string    `json:"role_title"`
	Location       string    `json:"location"`
	JobDescription string    `json:"job_description"`
	ApplyLink      string    `json:"apply_link"`
	LinkedInJobURL string    `json:"linkedin_job_url"`
	ResumeLink     string    `json:"resume_link"`
	Status         string    `json:"status"`
	DiscardReason  string    `json:"discard_reason"`
	SalaryText     string    `json:"salary_text"`
	IsEasyApply    string    `json:"is_easy_apply"`
	MatchRating    *float64  `json:"match_rating"`
	AppliedAt      time.Time `json:"applied_at"`
}

type UpdateJobRequest struct {
	CompanyName    *string         `json:"company_name"`
	RoleTitle      *string         `json:"role_title"`
	Location       *string         `json:"location"`
	JobDescription *string         `json:"job_description"`
	ApplyLink      *string         `json:"apply_link"`
	LinkedInJobURL *string         `json:"linkedin_job_url"`
	ResumeLink     *string         `json:"resume_link"`
	Status         *string         `json:"status"`
	DiscardReason  *string         `json:"discard_reason"`
	SalaryText     *string         `json:"salary_text"`
	IsEasyApply    *bool           `json:"is_easy_apply"`
	MatchRating    OptionalFloat64 `json:"match_rating"`
	AppliedAt      *time.Time      `json:"applied_at"`
}

type BulkDeleteJobsRequest struct {
	IDs []string `json:"ids"`
}

type BulkUpdateJobsStatusRequest struct {
	IDs           []string `json:"ids"`
	Status        string   `json:"status"`
	DiscardReason string   `json:"discard_reason"`
}

type JobResponse struct {
	ID             string     `json:"id"`
	CompanyName    string     `json:"company_name"`
	RoleTitle      string     `json:"role_title"`
	Location       string     `json:"location"`
	JobDescription string     `json:"job_description"`
	ApplyLink      string     `json:"apply_link"`
	LinkedInJobURL string     `json:"linkedin_job_url"`
	ResumeLink     string     `json:"resume_link"`
	Status         string     `json:"status"`
	DiscardReason  *string    `json:"discard_reason,omitempty"`
	SalaryText     string     `json:"salary_text"`
	IsEasyApply    bool       `json:"is_easy_apply"`
	MatchRating    *float64   `json:"match_rating,omitempty"`
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

type BulkUpdateJobsStatusResponse struct {
	UpdatedCount int `json:"updated_count"`
}

type ResumeGenerateTriggerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ResumeQueueItemResponse struct {
	JobID     string    `json:"job_id"`
	ApplyLink string    `json:"apply_link"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ResumeQueueListResponse struct {
	Data []ResumeQueueItemResponse `json:"data"`
}

type UpdateResumeLinkRequest struct {
	ResumeLink string `json:"resume_link"`
}
