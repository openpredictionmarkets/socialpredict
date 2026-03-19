package payout

import (
	"testing"

	"socialpredict/models"
	modelstesting "socialpredict/models/modelstesting"
)

// TestSoleBettorUnlikelyEvent tests the scenario from issue #135:
// A sole bettor places a YES bet on an extremely unlikely event at low
// probability. The market resolves NO — the bettor's position is worth zero.
// No payout is distributed, and account balances are unchanged by resolution.
func TestSoleBettorUnlikelyEvent_ResolvesAgainstBet(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	// Market has very low initial probability (0.02 ≈ "sun explodes tomorrow")
	market := modelstesting.GenerateMarket(10, "creator")
	market.InitialProbability = 0.02
	market.ResolutionResult = "NO"
	market.IsResolved = true
	db.Create(&market)

	// Sole bettor: Alice starts with 1000, bets 10 on YES
	alice := modelstesting.GenerateUser("alice_sole", 1000)
	db.Create(&alice)

	// Record that the bet was placed (balance already deducted at bet time)
	db.Model(&alice).Update("account_balance", int64(990))
	bet := modelstesting.GenerateBet(10, "YES", "alice_sole", uint(market.ID), 0)
	db.Create(&bet)

	// Run resolution — market resolves NO, so the sole YES bettor loses
	err := DistributePayoutsWithRefund(&market, db)
	if err != nil {
		t.Fatalf("DistributePayoutsWithRefund returned error: %v", err)
	}

	// Alice should receive no payout — her balance stays at 990
	var updated models.User
	if err := db.First(&updated, "username = ?", "alice_sole").Error; err != nil {
		t.Fatalf("failed to fetch alice_sole: %v", err)
	}
	if updated.AccountBalance != 990 {
		t.Errorf("alice_sole balance = %d, want 990 (no payout when sole bettor loses)", updated.AccountBalance)
	}
}

// TestSoleBettorUnlikelyEvent_ResolvesForBet is the complementary case:
// sole YES bettor, market resolves YES — bettor wins back their stake.
func TestSoleBettorUnlikelyEvent_ResolvesForBet(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	market := modelstesting.GenerateMarket(11, "creator")
	market.InitialProbability = 0.02
	market.ResolutionResult = "YES"
	market.IsResolved = true
	db.Create(&market)

	alice := modelstesting.GenerateUser("alice_winner", 990) // already paid 10 for bet
	db.Create(&alice)

	bet := modelstesting.GenerateBet(10, "YES", "alice_winner", uint(market.ID), 0)
	db.Create(&bet)

	err := DistributePayoutsWithRefund(&market, db)
	if err != nil {
		t.Fatalf("DistributePayoutsWithRefund returned error: %v", err)
	}

	var updatedWinner models.User
	if err := db.First(&updatedWinner, "username = ?", "alice_winner").Error; err != nil {
		t.Fatalf("failed to fetch alice_winner: %v", err)
	}
	// Sole YES bettor wins — gets the full 10 back
	if updatedWinner.AccountBalance != 1000 {
		t.Errorf("alice_winner balance = %d, want 1000 (sole bettor wins full stake)", updatedWinner.AccountBalance)
	}
}

func TestDistributePayoutsWithRefund_NARefund(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	market := modelstesting.GenerateMarket(1, "creator")
	market.ResolutionResult = "N/A"
	db.Create(&market)

	user := modelstesting.GenerateUser("refundbot", 0)
	db.Create(&user)

	bet := modelstesting.GenerateBet(50, "YES", "refundbot", uint(market.ID), 0)
	db.Create(&bet)

	err := DistributePayoutsWithRefund(&market, db)
	if err != nil {
		t.Fatalf("expected no error for N/A refund, got: %v", err)
	}

	// Verify the user received their refund
	var updatedUser models.User
	if err := db.First(&updatedUser, "username = ?", "refundbot").Error; err != nil {
		t.Fatalf("failed to fetch refundbot: %v", err)
	}

	expectedBalance := int64(50) // Should get the bet amount back
	if updatedUser.AccountBalance != expectedBalance {
		t.Errorf("refundbot balance = %d, want %d", updatedUser.AccountBalance, expectedBalance)
	}
}

func TestDistributePayoutsWithRefund_UnknownResolution(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	market := modelstesting.GenerateMarket(2, "creator")
	market.ResolutionResult = "MAYBE" // Invalid
	db.Create(&market)

	err := DistributePayoutsWithRefund(&market, db)
	if err == nil {
		t.Fatal("expected error for unknown resolution result")
	}
}

func TestCalculateAndAllocateProportionalPayouts_NoWinningShares(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	market := modelstesting.GenerateMarket(3, "creator")
	market.ResolutionResult = "YES"
	market.IsResolved = true
	db.Create(&market)

	// Create a user with a NO-side only bet (losing side)
	user := modelstesting.GenerateUser("loserbot", 0)
	db.Create(&user)

	bet := modelstesting.GenerateBet(100, "NO", "loserbot", uint(market.ID), 0)
	db.Create(&bet)

	err := calculateAndAllocateProportionalPayouts(&market, db)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var u models.User
	if err := db.First(&u, "username = ?", "loserbot").Error; err != nil {
		t.Fatalf("failed to fetch loserbot: %v", err)
	}

	expectedBalance := int64(0)
	if u.AccountBalance != expectedBalance {
		t.Errorf("loserbot balance = %d, want %d", u.AccountBalance, expectedBalance)
	}
}

func TestCalculateAndAllocateProportionalPayouts_SuccessfulPayout(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	market := modelstesting.GenerateMarket(4, "creator")
	market.ResolutionResult = "YES"
	market.IsResolved = true
	db.Create(&market)

	user := modelstesting.GenerateUser("winnerbot", 0)
	db.Create(&user)

	bet := modelstesting.GenerateBet(100, "YES", "winnerbot", uint(market.ID), 0)
	db.Create(&bet)

	err := calculateAndAllocateProportionalPayouts(&market, db)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var u models.User
	if err := db.First(&u, "username = ?", "winnerbot").Error; err != nil {
		t.Fatalf("failed to fetch winnerbot: %v", err)
	}

	// At resolution YES, winner gets full payout back from total volume
	expectedBalance := int64(100)
	if u.AccountBalance != expectedBalance {
		t.Errorf("winnerbot balance = %d, want %d", u.AccountBalance, expectedBalance)
	}
}
