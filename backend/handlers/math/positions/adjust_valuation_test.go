package positionsmath

import (
	"socialpredict/models/modelstesting"
	"testing"
	"time"

	"gorm.io/gorm"
)

// Helper: Create users and bets in DB and return bet times for asserts
func seedBetsAndTimes(t *testing.T, db *gorm.DB, marketID uint, userBetOffsets map[string]time.Duration) map[string]time.Time {
	betTimes := make(map[string]time.Time)
	for username, offset := range userBetOffsets {
		bet := modelstesting.GenerateBet(10, "YES", username, marketID, offset)
		db.Create(&bet)
		betTimes[username] = bet.PlacedAt
	}
	return betTimes
}

func TestGetAllUserEarliestBetsForMarket(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	market := modelstesting.GenerateMarket(1, "creator")
	db.Create(&market)

	// Simulate users with different first bet times
	userBetOffsets := map[string]time.Duration{
		"alice": 2 * time.Minute,
		"bob":   1 * time.Minute,
		"carol": 3 * time.Minute,
	}
	expectedTimes := seedBetsAndTimes(t, db, 1, userBetOffsets)

	earliestMap, err := GetAllUserEarliestBetsForMarket(db, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for user, wantTime := range expectedTimes {
		got, ok := earliestMap[user]
		if !ok {
			t.Errorf("expected user %s in map", user)
		}
		if !wantTime.Equal(got) {
			t.Errorf("user %s: want time %v, got %v", user, wantTime, got)
		}
	}
}

func TestAdjustUserValuationsToMarketVolume(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	market := modelstesting.GenerateMarket(2, "creator")
	db.Create(&market)

	// Users will have identical values, but alice's bet is earliest, then bob, then carol
	userBetOffsets := map[string]time.Duration{
		"alice": 0,
		"bob":   1 * time.Minute,
		"carol": 2 * time.Minute,
	}
	seedBetsAndTimes(t, db, 2, userBetOffsets)

	// All users have a rounded value of 10
	userVals := map[string]UserValuationResult{
		"alice": {Username: "alice", RoundedValue: 10},
		"bob":   {Username: "bob", RoundedValue: 10},
		"carol": {Username: "carol", RoundedValue: 10},
	}

	// Delta: need to add 2 (should go to alice then bob, since they are first by earliest bet)
	targetVolume := int64(32)
	adjusted, err := AdjustUserValuationsToMarketVolume(db, 2, userVals, targetVolume)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]int64{"alice": 11, "bob": 11, "carol": 10}
	for user, exp := range want {
		if adjusted[user].RoundedValue != exp {
			t.Errorf("user %s: want %d, got %d", user, exp, adjusted[user].RoundedValue)
		}
	}
	// Check total
	var sum int64
	for _, v := range adjusted {
		sum += v.RoundedValue
	}
	if sum != targetVolume {
		t.Errorf("expected total %d, got %d", targetVolume, sum)
	}

	// Test negative delta (removes from alice then bob)
	userVals = map[string]UserValuationResult{
		"alice": {Username: "alice", RoundedValue: 10},
		"bob":   {Username: "bob", RoundedValue: 10},
		"carol": {Username: "carol", RoundedValue: 10},
	}
	targetVolume = int64(28) // Remove 2
	adjusted, err = AdjustUserValuationsToMarketVolume(db, 2, userVals, targetVolume)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want = map[string]int64{"alice": 9, "bob": 9, "carol": 10}
	for user, exp := range want {
		if adjusted[user].RoundedValue != exp {
			t.Errorf("user %s: want %d, got %d", user, exp, adjusted[user].RoundedValue)
		}
	}
}
