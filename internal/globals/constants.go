package globals

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

const (
	StatusApplied   = "applied"
	StatusInterview = "interview"
	StatusOffer     = "offer"
	StatusRejected  = "rejected"
	StatusWithdrawn = "withdrawn"
)

var AllowedStatuses = map[string]struct{}{
	StatusApplied:   {},
	StatusInterview: {},
	StatusOffer:     {},
	StatusRejected:  {},
	StatusWithdrawn: {},
}
