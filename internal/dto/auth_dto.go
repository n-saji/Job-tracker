package dto

type JobBoardAccountResponse struct {
	JobBoard            string  `json:"job_board"`
	Status              string  `json:"status"`
	LastAuthenticatedAt *string `json:"last_authenticated_at"`
	LastVerifiedAt      *string `json:"last_verified_at"`
}

type AuthPreflightRequest struct {
	JobIDs []string `json:"job_ids"`
}

type AuthPreflightJobResult struct {
	JobID    string `json:"job_id"`
	JobBoard string `json:"job_board"`
	Status   string `json:"status"`
}

type AuthPreflightResponse struct {
	JobBoards map[string]string        `json:"job_boards"`
	Jobs      []AuthPreflightJobResult `json:"jobs"`
}
