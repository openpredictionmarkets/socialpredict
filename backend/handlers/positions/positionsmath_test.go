package positions

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"strconv"
	"testing"
	"time"
)

func TestCalculateMarketPositions_WPAM_DBPM(t *testing.T) {

	db := modelstesting.NewFakeDB(t)

	// Create test data for the market and bets
	market := models.Market{
		ID:         1,
		IsResolved: false,
	}
	bet1 := models.Bet{
		Username: "user1",
		MarketID: 1,
		Amount:   3,
		Outcome:  "YES",
		PlacedAt: time.Date(2024, 5, 18, 5, 7, 31, 428975000, time.UTC),
	}
	bet2 := models.Bet{
		Username: "user2",
		MarketID: 1,
		Amount:   2,
		Outcome:  "NO",
		PlacedAt: time.Date(2024, 5, 18, 5, 8, 31, 428975000, time.UTC),
	}

	db.Create(&market)
	db.Create(&bet1)
	db.Create(&bet2)

	marketIdStr := strconv.Itoa(int(market.ID))

	// Run the function
	netPositions, err := CalculateMarketPositions_WPAM_DBPM(db, marketIdStr)

	// Verify the function did not return an error
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the returned net positions
	expectedPositions := []dbpm.MarketPosition{
		{Username: "user1", YesSharesOwned: 3, NoSharesOwned: 0},
		{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 2},
	}

	if len(netPositions) != len(expectedPositions) {
		t.Fatalf("expected %d net positions, got %d", len(expectedPositions), len(netPositions))
	}

	for i, pos := range netPositions {
		expected := expectedPositions[i]
		if pos.Username != expected.Username || pos.YesSharesOwned != expected.YesSharesOwned || pos.NoSharesOwned != expected.NoSharesOwned {
			t.Errorf("expected position %+v, got %+v", expected, pos)
		}
	}
}

func TestCalculateMarketPositionForUser_WPAM_DBPM(t *testing.T) {

	db := modelstesting.NewFakeDB(t)

	// Create test data for the market and bets
	market := models.Market{
		ID:         1,
		IsResolved: false,
	}
	bet1 := models.Bet{
		Username: "user1",
		MarketID: 1,
		Amount:   3,
		Outcome:  "YES",
		PlacedAt: time.Date(2024, 5, 18, 5, 7, 31, 428975000, time.UTC),
	}
	bet2 := models.Bet{
		Username: "user2",
		MarketID: 1,
		Amount:   2,
		Outcome:  "NO",
		PlacedAt: time.Date(2024, 5, 18, 5, 8, 31, 428975000, time.UTC),
	}

	db.Create(&market)
	db.Create(&bet1)
	db.Create(&bet2)

	marketIdStr := strconv.Itoa(int(market.ID))

	// Run the function for user1
	userPosition, err := CalculateMarketPositionForUser_WPAM_DBPM(db, marketIdStr, "user1")

	// Verify the function did not return an error
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the returned user position for user1
	expectedPosition := UserMarketPosition{
		YesSharesOwned: 3,
		NoSharesOwned:  0,
	}

	if userPosition.YesSharesOwned != expectedPosition.YesSharesOwned || userPosition.NoSharesOwned != expectedPosition.NoSharesOwned {
		t.Errorf("expected position %+v, got %+v", expectedPosition, userPosition)
	}

	// Run the function for user2
	userPosition, err = CalculateMarketPositionForUser_WPAM_DBPM(db, marketIdStr, "user2")

	// Verify the function did not return an error
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the returned user position for user2
	expectedPosition = UserMarketPosition{
		YesSharesOwned: 0,
		NoSharesOwned:  2,
	}

	if userPosition.YesSharesOwned != expectedPosition.YesSharesOwned || userPosition.NoSharesOwned != expectedPosition.NoSharesOwned {
		t.Errorf("expected position %+v, got %+v", expectedPosition, userPosition)
	}
}

func TestCheckOppositeSharesOwned(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	// Create test data for the market and bets
	market := models.Market{
		ID:         1,
		IsResolved: false,
	}
	bet1 := models.Bet{
		Username: "user1",
		MarketID: 1,
		Amount:   3,
		Outcome:  "YES",
		PlacedAt: time.Date(2024, 5, 18, 5, 7, 31, 428975000, time.UTC),
	}
	bet2 := models.Bet{
		Username: "user2",
		MarketID: 1,
		Amount:   2,
		Outcome:  "NO",
		PlacedAt: time.Date(2024, 5, 18, 5, 8, 31, 428975000, time.UTC),
	}

	db.Create(&market)
	db.Create(&bet1)
	db.Create(&bet2)

	marketIdStr := strconv.Itoa(int(market.ID))

	// Run the function for user1 buying NO shares
	oppositeShares, outcome, err := CheckOppositeSharesOwned(db, marketIdStr, "user1", "NO")

	// Verify the function did not return an error
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the returned opposite shares for user1
	if oppositeShares != 3 || outcome != "YES" {
		t.Errorf("expected opposite shares 3 and outcome 'YES', got %d and outcome '%s'", oppositeShares, outcome)
	}

	// Run the function for user2 buying YES shares
	oppositeShares, outcome, err = CheckOppositeSharesOwned(db, marketIdStr, "user2", "YES")

	// Verify the function did not return an error
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the returned opposite shares for user2
	if oppositeShares != 2 || outcome != "NO" {
		t.Errorf("expected opposite shares 2 and outcome 'NO', got %d and outcome '%s'", oppositeShares, outcome)
	}

	// Run the function for user1 buying YES shares (no opposite shares to sell)
	oppositeShares, outcome, err = CheckOppositeSharesOwned(db, marketIdStr, "user1", "YES")

	// Verify the function did not return an error
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the returned opposite shares for user1 (should be 0)
	if oppositeShares != 0 || outcome != "" {
		t.Errorf("expected opposite shares 0 and outcome '', got %d and outcome '%s'", oppositeShares, outcome)
	}
}
