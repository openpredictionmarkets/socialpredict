package marketshandlers

import (
	"socialpredict/handlers/math/payout"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"testing"
)

func TestDistributePayouts_NA(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	user := modelstesting.GenerateUser("alice", 0)
	market := models.Market{ID: 1, ResolutionResult: "N/A", IsResolved: true}
	bet := modelstesting.GenerateBet(100, "YES", "alice", 1, 0)

	db.Create(&user)
	db.Create(&market)
	db.Create(&bet)

	err := payout.DistributePayoutsWithRefund(&market, db)
	// N/A resolution should return an error since refunds are not yet implemented
	if err == nil {
		t.Fatalf("Expected error for N/A resolution (not yet implemented), got nil")
	}
}

func TestDistributePayouts_YesWin(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	userYes := modelstesting.GenerateUser("yes_buyer", 0)
	userNo := modelstesting.GenerateUser("no_buyer", 0)
	market := models.Market{ID: 1, ResolutionResult: "YES", IsResolved: true}

	betYes := modelstesting.GenerateBet(100, "YES", "yes_buyer", 1, 0)
	betNo := modelstesting.GenerateBet(100, "NO", "no_buyer", 1, 0)

	db.Create(&userYes)
	db.Create(&userNo)
	db.Create(&market)
	db.Create(&betYes)
	db.Create(&betNo)

	err := payout.DistributePayoutsWithRefund(&market, db)
	if err != nil {
		t.Fatalf("Expected no error for YES resolution, got %v", err)
	}

	var updatedYes models.User
	var updatedNo models.User
	db.Where("username = ?", "yes_buyer").First(&updatedYes)
	db.Where("username = ?", "no_buyer").First(&updatedNo)

	if updatedYes.AccountBalance <= 0 {
		t.Fatalf("Expected yes_buyer to have positive balance, got %d", updatedYes.AccountBalance)
	}
	if updatedNo.AccountBalance != 0 {
		t.Fatalf("Expected no_buyer to have 0 balance, got %d", updatedNo.AccountBalance)
	}
}

func TestSingleUserWrongSide_NoPayout(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	market := models.Market{ID: 1, ResolutionResult: "YES", IsResolved: true}
	db.Create(&market)

	user := modelstesting.GenerateUser("loser", 0)
	db.Create(&user)

	bet := modelstesting.GenerateBet(100, "NO", "loser", 1, 0)
	db.Create(&bet)

	err := payout.DistributePayoutsWithRefund(&market, db)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var updatedUser models.User
	db.First(&updatedUser, "username = ?", "loser")

	if updatedUser.AccountBalance != 0 {
		t.Fatalf("Expected balance 0 for user on losing side, got %d", updatedUser.AccountBalance)
	}
}

func TestTwoUsers_OneWinnerOneLoser(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	market := models.Market{ID: 2, ResolutionResult: "YES", IsResolved: true}
	db.Create(&market)

	userYes := modelstesting.GenerateUser("winner", 0)
	userNo := modelstesting.GenerateUser("loser", 0)
	db.Create(&userYes)
	db.Create(&userNo)

	betYes := modelstesting.GenerateBet(100, "YES", "winner", 2, 0)
	betNo := modelstesting.GenerateBet(100, "NO", "loser", 2, 0)

	db.Create(&betYes)
	db.Create(&betNo)

	err := payout.DistributePayoutsWithRefund(&market, db)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var updatedYes models.User
	var updatedNo models.User
	db.First(&updatedYes, "username = ?", "winner")
	db.First(&updatedNo, "username = ?", "loser")

	if updatedNo.AccountBalance != 0 {
		t.Fatalf("Expected losing user to have 0 balance, got %d", updatedNo.AccountBalance)
	}
	if updatedYes.AccountBalance <= 0 {
		t.Fatalf("Expected winning user to have positive balance, got %d", updatedYes.AccountBalance)
	}
}
