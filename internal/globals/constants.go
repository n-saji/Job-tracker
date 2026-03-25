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

var AllowedStatuses = map[string]struct{}{
	StatusAdded:     {},
	StatusApplied:   {},
	StatusInterview: {},
	StatusOffer:     {},
	StatusRejected:  {},
	StatusWithdrawn: {},
	StatusDiscarded: {},
}

const (
	DiscardReasonHighApplicants    = "high_applicants"
	DiscardReasonSecurityClearance = "security_clearance"
	DiscardReasonLessExperience    = "less_experience"
	DiscardReasonCitizenship       = "citizenship"
	DiscardReasonNotFit            = "not_fit"
)

var AllowedDiscardReasons = map[string]struct{}{
	DiscardReasonHighApplicants:    {},
	DiscardReasonSecurityClearance: {},
	DiscardReasonLessExperience:    {},
	DiscardReasonCitizenship:       {},
	DiscardReasonNotFit:            {},
}
