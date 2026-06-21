package metricshandlers

import (
	"context"
	"time"

	analytics "socialpredict/internal/domain/analytics"
	"socialpredict/internal/domain/readmodels"
)

// SystemMetricsService defines the application-owned metrics seam exposed under /v0/.
// Runtime probes such as /health and /readyz remain separate infrastructure routes.
type SystemMetricsService interface {
	ComputeSystemMetrics(context.Context) (*analytics.SystemMetrics, error)
}

type SystemMetricsReadModelService interface {
	GetSystemMetricsReadModel(context.Context) (*analytics.SystemMetricsReadModel, error)
	RefreshSystemMetricsSnapshot(context.Context) (*analytics.SystemMetricsReadModel, error)
}

// GlobalLeaderboardService defines the application-owned leaderboard seam exposed under /v0/.
type GlobalLeaderboardService interface {
	ComputeGlobalLeaderboardSnapshot(context.Context) (*analytics.GlobalLeaderboardSnapshot, error)
}

type GlobalLeaderboardReadModelService interface {
	GetGlobalLeaderboardReadModel(context.Context, int, int) (*analytics.GlobalLeaderboardReadModel, error)
	RefreshGlobalLeaderboardSnapshot(context.Context) (*analytics.GlobalLeaderboardReadModel, error)
}

type FreshnessResponse struct {
	GeneratedAt            time.Time  `json:"generatedAt"`
	Source                 string     `json:"source"`
	TargetFreshnessSeconds int        `json:"targetFreshnessSeconds"`
	TransactionSafeRead    bool       `json:"transactionSafeRead"`
	IsStale                bool       `json:"isStale"`
	StaleReason            string     `json:"staleReason,omitempty"`
	MarkedStaleAt          *time.Time `json:"markedStaleAt,omitempty"`
}

func freshnessResponseFromDomain(freshness readmodels.Freshness) FreshnessResponse {
	return FreshnessResponse{
		GeneratedAt:            freshness.GeneratedAt,
		Source:                 freshness.Source,
		TargetFreshnessSeconds: freshness.TargetFreshnessSeconds,
		TransactionSafeRead:    freshness.TransactionSafeRead,
		IsStale:                freshness.IsStale,
		StaleReason:            freshness.StaleReason,
		MarkedStaleAt:          freshness.MarkedStaleAt,
	}
}
