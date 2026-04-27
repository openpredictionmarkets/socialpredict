package analytics

import (
	"context"
	"testing"
	"time"

	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/setup"

	"gorm.io/gorm"
)

type analyticsServiceTestConfig struct {
	repoOptions    []RepositoryOption
	serviceOptions []ServiceOption
}

type analyticsServiceTestOption func(*analyticsServiceTestConfig)

type financialSnapshotComputer interface {
	ComputeUserFinancials(context.Context, FinancialSnapshotRequest) (*FinancialSnapshot, error)
}

func withAnalyticsRepositoryOption(opt RepositoryOption) analyticsServiceTestOption {
	return func(cfg *analyticsServiceTestConfig) {
		cfg.repoOptions = append(cfg.repoOptions, opt)
	}
}

func withAnalyticsServiceOption(opt ServiceOption) analyticsServiceTestOption {
	return func(cfg *analyticsServiceTestConfig) {
		cfg.serviceOptions = append(cfg.serviceOptions, opt)
	}
}

func analyticsConfigFromSetup(econ *setup.EconomicConfig) Config {
	return Config{
		MaximumDebtAllowed: econ.Economics.User.MaximumDebtAllowed,
		CreateMarketCost:   econ.Economics.MarketIncentives.CreateMarketCost,
		InitialBetFee:      econ.Economics.Betting.BetFees.InitialBetFee,
	}
}

func newAnalyticsService(t *testing.T, db *gorm.DB, econ *setup.EconomicConfig, opts ...analyticsServiceTestOption) *Service {
	t.Helper()

	cfg := analyticsServiceTestConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	wpamCalculator := modelstesting.SeedWPAMFromConfig(econ)
	positionCalculator := positionsmath.NewPositionCalculator(
		positionsmath.WithProbabilityProvider(positionsmath.NewWPAMProbabilityProvider(wpamCalculator)),
	)
	cfg.repoOptions = append(cfg.repoOptions, WithRepositoryPositionCalculator(NewMarketPositionCalculator(positionCalculator)))
	repo := NewGormRepository(db, cfg.repoOptions...)
	cfg.serviceOptions = append(cfg.serviceOptions, WithPositionCalculator(NewMarketPositionCalculator(positionCalculator)))

	return NewService(repo, analyticsConfigFromSetup(econ), cfg.serviceOptions...)
}

func requireFinancialSnapshot(t *testing.T, svc financialSnapshotComputer, user models.User) *FinancialSnapshot {
	t.Helper()

	snapshot, err := svc.ComputeUserFinancials(context.Background(), FinancialSnapshotRequest{
		Username:       user.Username,
		AccountBalance: user.AccountBalance,
	})
	if err != nil {
		t.Fatalf("ComputeUserFinancials returned error: %v", err)
	}

	return snapshot
}

func TestComputeUserFinancials_NewUser_NoPositions(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	user := modelstesting.GenerateUser("testuser", 1000)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	econ := modelstesting.GenerateEconomicConfig()
	svc := newAnalyticsService(t, db, econ)

	snapshot := requireFinancialSnapshot(t, svc, user)

	if snapshot.AccountBalance != 1000 {
		t.Errorf("expected account balance 1000, got %d", snapshot.AccountBalance)
	}
	if snapshot.MaximumDebtAllowed != econ.Economics.User.MaximumDebtAllowed {
		t.Errorf("unexpected max debt: %d", snapshot.MaximumDebtAllowed)
	}
	if snapshot.AmountInPlay != 0 || snapshot.TradingProfits != 0 || snapshot.TotalProfits != 0 {
		t.Errorf("expected zeroed snapshot, got %+v", snapshot)
	}
}

func TestComputeUserFinancials_NegativeBalance(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	user := modelstesting.GenerateUser("borrower", -50)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	econ := modelstesting.GenerateEconomicConfig()
	svc := newAnalyticsService(t, db, econ)

	snapshot := requireFinancialSnapshot(t, svc, user)

	if snapshot.AmountBorrowed != 50 {
		t.Errorf("expected amountBorrowed 50, got %d", snapshot.AmountBorrowed)
	}
	expectedEquity := int64(-100)
	if snapshot.Equity != expectedEquity {
		t.Errorf("expected equity %d, got %d", expectedEquity, snapshot.Equity)
	}
}

func TestComputeUserFinancials_WithActivePositions(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	user := modelstesting.GenerateUser("trader", 500)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	market := modelstesting.GenerateMarket(1, user.Username)
	market.IsResolved = false
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bets := []models.Bet{
		modelstesting.GenerateBet(100, "YES", user.Username, uint(market.ID), 0),
		modelstesting.GenerateBet(50, "NO", user.Username, uint(market.ID), time.Minute),
	}
	for _, bet := range bets {
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	econ := modelstesting.GenerateEconomicConfig()
	svc := newAnalyticsService(t, db, econ)

	snapshot := requireFinancialSnapshot(t, svc, user)

	if snapshot.AmountInPlay == 0 {
		t.Errorf("expected non-zero amount in play, got %d", snapshot.AmountInPlay)
	}
	if snapshot.AmountInPlayActive == 0 {
		t.Errorf("expected active amount in play, got %d", snapshot.AmountInPlayActive)
	}
	if snapshot.TotalSpent == 0 {
		t.Errorf("expected total spent > 0")
	}
}
