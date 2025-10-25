package financials

import (
	"fmt"
	"testing"

	buybetshandlers "socialpredict/handlers/bets/buying"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/setup"
)

func TestComputeSystemMetrics_BalancedAfterFinalLockedBet(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	// Align global economics configuration with deterministic test values
	econConfig := setup.EconomicsConfig()
	originalEconomics := econConfig.Economics
	defer func() {
		econConfig.Economics = originalEconomics
	}()
	econConfig.Economics = modelstesting.GenerateEconomicConfig().Economics

	loadEcon := func() *setup.EconomicConfig {
		return econConfig
	}

	// Prepare users
	users := []models.User{
		modelstesting.GenerateUser("alice", 0),
		modelstesting.GenerateUser("bob", 0),
		modelstesting.GenerateUser("carol", 0),
	}

	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("failed to create user %s: %v", users[i].Username, err)
		}
	}

	// Create market and apply creation fee to the creator to mirror production flow
	market := modelstesting.GenerateMarket(7001, users[0].Username)
	market.IsResolved = false
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("failed to create market: %v", err)
	}

	creationFee := econConfig.Economics.MarketIncentives.CreateMarketCost

	var creator models.User
	if err := db.Where("username = ?", users[0].Username).First(&creator).Error; err != nil {
		t.Fatalf("failed to load market creator: %v", err)
	}
	creator.AccountBalance -= creationFee
	if err := db.Save(&creator).Error; err != nil {
		t.Fatalf("failed to charge market creation fee: %v", err)
	}

	placeBet := func(username string, amount int64, outcome string) {
		var u models.User
		if err := db.Where("username = ?", username).First(&u).Error; err != nil {
			t.Fatalf("failed to load user %s: %v", username, err)
		}
		betReq := models.Bet{
			MarketID: uint(market.ID),
			Amount:   amount,
			Outcome:  outcome,
		}
		if _, err := buybetshandlers.PlaceBetCore(&u, betReq, db, loadEcon); err != nil {
			t.Fatalf("place bet failed for %s: %v", username, err)
		}
	}

	// Sequence of bets that leaves the final entrant without a position
	placeBet("alice", 5, "YES")
	placeBet("bob", 10, "YES")
	placeBet("carol", 3, "YES")
	placeBet("alice", 10, "NO")
	placeBet("bob", 10, "NO")
	placeBet("carol", 30, "YES")

	metrics, err := ComputeSystemMetrics(db, loadEcon)
	if err != nil {
		t.Fatalf("compute metrics failed: %v", err)
	}

	// Gather expected values directly from the database to assert accounting balance
	var dbUsers []models.User
	if err := db.Find(&dbUsers).Error; err != nil {
		t.Fatalf("failed to load users: %v", err)
	}

	maxDebt := econConfig.Economics.User.MaximumDebtAllowed
	var expectedUnusedDebt int64
	for _, u := range dbUsers {
		usedDebt := int64(0)
		if u.AccountBalance < 0 {
			usedDebt = -u.AccountBalance
		}
		expectedUnusedDebt += maxDebt - usedDebt
	}

	var bets []models.Bet
	if err := db.Where("market_id = ?", market.ID).Order("placed_at ASC").Find(&bets).Error; err != nil {
		t.Fatalf("failed to load bets: %v", err)
	}

	var expectedActiveVolume int64
	for _, b := range bets {
		expectedActiveVolume += b.Amount
	}

	seen := make(map[string]bool)
	var expectedParticipationFees int64
	initialFee := econConfig.Economics.Betting.BetFees.InitialBetFee
	for _, b := range bets {
		if b.Amount <= 0 {
			continue
		}
		key := fmt.Sprintf("%d:%s", b.MarketID, b.Username)
		if !seen[key] {
			seen[key] = true
			expectedParticipationFees += initialFee
		}
	}
	expectedCreationFees := creationFee // single market
	totalUtilized := expectedUnusedDebt + expectedActiveVolume + expectedCreationFees + expectedParticipationFees
	totalCapacity := maxDebt * int64(len(dbUsers))

	if got := metrics.MoneyUtilized.ActiveBetVolume.Value.(int64); got != expectedActiveVolume {
		t.Fatalf("active volume mismatch: expected %d, got %d", expectedActiveVolume, got)
	}

	if got := metrics.MoneyUtilized.UnusedDebt.Value.(int64); got != expectedUnusedDebt {
		t.Fatalf("unused debt mismatch: expected %d, got %d", expectedUnusedDebt, got)
	}

	if got := metrics.MoneyUtilized.MarketCreationFees.Value.(int64); got != expectedCreationFees {
		t.Fatalf("market creation fees mismatch: expected %d, got %d", expectedCreationFees, got)
	}

	if got := metrics.MoneyUtilized.ParticipationFees.Value.(int64); got != expectedParticipationFees {
		t.Fatalf("participation fees mismatch: expected %d, got %d", expectedParticipationFees, got)
	}

	if got := metrics.MoneyUtilized.TotalUtilized.Value.(int64); got != totalUtilized {
		t.Fatalf("total utilized mismatch: expected %d, got %d", totalUtilized, got)
	}

	if got := metrics.MoneyCreated.UserDebtCapacity.Value.(int64); got != totalCapacity {
		t.Fatalf("user debt capacity mismatch: expected %d, got %d", totalCapacity, got)
	}

	if got := metrics.Verification.Surplus.Value.(int64); got != 0 {
		t.Fatalf("expected zero surplus, got %d", got)
	}

	if balanced, ok := metrics.Verification.Balanced.Value.(bool); !ok || !balanced {
		t.Fatalf("expected metrics to be balanced, got %v", metrics.Verification.Balanced.Value)
	}
}
