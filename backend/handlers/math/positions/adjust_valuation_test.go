package positionsmath

import (
	"socialpredict/models/modelstesting"
	"testing"
	"time"

	"gorm.io/gorm"
)

// Helper: Create users and bets in DB and return bet times for asserts
func seedBetsAndTimes(t *testing.T, db *gorm.DB, marketID uint, userBetOffsets map[string]time.Duration) map[string]time.Time {
	t.Helper()

	betTimes := make(map[string]time.Time)
	for username, offset := range userBetOffsets {
		bet := modelstesting.GenerateBet(10, "YES", username, marketID, offset)
		db.Create(&bet)
		betTimes[username] = bet.PlacedAt
	}
	return betTimes
}

func newEqualUserValuations(values map[string]int64) map[string]UserValuationResult {
	userVals := make(map[string]UserValuationResult, len(values))
	for username, roundedValue := range values {
		userVals[username] = UserValuationResult{Username: username, RoundedValue: roundedValue}
	}
	return userVals
}

func assertRoundedValues(t *testing.T, adjusted map[string]UserValuationResult, expected map[string]int64) {
	t.Helper()

	for user, want := range expected {
		if adjusted[user].RoundedValue != want {
			t.Errorf("user %s: want %d, got %d", user, want, adjusted[user].RoundedValue)
		}
	}
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

	userBetOffsets := map[string]time.Duration{
		"alice": 0,
		"bob":   1 * time.Minute,
		"carol": 2 * time.Minute,
	}
	seedBetsAndTimes(t, db, 2, userBetOffsets)

	tests := []struct {
		name         string
		targetVolume int64
		expected     map[string]int64
	}{
		{
			name:         "positive delta favors earlier bets",
			targetVolume: 32,
			expected:     map[string]int64{"alice": 11, "bob": 11, "carol": 10},
		},
		{
			name:         "negative delta removes from earlier bets first",
			targetVolume: 28,
			expected:     map[string]int64{"alice": 9, "bob": 9, "carol": 10},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userVals := newEqualUserValuations(map[string]int64{
				"alice": 10,
				"bob":   10,
				"carol": 10,
			})

			adjusted, err := AdjustUserValuationsToMarketVolume(db, 2, userVals, test.targetVolume)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertRoundedValues(t, adjusted, test.expected)

			var sum int64
			for _, v := range adjusted {
				sum += v.RoundedValue
			}
			if sum != test.targetVolume {
				t.Errorf("expected total %d, got %d", test.targetVolume, sum)
			}
		})
	}
}
