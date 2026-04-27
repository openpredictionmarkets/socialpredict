package analytics_test

import (
	"context"
	"testing"

	"socialpredict/internal/app"
	"socialpredict/internal/domain/analytics"
	dbets "socialpredict/internal/domain/bets"
	"socialpredict/internal/domain/boundary"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/setup"

	"gorm.io/gorm"
)

func analyticsConfigFromSetup(cfg *setup.EconomicConfig) analytics.Config {
	return analytics.Config{
		MaximumDebtAllowed: cfg.Economics.User.MaximumDebtAllowed,
		CreateMarketCost:   cfg.Economics.MarketIncentives.CreateMarketCost,
		InitialBetFee:      cfg.Economics.Betting.BetFees.InitialBetFee,
	}
}

func newAnalyticsMetricsService(db *gorm.DB, config analytics.Config, opts ...analytics.ServiceOption) *analytics.Service {
	repo := analytics.NewGormRepository(db)
	return analytics.NewService(repo, config, opts...)
}

type analyticsSystemMetricsComputer interface {
	ComputeSystemMetrics(context.Context) (*analytics.SystemMetrics, error)
}

func requireAnalyticsMetricInt64(t *testing.T, metric analytics.Int64MetricReader) int64 {
	t.Helper()
	return metric.Int64Value()
}

func requireAnalyticsSystemMetrics(t *testing.T, svc analyticsSystemMetricsComputer) *analytics.SystemMetrics {
	t.Helper()

	metrics, err := svc.ComputeSystemMetrics(context.Background())
	if err != nil {
		t.Fatalf("compute metrics: %v", err)
	}

	return metrics
}

func calculateParticipationFees(cfg *setup.EconomicConfig, bets []boundary.Bet) int64 {
	var total int64
	type betKey struct {
		username string
		marketID uint
	}
	seen := make(map[betKey]bool)
	initialFee := cfg.Economics.Betting.BetFees.InitialBetFee

	for _, bet := range bets {
		if bet.Amount <= 0 {
			continue
		}

		key := betKey{username: bet.Username, marketID: bet.MarketID}
		if !seen[key] {
			seen[key] = true
			total += initialFee
		}
	}

	return total
}

func TestComputeSystemMetrics_BalancedAfterFinalLockedBet(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econConfig, _ := modelstesting.UseStandardTestEconomics(t)

	users := []models.User{
		modelstesting.GenerateUser("alice", 0),
		modelstesting.GenerateUser("bob", 0),
		modelstesting.GenerateUser("carol", 0),
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	market := modelstesting.GenerateMarket(7001, users[0].Username)
	market.IsResolved = false
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	creationFee := econConfig.Economics.MarketIncentives.CreateMarketCost
	if err := modelstesting.AdjustUserBalance(db, users[0].Username, -creationFee); err != nil {
		t.Fatalf("apply creation fee: %v", err)
	}

	container := app.BuildApplicationWithConfigService(db, configsvc.NewStaticService(econConfig))
	betsService := container.GetBetsService()

	placeBet := func(username string, amount int64, outcome string) {
		if _, err := betsService.Place(context.Background(), dbets.PlaceRequest{
			Username: username,
			MarketID: uint(market.ID),
			Amount:   amount,
			Outcome:  outcome,
		}); err != nil {
			t.Fatalf("place bet for %s: %v", username, err)
		}
	}

	placeBet("alice", 10, "YES")
	placeBet("bob", 10, "NO")
	placeBet("alice", 10, "YES")
	placeBet("bob", 10, "NO")
	placeBet("carol", 30, "YES")

	svc := newAnalyticsMetricsService(db, analyticsConfigFromSetup(econConfig))
	metrics := requireAnalyticsSystemMetrics(t, svc)

	maxDebt := econConfig.Economics.User.MaximumDebtAllowed

	var usersAfter []models.User
	if err := db.Find(&usersAfter).Error; err != nil {
		t.Fatalf("load users: %v", err)
	}

	var expectedUnusedDebt int64
	for _, u := range usersAfter {
		used := int64(0)
		if u.AccountBalance < 0 {
			used = -u.AccountBalance
		}
		expectedUnusedDebt += maxDebt - used
	}

	bets, err := analytics.NewGormRepository(db).ListBetsForMarket(context.Background(), uint(market.ID))
	if err != nil {
		t.Fatalf("list bets: %v", err)
	}

	var expectedActiveVolume int64
	for _, b := range bets {
		expectedActiveVolume += b.Amount
	}

	participationFees := calculateParticipationFees(econConfig, bets)
	totalUtilized := expectedUnusedDebt + expectedActiveVolume + creationFee + participationFees
	totalCapacity := maxDebt * int64(len(usersAfter))

	if got := requireAnalyticsMetricInt64(t, metrics.MoneyUtilized.ActiveBetVolume); got != expectedActiveVolume {
		t.Fatalf("active volume mismatch: want %d got %d", expectedActiveVolume, got)
	}
	if got := requireAnalyticsMetricInt64(t, metrics.MoneyUtilized.UnusedDebt); got != expectedUnusedDebt {
		t.Fatalf("unused debt mismatch: want %d got %d", expectedUnusedDebt, got)
	}
	if got := requireAnalyticsMetricInt64(t, metrics.MoneyUtilized.MarketCreationFees); got != creationFee {
		t.Fatalf("creation fee mismatch: want %d got %d", creationFee, got)
	}
	if got := requireAnalyticsMetricInt64(t, metrics.MoneyUtilized.ParticipationFees); got != participationFees {
		t.Fatalf("participation fees mismatch: want %d got %d", participationFees, got)
	}
	if got := requireAnalyticsMetricInt64(t, metrics.MoneyUtilized.TotalUtilized); got != totalUtilized {
		t.Fatalf("total utilized mismatch: want %d got %d", totalUtilized, got)
	}
	if got := requireAnalyticsMetricInt64(t, metrics.MoneyCreated.UserDebtCapacity); got != totalCapacity {
		t.Fatalf("debt capacity mismatch: want %d got %d", totalCapacity, got)
	}
	if got := requireAnalyticsMetricInt64(t, metrics.Verification.Surplus); got != 0 {
		t.Fatalf("expected zero surplus, got %d", got)
	}
}

func TestResolveMarket_DistributesAllBetVolume(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econConfig, _ := modelstesting.UseStandardTestEconomics(t)

	users := []models.User{
		modelstesting.GenerateUser("patrick", 0),
		modelstesting.GenerateUser("jimmy", 0),
		modelstesting.GenerateUser("jyron", 0),
		modelstesting.GenerateUser("testuser03", 0),
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	market := modelstesting.GenerateMarket(8002, users[0].Username)
	market.IsResolved = false
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	creationFee := econConfig.Economics.MarketIncentives.CreateMarketCost
	if err := modelstesting.AdjustUserBalance(db, users[0].Username, -creationFee); err != nil {
		t.Fatalf("apply creation fee: %v", err)
	}

	container := app.BuildApplicationWithConfigService(db, configsvc.NewStaticService(econConfig))
	betsService := container.GetBetsService()
	placeBet := func(username string, amount int64, outcome string) {
		if _, err := betsService.Place(context.Background(), dbets.PlaceRequest{
			Username: username,
			MarketID: uint(market.ID),
			Amount:   amount,
			Outcome:  outcome,
		}); err != nil {
			t.Fatalf("place bet for %s: %v", username, err)
		}
	}

	placeBet("patrick", 50, "NO")
	placeBet("jimmy", 51, "NO")
	placeBet("jimmy", 51, "NO")
	placeBet("jyron", 10, "YES")
	placeBet("testuser03", 30, "YES")

	if err := container.GetMarketsService().ResolveMarket(context.Background(), int64(market.ID), "YES", market.CreatorUsername); err != nil {
		t.Fatalf("ResolveMarket: %v", err)
	}

	repo := analytics.NewGormRepository(db)
	metricsSvc := newAnalyticsMetricsService(db, analyticsConfigFromSetup(econConfig))
	metrics := requireAnalyticsSystemMetrics(t, metricsSvc)

	if surplus := metrics.Verification.SurplusValue(); surplus != 0 {
		t.Fatalf("expected zero surplus after resolution, got %d", surplus)
	}

	// Ensure no user holds simultaneous positive YES and NO shares post-resolution
	for _, user := range users {
		positions, err := repo.UserMarketPositions(context.Background(), user.Username)
		if err != nil {
			t.Fatalf("calculate positions: %v", err)
		}
		for _, pos := range positions {
			if pos.YesSharesOwned > 0 && pos.NoSharesOwned > 0 {
				t.Fatalf("user %s holds both YES and NO shares post-resolution", user.Username)
			}
		}
	}
}
