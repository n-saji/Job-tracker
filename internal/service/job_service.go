package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"job_tracker_be/internal/dao"
	"job_tracker_be/internal/dto"
	"job_tracker_be/internal/globals"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type JobService struct {
	dao dao.JobDAO
}

func NewJobService(jobDAO dao.JobDAO) *JobService {
	return &JobService{dao: jobDAO}
}

func (s *JobService) Create(ctx context.Context, req dto.CreateJobRequest) (*dao.Job, error) {
	if err := validateCreate(req); err != nil {
		return nil, err
	}

	normalizedApplyLink := normalizeApplyLink(req.ApplyLink)
	if normalizedApplyLink == "" {
		return nil, fmt.Errorf("apply_link is required: %w", globals.ErrBadRequest)
	}

	job, err := s.dao.Create(ctx, dao.CreateJobParams{
		CompanyName:    strings.TrimSpace(req.CompanyName),
		RoleTitle:      strings.TrimSpace(req.RoleTitle),
		Location:       strings.TrimSpace(req.Location),
		ApplyLink:      normalizedApplyLink,
		LinkedInJobURL: strings.TrimSpace(req.LinkedInJobURL),
		ResumeLink:     strings.TrimSpace(req.ResumeLink),
		Status:         strings.ToLower(strings.TrimSpace(req.Status)),
		DiscardReason:  normalizeOptionalDiscardReason(req.DiscardReason),
		SalaryText:     strings.TrimSpace(req.SalaryText),
		IsEasyApply:    bool(strings.EqualFold(req.IsEasyApply, "true")),
		AppliedAt:      req.AppliedAt,
	})
	if err != nil {
		if isUniqueError(err) {
			return nil, fmt.Errorf("job with apply_link already exists: %w", globals.ErrConflict)
		}
		return nil, err
	}
	return job, nil
}

func (s *JobService) GetByID(ctx context.Context, id string) (*dao.Job, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", globals.ErrBadRequest)
	}

	job, err := s.dao.GetByID(ctx, parsedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, globals.ErrNotFound
		}
		return nil, err
	}

	return job, nil
}

func (s *JobService) List(ctx context.Context, page, limit int, status, discardReason string, includeDiscarded bool, company, location string) ([]dao.Job, int64, int, int, error) {
	if page <= 0 {
		page = globals.DefaultPage
	}
	if limit <= 0 {
		limit = globals.DefaultLimit
	}
	if limit > globals.MaxLimit {
		limit = globals.MaxLimit
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" {
		if _, ok := globals.AllowedStatuses[status]; !ok {
			return nil, 0, 0, 0, fmt.Errorf("invalid status: %w", globals.ErrBadRequest)
		}
	}

	discardReason = strings.ToLower(strings.TrimSpace(discardReason))
	if discardReason != "" {
		if _, ok := globals.AllowedDiscardReasons[discardReason]; !ok {
			return nil, 0, 0, 0, fmt.Errorf("invalid discard_reason: %w", globals.ErrBadRequest)
		}
		if status == "" {
			status = globals.StatusDiscarded
		}
		if status != globals.StatusDiscarded {
			return nil, 0, 0, 0, fmt.Errorf("discard_reason filter is only allowed with discarded status: %w", globals.ErrBadRequest)
		}
	}

	jobs, total, err := s.dao.List(ctx, dao.ListJobsParams{
		Page:          page,
		Limit:         limit,
		Status:        status,
		DiscardReason: discardReason,
		IncludeDiscarded: includeDiscarded,
		Company:       strings.TrimSpace(company),
		Location:      strings.TrimSpace(location),
	})
	if err != nil {
		return nil, 0, 0, 0, err
	}

	return jobs, total, page, limit, nil
}

func (s *JobService) Update(ctx context.Context, id string, req dto.UpdateJobRequest) (*dao.Job, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", globals.ErrBadRequest)
	}

	current, err := s.dao.GetByID(ctx, parsedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, globals.ErrNotFound
		}
		return nil, err
	}

	params, err := validateAndBuildUpdate(req, current)
	if err != nil {
		return nil, err
	}

	job, err := s.dao.Update(ctx, parsedID, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, globals.ErrNotFound
		}
		if isUniqueError(err) {
			return nil, fmt.Errorf("job with apply_link already exists: %w", globals.ErrConflict)
		}
		return nil, err
	}

	return job, nil
}

func (s *JobService) Delete(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid id: %w", globals.ErrBadRequest)
	}

	if err := s.dao.SoftDelete(ctx, parsedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return globals.ErrNotFound
		}
		return err
	}

	return nil
}

func (s *JobService) DeleteMany(ctx context.Context, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("ids is required: %w", globals.ErrBadRequest)
	}

	parsedIDs := make([]uuid.UUID, 0, len(ids))
	seen := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			return 0, fmt.Errorf("id cannot be empty: %w", globals.ErrBadRequest)
		}

		parsedID, err := uuid.Parse(trimmed)
		if err != nil {
			return 0, fmt.Errorf("invalid id: %w", globals.ErrBadRequest)
		}

		if _, exists := seen[parsedID]; exists {
			continue
		}
		seen[parsedID] = struct{}{}
		parsedIDs = append(parsedIDs, parsedID)
	}

	deletedCount, err := s.dao.SoftDeleteMany(ctx, parsedIDs)
	if err != nil {
		return 0, err
	}

	return int(deletedCount), nil
}

func (s *JobService) ExistsByApplyLink(ctx context.Context, applyLink string) (bool, error) {
	normalized := normalizeApplyLink(applyLink)
	if normalized == "" {
		return false, fmt.Errorf("apply_link is required: %w", globals.ErrBadRequest)
	}
	return s.dao.ExistsByApplyLink(ctx, normalized)
}

func validateCreate(req dto.CreateJobRequest) error {
	if strings.TrimSpace(req.CompanyName) == "" {
		return fmt.Errorf("company_name is required: %w", globals.ErrBadRequest)
	}
	if strings.TrimSpace(req.RoleTitle) == "" {
		return fmt.Errorf("role_title is required: %w", globals.ErrBadRequest)
	}
	if strings.TrimSpace(req.Location) == "" {
		return fmt.Errorf("location is required: %w", globals.ErrBadRequest)
	}
	if strings.TrimSpace(req.ApplyLink) == "" {
		return fmt.Errorf("apply_link is required: %w", globals.ErrBadRequest)
	}
	status := strings.ToLower(strings.TrimSpace(req.Status))
	if _, ok := globals.AllowedStatuses[status]; !ok {
		return fmt.Errorf("invalid status: %w", globals.ErrBadRequest)
	}

	discardReason := strings.ToLower(strings.TrimSpace(req.DiscardReason))
	if status == globals.StatusDiscarded {
		if discardReason == "" {
			return fmt.Errorf("discard_reason is required when status is discarded: %w", globals.ErrBadRequest)
		}
		if _, ok := globals.AllowedDiscardReasons[discardReason]; !ok {
			return fmt.Errorf("invalid discard_reason: %w", globals.ErrBadRequest)
		}
	} else if discardReason != "" {
		return fmt.Errorf("discard_reason is only allowed when status is discarded: %w", globals.ErrBadRequest)
	}

	if req.AppliedAt.IsZero() {
		req.AppliedAt = time.Now()
	}
	return nil
}

func validateAndBuildUpdate(req dto.UpdateJobRequest, current *dao.Job) (dao.UpdateJobParams, error) {
	params := dao.UpdateJobParams{}
	provided := false
	targetStatus := current.Status

	if req.CompanyName != nil {
		provided = true
		value := strings.TrimSpace(*req.CompanyName)
		if value == "" {
			return params, fmt.Errorf("company_name cannot be empty: %w", globals.ErrBadRequest)
		}
		params.CompanyName = &value
	}
	if req.RoleTitle != nil {
		provided = true
		value := strings.TrimSpace(*req.RoleTitle)
		if value == "" {
			return params, fmt.Errorf("role_title cannot be empty: %w", globals.ErrBadRequest)
		}
		params.RoleTitle = &value
	}
	if req.Location != nil {
		provided = true
		value := strings.TrimSpace(*req.Location)
		if value == "" {
			return params, fmt.Errorf("location cannot be empty: %w", globals.ErrBadRequest)
		}
		params.Location = &value
	}
	if req.ApplyLink != nil {
		provided = true
		value := normalizeApplyLink(*req.ApplyLink)
		if value == "" {
			return params, fmt.Errorf("apply_link cannot be empty: %w", globals.ErrBadRequest)
		}
		params.ApplyLink = &value
	}
	if req.LinkedInJobURL != nil {
		provided = true
		value := strings.TrimSpace(*req.LinkedInJobURL)
		params.LinkedInJobURL = &value
	}
	if req.ResumeLink != nil {
		provided = true
		value := strings.TrimSpace(*req.ResumeLink)
		params.ResumeLink = &value
	}
	if req.Status != nil {
		provided = true
		value := strings.ToLower(strings.TrimSpace(*req.Status))
		if _, ok := globals.AllowedStatuses[value]; !ok {
			return params, fmt.Errorf("invalid status: %w", globals.ErrBadRequest)
		}
		params.Status = &value
		targetStatus = value
	}
	if req.DiscardReason != nil {
		provided = true
		value := strings.ToLower(strings.TrimSpace(*req.DiscardReason))
		if value == "" {
			params.ClearDiscardReason = true
		} else {
			if _, ok := globals.AllowedDiscardReasons[value]; !ok {
				return params, fmt.Errorf("invalid discard_reason: %w", globals.ErrBadRequest)
			}
			params.DiscardReason = &value
		}
	}
	if req.SalaryText != nil {
		provided = true
		value := strings.TrimSpace(*req.SalaryText)
		params.SalaryText = &value
	}
	if req.IsEasyApply != nil {
		provided = true
		params.IsEasyApply = req.IsEasyApply
	}
	if req.AppliedAt != nil {
		provided = true
		if req.AppliedAt.IsZero() {
			return params, fmt.Errorf("applied_at cannot be zero: %w", globals.ErrBadRequest)
		}
		params.AppliedAt = req.AppliedAt
	}

	if !provided {
		return params, fmt.Errorf("no fields to update: %w", globals.ErrBadRequest)
	}

	if targetStatus == globals.StatusDiscarded {
		hasCurrentDiscardReason := current.DiscardReason != nil && strings.TrimSpace(*current.DiscardReason) != ""
		hasTargetDiscardReason := params.DiscardReason != nil
		if params.ClearDiscardReason || (!hasTargetDiscardReason && !hasCurrentDiscardReason) {
			return params, fmt.Errorf("discard_reason is required when status is discarded: %w", globals.ErrBadRequest)
		}
	} else {
		if params.DiscardReason != nil {
			return params, fmt.Errorf("discard_reason is only allowed when status is discarded: %w", globals.ErrBadRequest)
		}

		if current.DiscardReason != nil || req.DiscardReason != nil || req.Status != nil {
			params.ClearDiscardReason = true
		}
	}

	return params, nil
}

func normalizeOptionalDiscardReason(raw string) *string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return nil
	}
	return &value
}

func normalizeApplyLink(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""
	if parsed.Path != "/" {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	}

	return parsed.String()
}

func isUniqueError(err error) bool {
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "duplicate") || strings.Contains(lower, "23505")
}

func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
