package payout

import (
	"strings"
	"testing"

	"socialpredict/models"
	modelstesting "socialpredict/models/modelstesting"
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
	if err == nil || !strings.Contains(err.Error(), "not yet implemented") {
		t.Fatalf("expected refund not implemented error, got: %v", err)
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
