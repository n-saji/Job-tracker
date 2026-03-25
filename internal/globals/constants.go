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
)

var AllowedStatuses = map[string]struct{}{
	StatusAdded:     {},
	StatusApplied:   {},
	StatusInterview: {},
	StatusOffer:     {},
	StatusRejected:  {},
	StatusWithdrawn: {},
}
