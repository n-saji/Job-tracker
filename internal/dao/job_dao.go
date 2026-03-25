package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Job struct {
	ID             uuid.UUID
	CompanyName    string
	RoleTitle      string
	Location       string
	ApplyLink      string
	LinkedInJobURL string
	ResumeLink     string
	Status         string
	SalaryText     string
	IsEasyApply    bool
	AppliedAt      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

type CreateJobParams struct {
	CompanyName    string
	RoleTitle      string
	Location       string
	ApplyLink      string
	LinkedInJobURL string
	ResumeLink     string
	Status         string
	SalaryText     string
	IsEasyApply    bool
	AppliedAt      time.Time
}

type UpdateJobParams struct {
	CompanyName    *string
	RoleTitle      *string
	Location       *string
	ApplyLink      *string
	LinkedInJobURL *string
	ResumeLink     *string
	Status         *string
	SalaryText     *string
	IsEasyApply    *bool
	AppliedAt      *time.Time
}

type ListJobsParams struct {
	Page     int
	Limit    int
	Status   string
	Company  string
	Location string
}

type JobDAO interface {
	Create(ctx context.Context, params CreateJobParams) (*Job, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Job, error)
	List(ctx context.Context, params ListJobsParams) ([]Job, int64, error)
	Update(ctx context.Context, id uuid.UUID, params UpdateJobParams) (*Job, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ExistsByApplyLink(ctx context.Context, applyLink string) (bool, error)
}

type PgxJobDAO struct {
	pool *pgxpool.Pool
}

func NewPgxJobDAO(pool *pgxpool.Pool) *PgxJobDAO {
	return &PgxJobDAO{pool: pool}
}

func (d *PgxJobDAO) Create(ctx context.Context, params CreateJobParams) (*Job, error) {
	query := `
		INSERT INTO jobs (
company_name, role_title, location, apply_link, linkedin_job_url,
resume_link, status, salary_text, is_easy_apply, applied_at
)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, company_name, role_title, location, apply_link, linkedin_job_url,
			resume_link, status, salary_text, is_easy_apply, applied_at, created_at, updated_at, deleted_at
	`

	job, err := scanJob(d.pool.QueryRow(ctx, query,
		params.CompanyName,
		params.RoleTitle,
		params.Location,
		params.ApplyLink,
		params.LinkedInJobURL,
		params.ResumeLink,
		params.Status,
		params.SalaryText,
		params.IsEasyApply,
		params.AppliedAt,
	))
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("duplicate apply_link: %w", err)
		}
		return nil, err
	}

	return job, nil
}

func (d *PgxJobDAO) GetByID(ctx context.Context, id uuid.UUID) (*Job, error) {
	query := `
		SELECT id, company_name, role_title, location, apply_link, linkedin_job_url,
			resume_link, status, salary_text, is_easy_apply, applied_at, created_at, updated_at, deleted_at
		FROM jobs
		WHERE id = $1 AND deleted_at IS NULL
	`

	job, err := scanJob(d.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return job, nil
}

func (d *PgxJobDAO) List(ctx context.Context, params ListJobsParams) ([]Job, int64, error) {
	baseWhere := "WHERE deleted_at IS NULL"
	args := make([]any, 0, 5)
	argPos := 1

	if params.Status != "" {
		baseWhere += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, params.Status)
		argPos++
	}
	if params.Company != "" {
		baseWhere += fmt.Sprintf(" AND company_name ILIKE $%d", argPos)
		args = append(args, "%"+params.Company+"%")
		argPos++
	}
	if params.Location != "" {
		baseWhere += fmt.Sprintf(" AND location ILIKE $%d", argPos)
		args = append(args, "%"+params.Location+"%")
		argPos++
	}

	countQuery := "SELECT COUNT(1) FROM jobs " + baseWhere
	var total int64
	if err := d.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	listQuery := "\n\t\tSELECT id, company_name, role_title, location, apply_link, linkedin_job_url,\n\t\t\tresume_link, status, salary_text, is_easy_apply, applied_at, created_at, updated_at, deleted_at\n\t\tFROM jobs " + baseWhere +
		fmt.Sprintf(" ORDER BY updated_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)

	listArgs := append(args, params.Limit, offset)
	rows, err := d.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	jobs := make([]Job, 0, params.Limit)
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, 0, err
		}
		jobs = append(jobs, *job)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
}

func (d *PgxJobDAO) Update(ctx context.Context, id uuid.UUID, params UpdateJobParams) (*Job, error) {
	setClauses := make([]string, 0, 10)
	args := make([]any, 0, 12)
	argPos := 1

	if params.CompanyName != nil {
		setClauses = append(setClauses, fmt.Sprintf("company_name = $%d", argPos))
		args = append(args, *params.CompanyName)
		argPos++
	}
	if params.RoleTitle != nil {
		setClauses = append(setClauses, fmt.Sprintf("role_title = $%d", argPos))
		args = append(args, *params.RoleTitle)
		argPos++
	}
	if params.Location != nil {
		setClauses = append(setClauses, fmt.Sprintf("location = $%d", argPos))
		args = append(args, *params.Location)
		argPos++
	}
	if params.ApplyLink != nil {
		setClauses = append(setClauses, fmt.Sprintf("apply_link = $%d", argPos))
		args = append(args, *params.ApplyLink)
		argPos++
	}
	if params.LinkedInJobURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("linkedin_job_url = $%d", argPos))
		args = append(args, *params.LinkedInJobURL)
		argPos++
	}
	if params.ResumeLink != nil {
		setClauses = append(setClauses, fmt.Sprintf("resume_link = $%d", argPos))
		args = append(args, *params.ResumeLink)
		argPos++
	}
	if params.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *params.Status)
		argPos++
	}
	if params.SalaryText != nil {
		setClauses = append(setClauses, fmt.Sprintf("salary_text = $%d", argPos))
		args = append(args, *params.SalaryText)
		argPos++
	}
	if params.IsEasyApply != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_easy_apply = $%d", argPos))
		args = append(args, *params.IsEasyApply)
		argPos++
	}
	if params.AppliedAt != nil {
		setClauses = append(setClauses, fmt.Sprintf("applied_at = $%d", argPos))
		args = append(args, *params.AppliedAt)
		argPos++
	}

	if len(setClauses) == 0 {
		return d.GetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query := fmt.Sprintf(`
		UPDATE jobs
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING id, company_name, role_title, location, apply_link, linkedin_job_url,
			resume_link, status, salary_text, is_easy_apply, applied_at, created_at, updated_at, deleted_at
	`, strings.Join(setClauses, ", "), argPos)

	args = append(args, id)
	job, err := scanJob(d.pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("duplicate apply_link: %w", err)
		}
		return nil, err
	}

	return job, nil
}

func (d *PgxJobDAO) SoftDelete(ctx context.Context, id uuid.UUID) error {
	commandTag, err := d.pool.Exec(ctx, `
		UPDATE jobs
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (d *PgxJobDAO) ExistsByApplyLink(ctx context.Context, applyLink string) (bool, error) {
	var exists bool
	if err := d.pool.QueryRow(ctx, `
		SELECT EXISTS(
SELECT 1
FROM jobs
WHERE apply_link = $1 AND deleted_at IS NULL
)
	`, applyLink).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanJob(row rowScanner) (*Job, error) {
	var job Job
	if err := row.Scan(
		&job.ID,
		&job.CompanyName,
		&job.RoleTitle,
		&job.Location,
		&job.ApplyLink,
		&job.LinkedInJobURL,
		&job.ResumeLink,
		&job.Status,
		&job.SalaryText,
		&job.IsEasyApply,
		&job.AppliedAt,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.DeletedAt,
	); err != nil {
		return nil, err
	}
	return &job, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
