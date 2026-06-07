package readmodels

import "time"

// Freshness describes when a display/read-model value was generated and how it
// should be interpreted. It is metadata only; it must not be used to authorize
// or settle transactions.
type Freshness struct {
	GeneratedAt            time.Time
	Source                 string
	TargetFreshnessSeconds int
	TransactionSafeRead    bool
}

// NewFreshness builds standardized display/read-model freshness metadata.
func NewFreshness(generatedAt time.Time, source string, target time.Duration, transactionSafeRead bool) Freshness {
	if source == "" {
		source = "read_model"
	}
	targetSeconds := int(target.Seconds())
	if targetSeconds < 0 {
		targetSeconds = 0
	}
	return Freshness{
		GeneratedAt:            generatedAt,
		Source:                 source,
		TargetFreshnessSeconds: targetSeconds,
		TransactionSafeRead:    transactionSafeRead,
	}
}
