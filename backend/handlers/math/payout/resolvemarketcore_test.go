package payout

import (
	"socialpredict/models"
	modelstesting "socialpredict/models/modelstesting"
	"testing"
)

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
		t.Fatalf("unexpected error: %v", err)
	}

	var u models.User
	_ = db.First(&u, "username = ?", "refundbot")
	if u.AccountBalance != 50 {
		t.Errorf("refundbot balance = %d, want 50", u.AccountBalance)
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
	db.Create(&market)

	// Create losing bets only (NO side)
	user := modelstesting.GenerateUser("loserbot", 0)
	db.Create(&user)

	bet := modelstesting.GenerateBet(100, "NO", "loserbot", uint(market.ID), 0)
	db.Create(&bet)

	err := calculateAndAllocateProportionalPayouts(&market, db)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var u models.User
	_ = db.First(&u, "username = ?", "loserbot")
	if u.AccountBalance != 0 {
		t.Errorf("loserbot balance = %d, want 0", u.AccountBalance)
	}
}
