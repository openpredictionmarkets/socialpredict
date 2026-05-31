package markets

import (
	"strings"
	"time"
)

const (
	MarketStatusActive   = "active"
	MarketStatusClosed   = "closed"
	MarketStatusResolved = "resolved"
	MarketStatusAll      = "all"
)

const (
	MarketLifecycleProposed  = "proposed"
	MarketLifecycleRejected  = "rejected"
	MarketLifecyclePublished = "published"
	MarketLifecycleClosed    = "closed"
	MarketLifecycleResolved  = "resolved"
	MarketLifecycleCancelled = "cancelled"
)

func NormalizeLifecycleStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case MarketLifecycleProposed:
		return MarketLifecycleProposed
	case MarketLifecycleRejected:
		return MarketLifecycleRejected
	case MarketLifecycleClosed:
		return MarketLifecycleClosed
	case MarketLifecycleResolved:
		return MarketLifecycleResolved
	case MarketLifecycleCancelled:
		return MarketLifecycleCancelled
	case "", "active", MarketLifecyclePublished:
		return MarketLifecyclePublished
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func LifecycleFromPublicStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case MarketStatusResolved:
		return MarketLifecycleResolved
	case MarketLifecycleProposed, MarketLifecycleRejected, MarketLifecycleCancelled:
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return MarketLifecyclePublished
	}
}

func PublicStatusFromLifecycle(lifecycle string, resolved bool, resolutionTime time.Time, now time.Time) string {
	lifecycle = NormalizeLifecycleStatus(lifecycle)
	if resolved || lifecycle == MarketLifecycleResolved {
		return MarketStatusResolved
	}
	if lifecycle == MarketLifecycleProposed || lifecycle == MarketLifecycleRejected || lifecycle == MarketLifecycleCancelled {
		return lifecycle
	}
	if lifecycle == MarketLifecycleClosed || (!resolutionTime.IsZero() && !resolutionTime.After(now)) {
		return MarketStatusClosed
	}
	return MarketStatusActive
}

func (m *Market) ApplyLifecycleStatus(lifecycle string, now time.Time) error {
	if m == nil {
		return ErrInvalidInput
	}

	normalized := NormalizeLifecycleStatus(lifecycle)
	switch normalized {
	case MarketLifecycleProposed, MarketLifecycleRejected, MarketLifecyclePublished, MarketLifecycleClosed, MarketLifecycleResolved, MarketLifecycleCancelled:
	default:
		return ErrInvalidState
	}

	m.LifecycleStatus = normalized
	m.Status = PublicStatusFromLifecycle(normalized, normalized == MarketLifecycleResolved, m.ResolutionDateTime, now)
	if normalized == MarketLifecycleResolved {
		m.FinalResolutionDateTime = now
	}
	m.UpdatedAt = now
	return nil
}

func (m *Market) MarkProposed(now time.Time) error {
	return m.ApplyLifecycleStatus(MarketLifecycleProposed, now)
}

func (m *Market) Publish(now time.Time) error {
	return m.ApplyLifecycleStatus(MarketLifecyclePublished, now)
}

func (m *Market) Reject(now time.Time) error {
	return m.ApplyLifecycleStatus(MarketLifecycleRejected, now)
}

func (m *Market) Cancel(now time.Time) error {
	return m.ApplyLifecycleStatus(MarketLifecycleCancelled, now)
}
