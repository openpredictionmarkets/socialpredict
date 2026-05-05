package metricshandlers

import (
	"context"

	analytics "socialpredict/internal/domain/analytics"
)

// SystemMetricsService defines the application-owned metrics seam exposed under /v0/.
// Runtime probes such as /health and /readyz remain separate infrastructure routes.
type SystemMetricsService interface {
	ComputeSystemMetrics(context.Context) (*analytics.SystemMetrics, error)
}

// GlobalLeaderboardService defines the application-owned leaderboard seam exposed under /v0/.
type GlobalLeaderboardService interface {
	ComputeGlobalLeaderboardSnapshot(context.Context) (*analytics.GlobalLeaderboardSnapshot, error)
}
