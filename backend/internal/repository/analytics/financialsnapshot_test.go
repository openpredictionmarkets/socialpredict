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

func TestComputeUserFinancials_DerivesGroupWorkProfitsOnce(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	steward := modelstesting.GenerateUser("group_steward", 500)
	alice := modelstesting.GenerateUser("group_alice", 500)
	bob := modelstesting.GenerateUser("group_bob", 500)
	carol := modelstesting.GenerateUser("group_carol", 500)
	for _, user := range []models.User{steward, alice, bob, carol} {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user %s: %v", user.Username, err)
		}
	}

	childOne := modelstesting.GenerateMarket(101, steward.Username)
	childOne.IsResolved = true
	childOne.ResolutionResult = "YES"
	childOne.ProposalCost = 0
	childOne.StewardUsername = steward.Username
	childTwo := modelstesting.GenerateMarket(102, steward.Username)
	childTwo.IsResolved = true
	childTwo.ResolutionResult = "NO"
	childTwo.ProposalCost = 0
	childTwo.StewardUsername = steward.Username
	for _, market := range []models.Market{childOne, childTwo} {
		if err := db.Create(&market).Error; err != nil {
			t.Fatalf("create child market: %v", err)
		}
	}

	group := models.MarketGroup{
		QuestionTitle:      "Grouped market",
		Description:        "Parent",
		GroupType:          "MULTIPLE_CHOICE_BINARY",
		ProbabilityPolicy:  "INDEPENDENT_BINARY",
		ResolutionPolicy:   "INDEPENDENT_CHILDREN",
		LifecycleStatus:    "resolved",
		ProposalCost:       2,
		CreatorUsername:    steward.Username,
		StewardUsername:    steward.Username,
		ResolutionDateTime: time.Now().UTC().Add(time.Hour),
	}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create group: %v", err)
	}
	members := []models.MarketGroupMember{
		{GroupID: group.ID, MarketID: childOne.ID, AnswerLabel: "One", DisplayOrder: 0},
		{GroupID: group.ID, MarketID: childTwo.ID, AnswerLabel: "Two", DisplayOrder: 1},
	}
	for _, member := range members {
		if err := db.Create(&member).Error; err != nil {
			t.Fatalf("create member: %v", err)
		}
	}

	bets := []models.Bet{
		modelstesting.GenerateBet(10, "YES", alice.Username, uint(childOne.ID), 0),
		modelstesting.GenerateBet(10, "YES", bob.Username, uint(childOne.ID), time.Minute),
		modelstesting.GenerateBet(10, "NO", alice.Username, uint(childTwo.ID), 2*time.Minute),
		modelstesting.GenerateBet(10, "YES", carol.Username, uint(childTwo.ID), 3*time.Minute),
	}
	for _, bet := range bets {
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.MarketIncentives.CreateMarketCost = 1
	econ.Economics.Betting.BetFees.InitialBetFee = 2
	svc := newAnalyticsService(t, db, econ)

	snapshot := requireFinancialSnapshot(t, svc, steward)
	if snapshot.WorkProfits != 4 {
		t.Fatalf("group steward work profits = %d, want 4", snapshot.WorkProfits)
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

func TestComputeUserFinancials_WorkProfitsCanBeNegativeBelowCreationCostThreshold(t *testing.T) {
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
	if snapshot.WorkProfits != -4 {
		t.Fatalf("steward work profits below threshold = %d, want -4", snapshot.WorkProfits)
	}
}

func TestComputeUserFinancials_DerivesUnrealizedWorkProfitsFromUnresolvedStewardedMarkets(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	steward := modelstesting.GenerateUser("unrealized_steward", 500)
	alice := modelstesting.GenerateUser("unrealized_alice", 500)
	bob := modelstesting.GenerateUser("unrealized_bob", 500)
	for _, user := range []models.User{steward, alice, bob} {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user %s: %v", user.Username, err)
		}
	}

	standalone := modelstesting.GenerateMarket(401, steward.Username)
	standalone.IsResolved = false
	standalone.LifecycleStatus = "published"
	standalone.ProposalCost = 10
	standalone.StewardUsername = steward.Username
	if err := db.Create(&standalone).Error; err != nil {
		t.Fatalf("create standalone market: %v", err)
	}

	childOne := modelstesting.GenerateMarket(402, steward.Username)
	childOne.IsResolved = false
	childOne.LifecycleStatus = "published"
	childOne.ProposalCost = 0
	childOne.StewardUsername = steward.Username
	childTwo := modelstesting.GenerateMarket(403, steward.Username)
	childTwo.IsResolved = false
	childTwo.LifecycleStatus = "published"
	childTwo.ProposalCost = 0
	childTwo.StewardUsername = steward.Username
	for _, market := range []models.Market{childOne, childTwo} {
		if err := db.Create(&market).Error; err != nil {
			t.Fatalf("create child market: %v", err)
		}
	}

	group := models.MarketGroup{
		QuestionTitle:      "Unresolved group",
		Description:        "Parent",
		GroupType:          "MULTIPLE_CHOICE_BINARY",
		ProbabilityPolicy:  "INDEPENDENT_BINARY",
		ResolutionPolicy:   "INDEPENDENT_CHILDREN",
		LifecycleStatus:    "published",
		ProposalCost:       5,
		CreatorUsername:    steward.Username,
		StewardUsername:    steward.Username,
		ResolutionDateTime: time.Now().UTC().Add(time.Hour),
	}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create group: %v", err)
	}
	for _, member := range []models.MarketGroupMember{
		{GroupID: group.ID, MarketID: childOne.ID, AnswerLabel: "One", DisplayOrder: 0},
		{GroupID: group.ID, MarketID: childTwo.ID, AnswerLabel: "Two", DisplayOrder: 1},
	} {
		if err := db.Create(&member).Error; err != nil {
			t.Fatalf("create group member: %v", err)
		}
	}

	bets := []models.Bet{
		modelstesting.GenerateBet(10, "YES", alice.Username, uint(standalone.ID), 0),
		modelstesting.GenerateBet(10, "NO", bob.Username, uint(standalone.ID), time.Minute),
		modelstesting.GenerateBet(5, "YES", alice.Username, uint(childOne.ID), 2*time.Minute),
		modelstesting.GenerateBet(5, "NO", alice.Username, uint(childTwo.ID), 3*time.Minute),
		modelstesting.GenerateBet(5, "YES", bob.Username, uint(childTwo.ID), 4*time.Minute),
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
	// Standalone: 2 unique participants * 3 fee - 10 cost = -4.
	// Group: 2 unique participants across both children * 3 fee - 5 cost = 1.
	if snapshot.UnrealizedWorkIncome != 12 {
		t.Fatalf("unrealized work income = %d, want 12", snapshot.UnrealizedWorkIncome)
	}
	if snapshot.UnrealizedWorkProfits != -3 {
		t.Fatalf("unrealized work profits = %d, want -3", snapshot.UnrealizedWorkProfits)
	}
	if snapshot.WorkProfits != 0 {
		t.Fatalf("resolved work profits = %d, want 0 for unresolved markets", snapshot.WorkProfits)
	}
}

func TestComputeUserFinancials_UnrealizedWorkSeparatesCreatorCostFromStewardIncome(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	creator := modelstesting.GenerateUser("unrealized_creator_only", 500)
	steward := modelstesting.GenerateUser("unrealized_income_steward", 500)
	alice := modelstesting.GenerateUser("unrealized_income_alice", 500)
	for _, user := range []models.User{creator, steward, alice} {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user %s: %v", user.Username, err)
		}
	}

	market := modelstesting.GenerateMarket(501, creator.Username)
	market.IsResolved = false
	market.LifecycleStatus = "published"
	market.ProposalCost = 8
	market.StewardUsername = steward.Username
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}
	bet := modelstesting.GenerateBet(10, "YES", alice.Username, uint(market.ID), 0)
	if err := db.Create(&bet).Error; err != nil {
		t.Fatalf("create bet: %v", err)
	}

	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.Betting.BetFees.InitialBetFee = 3
	svc := newAnalyticsService(t, db, econ)

	creatorSnapshot := requireFinancialSnapshot(t, svc, creator)
	if creatorSnapshot.UnrealizedWorkIncome != 0 || creatorSnapshot.UnrealizedWorkProfits != -8 {
		t.Fatalf("creator unrealized work = income %d profit %d, want 0/-8", creatorSnapshot.UnrealizedWorkIncome, creatorSnapshot.UnrealizedWorkProfits)
	}

	stewardSnapshot := requireFinancialSnapshot(t, svc, steward)
	if stewardSnapshot.UnrealizedWorkIncome != 3 || stewardSnapshot.UnrealizedWorkProfits != 3 {
		t.Fatalf("steward unrealized work = income %d profit %d, want 3/3", stewardSnapshot.UnrealizedWorkIncome, stewardSnapshot.UnrealizedWorkProfits)
	}
}
