package analytics

import (
	"context"
	"reflect"
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

type systemMetricsComputer interface {
	ComputeSystemMetrics(context.Context) (*SystemMetrics, error)
}

func requireMetricInt64(t *testing.T, metric Int64MetricReader) int64 {
	t.Helper()
	return metric.Int64Value()
}

func requireSystemMetrics(t *testing.T, svc systemMetricsComputer) *SystemMetrics {
	t.Helper()

	metrics, err := svc.ComputeSystemMetrics(context.Background())
	if err != nil {
		t.Fatalf("ComputeSystemMetrics returned error: %v", err)
	}

	return metrics
}

func requireTableCount(t *testing.T, db *gorm.DB, model any) int64 {
	t.Helper()

	var count int64
	if err := db.Model(model).Count(&count).Error; err != nil {
		t.Fatalf("count %T: %v", model, err)
	}
	return count
}

func TestComputeSystemMetrics_EmptyDatabase(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econ := modelstesting.GenerateEconomicConfig()

	svc := newAnalyticsService(t, db, econ)
	metrics := requireSystemMetrics(t, svc)

	if val := metrics.MoneyCreated.UserDebtCapacityValue(); val != 0 {
		t.Fatalf("expected user debt capacity 0, got %d", val)
	}
	if val := metrics.MoneyUtilized.TotalUtilizedValue(); val != 0 {
		t.Fatalf("expected total utilized 0, got %d", val)
	}
}

func TestComputeSystemMetrics_WithData(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.MarketIncentives.CreateMarketCost = 50
	econ.Economics.Betting.BetFees.InitialBetFee = 5
	econ.Economics.User.MaximumDebtAllowed = 500

	users := []models.User{
		modelstesting.GenerateUser("user1", 950),
		modelstesting.GenerateUser("user2", -100),
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	market := modelstesting.GenerateMarket(1, users[0].Username)
	market.IsResolved = false
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bets := []models.Bet{
		modelstesting.GenerateBet(50, "YES", "user1", uint(market.ID), 0),
		modelstesting.GenerateBet(30, "YES", "user2", uint(market.ID), 0),
	}
	for _, bet := range bets {
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	svc := newAnalyticsService(t, db, econ)
	metrics := requireSystemMetrics(t, svc)

	if val := requireMetricInt64(t, metrics.MoneyCreated.UserDebtCapacity); val != 1000 {
		t.Errorf("expected user debt capacity 1000, got %d", val)
	}
	if val := requireMetricInt64(t, metrics.MoneyUtilized.UnusedDebt); val != 900 {
		t.Errorf("expected unused debt 900, got %d", val)
	}
	if val := requireMetricInt64(t, metrics.MoneyUtilized.MarketCreationFees); val != 50 {
		t.Errorf("expected market creation fees 50, got %d", val)
	}
	if val := requireMetricInt64(t, metrics.MoneyUtilized.ParticipationFees); val != 10 {
		t.Errorf("expected participation fees 10, got %d", val)
	}
}

func TestComputeSystemMetrics_UsesInjectedSnapshot(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.MarketIncentives.CreateMarketCost = 50
	econ.Economics.Betting.BetFees.InitialBetFee = 5
	econ.Economics.User.MaximumDebtAllowed = 500

	users := []models.User{
		modelstesting.GenerateUser("user1", 0),
		modelstesting.GenerateUser("user2", 0),
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	market := modelstesting.GenerateMarket(1, users[0].Username)
	market.IsResolved = false
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bet := modelstesting.GenerateBet(50, "YES", "user1", uint(market.ID), 0)
	if err := db.Create(&bet).Error; err != nil {
		t.Fatalf("create bet: %v", err)
	}

	snapshot := Config{
		MaximumDebtAllowed: econ.Economics.User.MaximumDebtAllowed,
		CreateMarketCost:   econ.Economics.MarketIncentives.CreateMarketCost,
		InitialBetFee:      econ.Economics.Betting.BetFees.InitialBetFee,
	}
	svc := NewService(NewGormRepository(db), snapshot)

	econ.Economics.User.MaximumDebtAllowed = 999
	econ.Economics.MarketIncentives.CreateMarketCost = 999
	econ.Economics.Betting.BetFees.InitialBetFee = 999

	metrics := requireSystemMetrics(t, svc)
	if val := requireMetricInt64(t, metrics.MoneyCreated.UserDebtCapacity); val != 1000 {
		t.Fatalf("expected frozen user debt capacity 1000, got %d", val)
	}
	if val := requireMetricInt64(t, metrics.MoneyUtilized.MarketCreationFees); val != 50 {
		t.Fatalf("expected frozen market creation fees 50, got %d", val)
	}
	if val := requireMetricInt64(t, metrics.MoneyUtilized.ParticipationFees); val != 5 {
		t.Fatalf("expected frozen participation fees 5, got %d", val)
	}
}

func TestComputeSystemMetrics_ReplayIsDeterministicAndReadOnly(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.MarketIncentives.CreateMarketCost = 25
	econ.Economics.Betting.BetFees.InitialBetFee = 7
	econ.Economics.User.MaximumDebtAllowed = 300

	users := []models.User{
		modelstesting.GenerateUser("alice", 200),
		modelstesting.GenerateUser("bob", -50),
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	activeMarket := modelstesting.GenerateMarket(1101, users[0].Username)
	activeMarket.IsResolved = false
	if err := db.Create(&activeMarket).Error; err != nil {
		t.Fatalf("create active market: %v", err)
	}

	resolvedMarket := modelstesting.GenerateMarket(1102, users[1].Username)
	resolvedMarket.IsResolved = true
	if err := db.Create(&resolvedMarket).Error; err != nil {
		t.Fatalf("create resolved market: %v", err)
	}

	bets := []models.Bet{
		modelstesting.GenerateBet(40, "YES", "alice", uint(activeMarket.ID), 0),
		modelstesting.GenerateBet(15, "NO", "bob", uint(activeMarket.ID), 0),
		modelstesting.GenerateBet(90, "YES", "alice", uint(resolvedMarket.ID), 0),
	}
	for _, bet := range bets {
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	svc := newAnalyticsService(t, db, econ)
	beforeCounts := map[string]int64{
		"users":   requireTableCount(t, db, &models.User{}),
		"markets": requireTableCount(t, db, &models.Market{}),
		"bets":    requireTableCount(t, db, &models.Bet{}),
	}

	first := requireSystemMetrics(t, svc)
	for i := 0; i < 5; i++ {
		replayed := requireSystemMetrics(t, svc)
		if !reflect.DeepEqual(first, replayed) {
			t.Fatalf("metrics replay %d changed result:\nfirst:  %+v\nreplay: %+v", i+1, first, replayed)
		}
	}

	afterCounts := map[string]int64{
		"users":   requireTableCount(t, db, &models.User{}),
		"markets": requireTableCount(t, db, &models.Market{}),
		"bets":    requireTableCount(t, db, &models.Bet{}),
	}
	if !reflect.DeepEqual(beforeCounts, afterCounts) {
		t.Fatalf("metrics replay changed persisted rows: before=%v after=%v", beforeCounts, afterCounts)
	}
}
