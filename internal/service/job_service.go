package service

import (
	"context"
	"encoding/json"
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

	status, verdict, discardReason, rejectReason, err := resolveJobStatusInputs(req.Status, req.DiscardReason, req.Verdict, req.RejectReason)
	if err != nil {
		return nil, err
	}

	sectionScores, extracted, flags, totalScore, err := marshalAssessmentPayload(req.SectionScores, req.Extracted, req.Flags, req.TotalScore)
	if err != nil {
		return nil, err
	}

	job, err := s.dao.Create(ctx, dao.CreateJobParams{
		CompanyName:    strings.TrimSpace(req.CompanyName),
		RoleTitle:      strings.TrimSpace(req.RoleTitle),
		Location:       strings.TrimSpace(req.Location),
		JobDescription: strings.TrimSpace(req.JobDescription),
		ApplyLink:      normalizedApplyLink,
		LinkedInJobURL: strings.TrimSpace(req.LinkedInJobURL),
		ResumeLink:     strings.TrimSpace(req.ResumeLink),
		Status:         status,
		Verdict:        verdict,
		DiscardReason:  discardReason,
		RejectReason:   rejectReason,
		TotalScore:     totalScore,
		SectionScores:  sectionScores,
		Extracted:      extracted,
		Flags:          flags,
		SalaryText:     strings.TrimSpace(req.SalaryText),
		IsEasyApply:    bool(strings.EqualFold(req.IsEasyApply, "true")),
		MatchRating:    req.MatchRating,
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

func (s *JobService) List(ctx context.Context, page, limit int, status, discardReason string, includeDiscarded bool, company, location, verdict string, minMatchRating, maxMatchRating *float64, scoreField string, minScore, maxScore *int, sortMatch, scoreSort string) ([]dao.Job, int64, int, int, error) {
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

	verdict = strings.ToUpper(strings.TrimSpace(verdict))
	if verdict != "" {
		if _, ok := globals.AllowedVerdicts[verdict]; !ok {
			return nil, 0, 0, 0, fmt.Errorf("invalid verdict: %w", globals.ErrBadRequest)
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

	if minMatchRating != nil {
		if *minMatchRating < 0 || *minMatchRating > 10 {
			return nil, 0, 0, 0, fmt.Errorf("min_match_rating must be between 0 and 10: %w", globals.ErrBadRequest)
		}
	}
	if maxMatchRating != nil {
		if *maxMatchRating < 0 || *maxMatchRating > 10 {
			return nil, 0, 0, 0, fmt.Errorf("max_match_rating must be between 0 and 10: %w", globals.ErrBadRequest)
		}
	}
	if minMatchRating != nil && maxMatchRating != nil && *minMatchRating > *maxMatchRating {
		return nil, 0, 0, 0, fmt.Errorf("min_match_rating cannot be greater than max_match_rating: %w", globals.ErrBadRequest)
	}

	sortMatch = strings.ToLower(strings.TrimSpace(sortMatch))
	if sortMatch != "" && sortMatch != "asc" && sortMatch != "desc" {
		return nil, 0, 0, 0, fmt.Errorf("invalid sort_match: %w", globals.ErrBadRequest)
	}

	scoreField = strings.ToLower(strings.TrimSpace(scoreField))
	if scoreField != "" {
		if _, ok := globals.AllowedScoreFields[scoreField]; !ok {
			return nil, 0, 0, 0, fmt.Errorf("invalid score_field: %w", globals.ErrBadRequest)
		}
	} else if minScore != nil || maxScore != nil || scoreSort != "" {
		return nil, 0, 0, 0, fmt.Errorf("score_field is required when using score filters or score sort: %w", globals.ErrBadRequest)
	}
	if minScore != nil && *minScore < 0 {
		return nil, 0, 0, 0, fmt.Errorf("min_score must be between 0 and 100: %w", globals.ErrBadRequest)
	}
	if maxScore != nil && *maxScore < 0 {
		return nil, 0, 0, 0, fmt.Errorf("max_score must be between 0 and 100: %w", globals.ErrBadRequest)
	}
	if minScore != nil && maxScore != nil && *minScore > *maxScore {
		return nil, 0, 0, 0, fmt.Errorf("min_score cannot be greater than max_score: %w", globals.ErrBadRequest)
	}
	scoreSort = strings.ToLower(strings.TrimSpace(scoreSort))
	if scoreField != "" && scoreSort == "" {
		scoreSort = "desc"
	}
	if scoreSort != "" && scoreSort != "asc" && scoreSort != "desc" {
		return nil, 0, 0, 0, fmt.Errorf("invalid score_sort: %w", globals.ErrBadRequest)
	}

	jobs, total, err := s.dao.List(ctx, dao.ListJobsParams{
		Page:             page,
		Limit:            limit,
		Status:           status,
		DiscardReason:    discardReason,
		IncludeDiscarded: includeDiscarded,
		Company:          strings.TrimSpace(company),
		Location:         strings.TrimSpace(location),
		Verdict:          verdict,
		MinMatchRating:   minMatchRating,
		MaxMatchRating:   maxMatchRating,
		SortMatch:        sortMatch,
		ScoreField:       scoreField,
		MinScore:         minScore,
		MaxScore:         maxScore,
		ScoreSort:        scoreSort,
	})
	if err != nil {
		return nil, 0, 0, 0, err
	}

	return jobs, total, page, limit, nil
}

func (s *JobService) ListResumes(ctx context.Context, page, limit int) ([]dao.ResumeItem, int64, int, int, error) {
	if page <= 0 {
		page = globals.DefaultPage
	}
	if limit <= 0 {
		limit = globals.DefaultLimit
	}
	if limit > globals.MaxLimit {
		limit = globals.MaxLimit
	}

	items, total, err := s.dao.ListResumes(ctx, page, limit)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	return items, total, page, limit, nil
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

func (s *JobService) BulkUpdateStatus(ctx context.Context, ids []string, status, discardReason string) (int, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("ids is required: %w", globals.ErrBadRequest)
	}

	normalizedStatus := strings.ToLower(strings.TrimSpace(status))
	if _, ok := globals.AllowedStatuses[normalizedStatus]; !ok {
		return 0, fmt.Errorf("invalid status: %w", globals.ErrBadRequest)
	}

	normalizedDiscardReason := strings.ToLower(strings.TrimSpace(discardReason))
	var targetDiscardReason *string
	if normalizedStatus == globals.StatusDiscarded {
		if normalizedDiscardReason == "" {
			return 0, fmt.Errorf("discard_reason is required when status is discarded: %w", globals.ErrBadRequest)
		}
		if _, ok := globals.AllowedDiscardReasons[normalizedDiscardReason]; !ok {
			return 0, fmt.Errorf("invalid discard_reason: %w", globals.ErrBadRequest)
		}
		targetDiscardReason = &normalizedDiscardReason
	} else if normalizedDiscardReason != "" {
		return 0, fmt.Errorf("discard_reason is only allowed when status is discarded: %w", globals.ErrBadRequest)
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

	updatedCount, err := s.dao.BulkUpdateStatus(ctx, parsedIDs, normalizedStatus, targetDiscardReason)
	if err != nil {
		return 0, err
	}

	return int(updatedCount), nil
}

func (s *JobService) ExistsByApplyLink(ctx context.Context, applyLink string) (bool, error) {
	normalized := normalizeApplyLink(applyLink)
	if normalized == "" {
		return false, fmt.Errorf("apply_link is required: %w", globals.ErrBadRequest)
	}
	return s.dao.ExistsByApplyLink(ctx, normalized)
}

func (s *JobService) GetApplyRateStats(ctx context.Context) (dto.ApplyRateStatsResponse, error) {
	totalApplied, firstAppliedAt, err := s.dao.GetApplyStatsBase(ctx)
	if err != nil {
		return dto.ApplyRateStatsResponse{}, err
	}

	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startOfWeek := startOfToday.AddDate(0, 0, -6)
	startOfMonth := startOfToday.AddDate(0, 0, -29)

	dailyCount, err := s.dao.CountAppliedBetween(ctx, startOfToday, now)
	if err != nil {
		return dto.ApplyRateStatsResponse{}, err
	}

	weeklyCount, err := s.dao.CountAppliedBetween(ctx, startOfWeek, now)
	if err != nil {
		return dto.ApplyRateStatsResponse{}, err
	}

	monthlyCount, err := s.dao.CountAppliedBetween(ctx, startOfMonth, now)
	if err != nil {
		return dto.ApplyRateStatsResponse{}, err
	}

	if totalApplied == 0 || firstAppliedAt == nil {
		return dto.ApplyRateStatsResponse{
			DailyCount:   dailyCount,
			WeeklyCount:  weeklyCount,
			MonthlyCount: monthlyCount,
		}, nil
	}

	first := firstAppliedAt.UTC()

	firstDay := time.Date(first.Year(), first.Month(), first.Day(), 0, 0, 0, 0, time.UTC)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	elapsedDays := int(today.Sub(firstDay).Hours()/24) + 1
	if elapsedDays < 1 {
		elapsedDays = 1
	}

	elapsedWeeks := ((elapsedDays - 1) / 7) + 1
	if elapsedWeeks < 1 {
		elapsedWeeks = 1
	}

	elapsedMonths := (today.Year()-firstDay.Year())*12 + int(today.Month()-firstDay.Month()) + 1
	if elapsedMonths < 1 {
		elapsedMonths = 1
	}

	total := float64(totalApplied)

	return dto.ApplyRateStatsResponse{
		DailyCount:     dailyCount,
		WeeklyCount:    weeklyCount,
		MonthlyCount:   monthlyCount,
		DailyAverage:   total / float64(elapsedDays),
		WeeklyAverage:  total / float64(elapsedWeeks),
		MonthlyAverage: total / float64(elapsedMonths),
	}, nil
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
	verdict := strings.ToUpper(strings.TrimSpace(derefString(req.Verdict)))
	if verdict == "" && isVerdictLikeStatus(status) {
		verdict = strings.ToUpper(status)
	}
	if verdict != "" {
		if _, ok := globals.AllowedVerdicts[verdict]; !ok {
			return fmt.Errorf("invalid verdict: %w", globals.ErrBadRequest)
		}
	} else if _, ok := globals.AllowedStatuses[status]; !ok {
		return fmt.Errorf("invalid status: %w", globals.ErrBadRequest)
	}

	discardReason := strings.ToLower(strings.TrimSpace(req.DiscardReason))
	rejectReason := strings.ToLower(strings.TrimSpace(derefString(req.RejectReason)))
	if verdict == globals.VerdictReject {
		if discardReason == "" && rejectReason == "" {
			return fmt.Errorf("reject_reason is required when verdict is reject: %w", globals.ErrBadRequest)
		}
		reason := rejectReason
		if reason == "" {
			reason = discardReason
		}
		if _, ok := globals.AllowedDiscardReasons[reason]; !ok {
			return fmt.Errorf("invalid reject_reason: %w", globals.ErrBadRequest)
		}
	} else if verdict != "" {
		if discardReason != "" || rejectReason != "" {
			return fmt.Errorf("discard_reason is only allowed when verdict is reject: %w", globals.ErrBadRequest)
		}
	} else if status == globals.StatusDiscarded {
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

	if req.MatchRating != nil {
		if *req.MatchRating < 0 || *req.MatchRating > 10 {
			return fmt.Errorf("match_rating must be between 0 and 10: %w", globals.ErrBadRequest)
		}
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
	if req.JobDescription != nil {
		provided = true
		value := strings.TrimSpace(*req.JobDescription)
		params.JobDescription = &value
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
		resolvedStatus, resolvedVerdict, resolvedDiscardReason, resolvedRejectReason, err := resolveJobStatusInputs(*req.Status, derefString(req.DiscardReason), req.Verdict, req.RejectReason)
		if err != nil {
			return params, err
		}
		params.Status = &resolvedStatus
		targetStatus = resolvedStatus
		params.Verdict = resolvedVerdict
		params.DiscardReason = resolvedDiscardReason
		params.RejectReason = resolvedRejectReason
	}
	if req.DiscardReason != nil {
		provided = true
		if req.Status == nil {
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
	}
	if req.Verdict != nil {
		provided = true
		resolvedStatus, resolvedVerdict, resolvedDiscardReason, resolvedRejectReason, err := resolveJobStatusInputs(derefString(req.Status), derefString(req.DiscardReason), req.Verdict, req.RejectReason)
		if err != nil {
			return params, err
		}
		params.Status = &resolvedStatus
		targetStatus = resolvedStatus
		params.Verdict = resolvedVerdict
		params.DiscardReason = resolvedDiscardReason
		params.RejectReason = resolvedRejectReason
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
	if req.MatchRating.Set {
		provided = true
		if req.MatchRating.Value == nil {
			params.ClearMatchRating = true
		} else {
			value := *req.MatchRating.Value
			if value < 0 || value > 10 {
				return params, fmt.Errorf("match_rating must be between 0 and 10: %w", globals.ErrBadRequest)
			}
			params.MatchRating = &value
		}
	}
	if req.AppliedAt != nil {
		provided = true
		if req.AppliedAt.IsZero() {
			return params, fmt.Errorf("applied_at cannot be zero: %w", globals.ErrBadRequest)
		}
		params.AppliedAt = req.AppliedAt
	}
	if req.TotalScore != nil {
		provided = true
		if *req.TotalScore < 0 || *req.TotalScore > 100 {
			return params, fmt.Errorf("total_score must be between 0 and 100: %w", globals.ErrBadRequest)
		}
		params.TotalScore = req.TotalScore
	}
	if req.SectionScores != nil {
		provided = true
		sectionScores, err := marshalJSONOrDefault(req.SectionScores, dto.JobSectionScores{})
		if err != nil {
			return params, err
		}
		params.SectionScores = sectionScores
	}
	if req.Extracted != nil {
		provided = true
		extracted, err := marshalJSONOrDefault(req.Extracted, dto.JobExtractedData{})
		if err != nil {
			return params, err
		}
		params.Extracted = extracted
	}
	if req.Flags != nil {
		provided = true
		flags, err := marshalJSONOrDefault(req.Flags, []string{})
		if err != nil {
			return params, err
		}
		params.Flags = flags
	}
	if req.RejectReason != nil {
		provided = true
		value := strings.ToLower(strings.TrimSpace(*req.RejectReason))
		if value == "" {
			params.RejectReason = nil
		} else {
			params.RejectReason = &value
		}
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

func resolveJobStatusInputs(rawStatus string, rawDiscardReason string, rawVerdict *string, rawRejectReason *string) (string, *string, *string, *string, error) {
	status := strings.ToLower(strings.TrimSpace(rawStatus))
	discardReason := strings.ToLower(strings.TrimSpace(rawDiscardReason))
	rejectReason := strings.ToLower(strings.TrimSpace(derefString(rawRejectReason)))
	verdict := strings.ToUpper(strings.TrimSpace(derefString(rawVerdict)))

	if verdict == "" && isVerdictLikeStatus(status) {
		verdict = strings.ToUpper(status)
	}

	if verdict != "" {
		if _, ok := globals.AllowedVerdicts[verdict]; !ok {
			return "", nil, nil, nil, fmt.Errorf("invalid verdict: %w", globals.ErrBadRequest)
		}
		resolvedStatus, resolvedDiscardReason, err := deriveStatusFromVerdict(verdict, rejectReason, discardReason)
		if err != nil {
			return "", nil, nil, nil, err
		}

		if status != "" && !isVerdictLikeStatus(status) && status != resolvedStatus {
			return "", nil, nil, nil, fmt.Errorf("status does not match verdict: %w", globals.ErrBadRequest)
		}

		resolvedVerdict := verdict
		resolvedRejectReason := rejectReason
		if resolvedRejectReason == "" {
			resolvedRejectReason = discardReason
		}
		return resolvedStatus, &resolvedVerdict, resolvedDiscardReason, stringPtrOrNil(resolvedRejectReason), nil
	}

	if status == "" {
		return "", nil, nil, nil, fmt.Errorf("status or verdict is required: %w", globals.ErrBadRequest)
	}
	if _, ok := globals.AllowedStatuses[status]; !ok {
		return "", nil, nil, nil, fmt.Errorf("invalid status: %w", globals.ErrBadRequest)
	}

	if status == globals.StatusDiscarded {
		if discardReason == "" {
			return "", nil, nil, nil, fmt.Errorf("discard_reason is required when status is discarded: %w", globals.ErrBadRequest)
		}
		if _, ok := globals.AllowedDiscardReasons[discardReason]; !ok {
			return "", nil, nil, nil, fmt.Errorf("invalid discard_reason: %w", globals.ErrBadRequest)
		}
		return status, nil, &discardReason, stringPtrOrNil(rejectReason), nil
	}

	if discardReason != "" {
		return "", nil, nil, nil, fmt.Errorf("discard_reason is only allowed when status is discarded: %w", globals.ErrBadRequest)
	}

	return status, nil, nil, stringPtrOrNil(rejectReason), nil
}

func deriveStatusFromVerdict(verdict, rejectReason, fallbackDiscardReason string) (string, *string, error) {
	switch verdict {
	case globals.VerdictApply, globals.VerdictReview:
		if rejectReason != "" || fallbackDiscardReason != "" {
			return "", nil, fmt.Errorf("discard_reason is only allowed when verdict is reject: %w", globals.ErrBadRequest)
		}
		return globals.StatusAdded, nil, nil
	case globals.VerdictReject:
		reason := rejectReason
		if reason == "" {
			reason = fallbackDiscardReason
		}
		if reason == "" {
			return "", nil, fmt.Errorf("reject_reason is required when verdict is reject: %w", globals.ErrBadRequest)
		}
		if _, ok := globals.AllowedDiscardReasons[reason]; !ok {
			return "", nil, fmt.Errorf("invalid reject_reason: %w", globals.ErrBadRequest)
		}
		return globals.StatusDiscarded, &reason, nil
	default:
		return "", nil, fmt.Errorf("invalid verdict: %w", globals.ErrBadRequest)
	}
}

func isVerdictLikeStatus(status string) bool {
	_, ok := globals.AllowedVerdicts[strings.ToUpper(strings.TrimSpace(status))]
	return ok
}

func marshalAssessmentPayload(sectionScores *dto.JobSectionScores, extracted *dto.JobExtractedData, flags []string, totalScore *int) ([]byte, []byte, []byte, *int, error) {
	var sectionBytes []byte
	var err error
	if sectionScores == nil {
		sectionBytes = []byte("{}")
	} else {
		sectionBytes, err = json.Marshal(sectionScores)
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}

	var extractedBytes []byte
	if extracted == nil {
		extractedBytes = []byte("{}")
	} else {
		extractedBytes, err = json.Marshal(extracted)
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}

	if flags == nil {
		flags = []string{}
	}
	flagsBytes, err := json.Marshal(flags)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if totalScore != nil && (*totalScore < 0 || *totalScore > 100) {
		return nil, nil, nil, nil, fmt.Errorf("total_score must be between 0 and 100: %w", globals.ErrBadRequest)
	}

	return sectionBytes, extractedBytes, flagsBytes, totalScore, nil
}

func marshalJSONOrDefault[T any](value *T, zero T) ([]byte, error) {
	if value == nil {
		return json.Marshal(zero)
	}
	return json.Marshal(value)
}

func stringPtrOrNil(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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
