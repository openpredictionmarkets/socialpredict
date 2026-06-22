package analytics

import (
	domainanalytics "socialpredict/internal/domain/analytics"
	"socialpredict/internal/domain/boundary"
	positionsmath "socialpredict/internal/domain/math/positions"
)

type (
	AnalyticsReadModelSnapshot            = domainanalytics.AnalyticsReadModelSnapshot
	AnalyticsReadModelSnapshotRepository  = domainanalytics.AnalyticsReadModelSnapshotRepository
	Config                                = domainanalytics.Config
	DebtRepository                        = domainanalytics.DebtRepository
	FeeRepository                         = domainanalytics.FeeRepository
	FinancialSnapshot                     = domainanalytics.FinancialSnapshot
	FinancialSnapshotRequest              = domainanalytics.FinancialSnapshotRequest
	FinancialsRepository                  = domainanalytics.FinancialsRepository
	GlobalLeaderboardReadModel            = domainanalytics.GlobalLeaderboardReadModel
	GlobalUserProfitability               = domainanalytics.GlobalUserProfitability
	Int64MetricReader                     = domainanalytics.Int64MetricReader
	LeaderboardRepository                 = domainanalytics.LeaderboardRepository
	MarketGroupFeeRepository              = domainanalytics.MarketGroupFeeRepository
	MarketGroupFinancialsRepository       = domainanalytics.MarketGroupFinancialsRepository
	MarketPositionCalculator              = domainanalytics.MarketPositionCalculator
	MarketRecord                          = domainanalytics.MarketRecord
	Repository                            = domainanalytics.Repository
	Service                               = domainanalytics.Service
	ServiceOption                         = domainanalytics.ServiceOption
	StatsRepository                       = domainanalytics.StatsRepository
	SystemMetrics                         = domainanalytics.SystemMetrics
	SystemMetricsReadModel                = domainanalytics.SystemMetricsReadModel
	UserAccount                           = domainanalytics.UserAccount
	UserFinancialMetricSnapshot           = domainanalytics.UserFinancialMetricSnapshot
	UserFinancialMetricSnapshotRepository = domainanalytics.UserFinancialMetricSnapshotRepository
	VolumeRepository                      = domainanalytics.VolumeRepository
	WorkProfitMarketGroupRecord           = domainanalytics.WorkProfitMarketGroupRecord
	WorkProfitMarketRecord                = domainanalytics.WorkProfitMarketRecord
)

var (
	NewMarketPositionCalculator = domainanalytics.NewMarketPositionCalculator
	NewService                  = domainanalytics.NewService
	WithPositionCalculator      = domainanalytics.WithPositionCalculator
)

const (
	AnalyticsSnapshotKindSystemMetrics     = domainanalytics.AnalyticsSnapshotKindSystemMetrics
	AnalyticsSnapshotKindGlobalLeaderboard = domainanalytics.AnalyticsSnapshotKindGlobalLeaderboard
	SystemMetricsSnapshotKey               = domainanalytics.SystemMetricsSnapshotKey
	GlobalLeaderboardSnapshotKey           = domainanalytics.GlobalLeaderboardSnapshotKey
)

type defaultMarketPositionCalculator struct{}

func (defaultMarketPositionCalculator) Calculate(snapshot positionsmath.MarketSnapshot, bets []boundary.Bet) ([]positionsmath.MarketPosition, error) {
	return positionsmath.NewPositionCalculator().CalculateMarketPositions(snapshot, bets)
}
