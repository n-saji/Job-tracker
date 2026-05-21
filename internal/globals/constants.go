package globals

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

const (
	StatusAdded     = "added"
	StatusApplied   = "applied"
	StatusInterview = "interview"
	StatusOffer     = "offer"
	StatusRejected  = "rejected"
	StatusWithdrawn = "withdrawn"
	StatusDiscarded = "discarded"
)

const (
	VerdictApply  = "APPLY"
	VerdictReview = "REVIEW"
	VerdictReject = "REJECT"
)

var AllowedStatuses = map[string]struct{}{
	StatusAdded:     {},
	StatusApplied:   {},
	StatusInterview: {},
	StatusOffer:     {},
	StatusRejected:  {},
	StatusWithdrawn: {},
	StatusDiscarded: {},
}

var AllowedVerdicts = map[string]struct{}{
	VerdictApply:  {},
	VerdictReview: {},
	VerdictReject: {},
}

const (
	DiscardReasonHighApplicants    = "high_applicants"
	DiscardReasonSecurityClearance = "security_clearance"
	DiscardReasonLessExperience    = "less_experience"
	DiscardReasonCitizenship       = "citizenship"
	DiscardReasonNotFit            = "not_fit"
	DiscardReasonSponsorship      = "sponsorship"
	DiscardReasonPrimaryStack     = "primary_stack_mismatch"
)

var AllowedDiscardReasons = map[string]struct{}{
	DiscardReasonHighApplicants:    {},
	DiscardReasonSecurityClearance: {},
	DiscardReasonLessExperience:    {},
	DiscardReasonCitizenship:       {},
	DiscardReasonNotFit:            {},
	DiscardReasonSponsorship:      {},
	DiscardReasonPrimaryStack:     {},
}

const (
	ScoreFieldTotalScore      = "total_score"
	ScoreFieldSkillsMatch     = "skills_match"
	ScoreFieldYearsExperience = "years_of_experience"
	ScoreFieldLocation        = "location"
	ScoreFieldTitleAlignment  = "title_alignment"
	ScoreFieldEmploymentType  = "employment_type"
	ScoreFieldDomainRelevance = "domain_relevance"
)

var AllowedScoreFields = map[string]struct{}{
	ScoreFieldTotalScore:      {},
	ScoreFieldSkillsMatch:     {},
	ScoreFieldYearsExperience: {},
	ScoreFieldLocation:        {},
	ScoreFieldTitleAlignment:  {},
	ScoreFieldEmploymentType:  {},
	ScoreFieldDomainRelevance: {},
}
