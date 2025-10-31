package analytics_test

import (
	"context"
	"testing"

	"socialpredict/internal/app"
	"socialpredict/internal/domain/analytics"
	dbets "socialpredict/internal/domain/bets"
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestComputeSystemMetrics_BalancedAfterFinalLockedBet(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econConfig, loadEcon := modelstesting.UseStandardTestEconomics(t)

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

	container := app.BuildApplication(db, econConfig)
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

	svc := analytics.NewService(analytics.NewGormRepository(db), loadEcon)
	metrics, err := svc.ComputeSystemMetrics(context.Background())
	if err != nil {
		t.Fatalf("compute metrics: %v", err)
	}

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

	participationFees := modelstesting.CalculateParticipationFees(econConfig, bets)
	totalUtilized := expectedUnusedDebt + expectedActiveVolume + creationFee + participationFees
	totalCapacity := maxDebt * int64(len(usersAfter))

	if got := metrics.MoneyUtilized.ActiveBetVolume.Value.(int64); got != expectedActiveVolume {
		t.Fatalf("active volume mismatch: want %d got %d", expectedActiveVolume, got)
	}
	if got := metrics.MoneyUtilized.UnusedDebt.Value.(int64); got != expectedUnusedDebt {
		t.Fatalf("unused debt mismatch: want %d got %d", expectedUnusedDebt, got)
	}
	if got := metrics.MoneyUtilized.MarketCreationFees.Value.(int64); got != creationFee {
		t.Fatalf("creation fee mismatch: want %d got %d", creationFee, got)
	}
	if got := metrics.MoneyUtilized.ParticipationFees.Value.(int64); got != participationFees {
		t.Fatalf("participation fees mismatch: want %d got %d", participationFees, got)
	}
	if got := metrics.MoneyUtilized.TotalUtilized.Value.(int64); got != totalUtilized {
		t.Fatalf("total utilized mismatch: want %d got %d", totalUtilized, got)
	}
	if got := metrics.MoneyCreated.UserDebtCapacity.Value.(int64); got != totalCapacity {
		t.Fatalf("debt capacity mismatch: want %d got %d", totalCapacity, got)
	}
	if got := metrics.Verification.Surplus.Value.(int64); got != 0 {
		t.Fatalf("expected zero surplus, got %d", got)
	}
}

func TestResolveMarket_DistributesAllBetVolume(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econConfig, loadEcon := modelstesting.UseStandardTestEconomics(t)

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

	container := app.BuildApplication(db, econConfig)
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

	metricsSvc := analytics.NewService(analytics.NewGormRepository(db), loadEcon)
	metrics, err := metricsSvc.ComputeSystemMetrics(context.Background())
	if err != nil {
		t.Fatalf("metrics after resolve: %v", err)
	}

	if surplus, _ := metrics.Verification.Surplus.Value.(int64); surplus != 0 {
		t.Fatalf("expected zero surplus after resolution, got %d", surplus)
	}

	// Ensure no user holds simultaneous positive YES and NO shares post-resolution
	for _, user := range users {
		positions, err := positionsmath.CalculateAllUserMarketPositions_WPAM_DBPM(db, user.Username)
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
