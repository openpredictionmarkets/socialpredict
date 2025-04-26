package marketshandlers

import (
	"testing"
	"errors"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"socialpredict/models"
	"socialpredict/handlers/users"
	"socialpredict/handlers/math/probabilities/wpam"
)

// Mock update user balance function for testing
func mockUpdateUserBalance(username string, amount int64, db *gorm.DB, operation string) error {
	// Just for testing: update a UserBalance table or a mock in memory
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return err
	}

	if operation == "win" || operation == "refund" {
		user.Balance += amount
	} else {
		return errors.New("unsupported operation")
	}

	return db.Save(&user).Error
}

// Setup an in-memory database for testing
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	// Create necessary tables
	db.AutoMigrate(&models.Market{}, &models.Bet{}, &wpam.ProbabilityChange{}, &models.User{})

	return db
}

// Test distributePayouts with N/A resolution (refunds case)
func TestDistributePayouts_NA(t *testing.T) {
	db := setupTestDB()

	// Replace real UpdateUserBalance with mock
	users.UpdateUserBalance = mockUpdateUserBalance

	// Setup test data
	user := models.User{Username: "alice", Balance: 0}
	market := models.Market{ID: 1, ResolutionResult: "N/A", IsResolved: true}
	bet := models.Bet{Username: "alice", Amount: 100, MarketID: 1}

	db.Create(&user)
	db.Create(&market)
	db.Create(&bet)

	err := distributePayouts(&market, db)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var updatedUser models.User
	db.Where("username = ?", "alice").First(&updatedUser)

	if updatedUser.Balance != 100 {
		t.Fatalf("Expected balance 100 after refund, got %d", updatedUser.Balance)
	}
}

// Test calculateDBPMPayouts for YES market resolution
func TestCalculateDBPMPayouts_YesWin(t *testing.T) {
	db := setupTestDB()

	// Replace real UpdateUserBalance with mock
	users.UpdateUserBalance = mockUpdateUserBalance

	// Setup test data
	userYes := models.User{Username: "yes_buyer", Balance: 0}
	userNo := models.User{Username: "no_buyer", Balance: 0}
	market := models.Market{ID: 1, ResolutionResult: "YES", IsResolved: true}

	betYes := models.Bet{Username: "yes_buyer", Amount: 100, Outcome: "YES", MarketID: 1}
	betNo := models.Bet{Username: "no_buyer", Amount: 100, Outcome: "NO", MarketID: 1}

	probChanges := []wpam.ProbabilityChange{
		{MarketID: 1, Probability: 0.5, CreatedAt: time.Now().Add(-2 * time.Hour)}, // Initial
		{MarketID: 1, Probability: 0.6, CreatedAt: time.Now().Add(-1 * time.Hour)}, // After yes bet
		{MarketID: 1, Probability: 0.4, CreatedAt: time.Now()},                     // After no bet
	}

	db.Create(&userYes)
	db.Create(&userNo)
	db.Create(&market)
	db.Create(&betYes)
	db.Create(&betNo)
	for _, p := range probChanges {
		db.Create(&p)
	}

	err := calculateDBPMPayouts(&market, db)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Assert: yes_buyer got a payout, no_buyer did not
	var updatedYes models.User
	var updatedNo models.User
	db.Where("username = ?", "yes_buyer").First(&updatedYes)
	db.Where("username = ?", "no_buyer").First(&updatedNo)

	if updatedYes.Balance <= 0 {
		t.Fatalf("Expected yes_buyer to have positive balance, got %d", updatedYes.Balance)
	}
	if updatedNo.Balance != 0 {
		t.Fatalf("Expected no_buyer to have 0 balance, got %d", updatedNo.Balance)
	}
}

func TestSingleUserWrongSide_NoPayout(t *testing.T) {
	db := setupTestDB()

	// Mock UpdateUserBalance
	users.UpdateUserBalance = mockUpdateUserBalance

	// Setup market
	market := models.Market{ID: 1, ResolutionResult: "YES", IsResolved: true}
	db.Create(&market)

	// Setup user and their final NO position
	user := models.User{Username: "loser", Balance: 0}
	db.Create(&user)

	// User places a NO bet
	bet := models.Bet{Username: "loser", Amount: 100, Outcome: "NO", MarketID: market.ID}
	db.Create(&bet)

	// Setup probability changes reflecting normal market movement
	probChanges := []wpam.ProbabilityChange{
		{MarketID: 1, Probability: 0.5, CreatedAt: time.Now().Add(-2 * time.Hour)},
		{MarketID: 1, Probability: 0.4, CreatedAt: time.Now()},
	}
	for _, p := range probChanges {
		db.Create(&p)
	}

	// Run payout
	err := calculateDBPMPayouts(&market, db)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check balance
	var updatedUser models.User
	db.First(&updatedUser, "username = ?", "loser")

	if updatedUser.Balance != 0 {
		t.Fatalf("Expected balance 0 for user on losing side, got %d", updatedUser.Balance)
	}
}

func TestTwoUsers_OneWinnerOneLoser(t *testing.T) {
	db := setupTestDB()

	// Mock UpdateUserBalance
	users.UpdateUserBalance = mockUpdateUserBalance

	// Setup market
	market := models.Market{ID: 2, ResolutionResult: "YES", IsResolved: true}
	db.Create(&market)

	// Setup users
	userYes := models.User{Username: "winner", Balance: 0}
	userNo := models.User{Username: "loser", Balance: 0}
	db.Create(&userYes)
	db.Create(&userNo)

	// YES user places YES bet
	betYes := models.Bet{Username: "winner", Amount: 100, Outcome: "YES", MarketID: market.ID}
	db.Create(&betYes)

	// NO user places NO bet
	betNo := models.Bet{Username: "loser", Amount: 100, Outcome: "NO", MarketID: market.ID}
	db.Create(&betNo)

	// Setup probability changes
	probChanges := []wpam.ProbabilityChange{
		{MarketID: 2, Probability: 0.5, CreatedAt: time.Now().Add(-2 * time.Hour)},
		{MarketID: 2, Probability: 0.6, CreatedAt: time.Now().Add(-1 * time.Hour)},
		{MarketID: 2, Probability: 0.4, CreatedAt: time.Now()},
	}
	for _, p := range probChanges {
		db.Create(&p)
	}

	// Run payout
	err := calculateDBPMPayouts(&market, db)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check balances
	var updatedYes models.User
	var updatedNo models.User
	db.First(&updatedYes, "username = ?", "winner")
	db.First(&updatedNo, "username = ?", "loser")

	if updatedNo.Balance != 0 {
		t.Fatalf("Expected losing user to have 0 balance, got %d", updatedNo.Balance)
	}

	if updatedYes.Balance <= 0 {
		t.Fatalf("Expected winning user to have positive balance, got %d", updatedYes.Balance)
	}
}
