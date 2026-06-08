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

func TestComputeUserFinancials_DerivesModeratorWorkProfitsFromResolvedStewardedMarkets(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	creator := modelstesting.GenerateUser("moderator", 500)
	steward := modelstesting.GenerateUser("steward", 500)
	alice := modelstesting.GenerateUser("alice", 500)
	bob := modelstesting.GenerateUser("bob", 500)
	for _, user := range []models.User{creator, steward, alice, bob} {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user %s: %v", user.Username, err)
		}
	}

	market := modelstesting.GenerateMarket(1, creator.Username)
	market.IsResolved = true
	market.ResolutionResult = "YES"
	market.ProposalCost = 10
	market.StewardUsername = steward.Username
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bets := []models.Bet{
		modelstesting.GenerateBet(100, "YES", alice.Username, uint(market.ID), 0),
		modelstesting.GenerateBet(-20, "YES", alice.Username, uint(market.ID), time.Minute),
		modelstesting.GenerateBet(50, "NO", bob.Username, uint(market.ID), 2*time.Minute),
		modelstesting.GenerateBet(10, "YES", bob.Username, uint(market.ID), 3*time.Minute),
	}
	for _, bet := range bets {
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.Betting.BetFees.InitialBetFee = 7
	econ.Economics.MarketIncentives.CreateMarketCost = 99
	svc := newAnalyticsService(t, db, econ)

	creatorSnapshot := requireFinancialSnapshot(t, svc, creator)
	if creatorSnapshot.WorkProfits != 0 {
		t.Fatalf("creator work profits = %d, want 0 after steward reassignment", creatorSnapshot.WorkProfits)
	}

	stewardSnapshot := requireFinancialSnapshot(t, svc, steward)

	expectedWorkProfits := int64(4)
	if stewardSnapshot.WorkProfits != expectedWorkProfits {
		t.Fatalf("steward work profits = %d, want %d", stewardSnapshot.WorkProfits, expectedWorkProfits)
	}
	if stewardSnapshot.TotalProfits != stewardSnapshot.TradingProfits+expectedWorkProfits {
		t.Fatalf("total profits = %d, want trading %d + work %d", stewardSnapshot.TotalProfits, stewardSnapshot.TradingProfits, expectedWorkProfits)
	}
}

func TestComputeUserFinancials_SubtractsCreationCostWhenCreatorRemainsSteward(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	creator := modelstesting.GenerateUser("creator_steward", 500)
	alice := modelstesting.GenerateUser("alice2", 500)
	bob := modelstesting.GenerateUser("bob2", 500)
	for _, user := range []models.User{creator, alice, bob} {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user %s: %v", user.Username, err)
		}
	}

	market := modelstesting.GenerateMarket(2, creator.Username)
	market.IsResolved = true
	market.ResolutionResult = "YES"
	market.ProposalCost = 10
	market.StewardUsername = creator.Username
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bets := []models.Bet{
		modelstesting.GenerateBet(100, "YES", alice.Username, uint(market.ID), 0),
		modelstesting.GenerateBet(50, "NO", bob.Username, uint(market.ID), time.Minute),
	}
	for _, bet := range bets {
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.Betting.BetFees.InitialBetFee = 7
	svc := newAnalyticsService(t, db, econ)

	snapshot := requireFinancialSnapshot(t, svc, creator)

	expectedWorkProfits := int64(4)
	if snapshot.WorkProfits != expectedWorkProfits {
		t.Fatalf("creator-steward work profits = %d, want %d", snapshot.WorkProfits, expectedWorkProfits)
	}
}

func TestComputeUserFinancials_WorkProfitsRemainZeroBelowCreationCostThreshold(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	creator := modelstesting.GenerateUser("threshold_creator", 500)
	steward := modelstesting.GenerateUser("threshold_steward", 500)
	alice := modelstesting.GenerateUser("threshold_alice", 500)
	bob := modelstesting.GenerateUser("threshold_bob", 500)
	for _, user := range []models.User{creator, steward, alice, bob} {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user %s: %v", user.Username, err)
		}
	}

	market := modelstesting.GenerateMarket(3, creator.Username)
	market.IsResolved = true
	market.ResolutionResult = "YES"
	market.ProposalCost = 10
	market.StewardUsername = steward.Username
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bets := []models.Bet{
		modelstesting.GenerateBet(100, "YES", alice.Username, uint(market.ID), 0),
		modelstesting.GenerateBet(50, "NO", bob.Username, uint(market.ID), time.Minute),
	}
	for _, bet := range bets {
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.Betting.BetFees.InitialBetFee = 3
	svc := newAnalyticsService(t, db, econ)

	snapshot := requireFinancialSnapshot(t, svc, steward)
	if snapshot.WorkProfits != 0 {
		t.Fatalf("steward work profits below threshold = %d, want 0", snapshot.WorkProfits)
	}
}
