package positionsmath

import (
	"testing"
	"time"
)

var adjustValuationBaseTime = time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)

func makeEarliestBets(base time.Time, userOffsets map[string]time.Duration) map[string]time.Time {
	earliest := make(map[string]time.Time, len(userOffsets))
	for user, offset := range userOffsets {
		earliest[user] = base.Add(offset)
	}
	return earliest
}

func makeUniformUserValuations(value int64, usernames ...string) map[string]UserValuationResult {
	userVals := make(map[string]UserValuationResult, len(usernames))
	for _, username := range usernames {
		userVals[username] = UserValuationResult{Username: username, RoundedValue: value}
	}
	return userVals
}

func assertRoundedValues(t *testing.T, adjusted map[string]UserValuationResult, want map[string]int64) {
	t.Helper()
	for user, expected := range want {
		if adjusted[user].RoundedValue != expected {
			t.Fatalf("user %s: want %d, got %d", user, expected, adjusted[user].RoundedValue)
		}
	}
}

func sumRoundedValues(adjusted map[string]UserValuationResult) int64 {
	var sum int64
	for _, value := range adjusted {
		sum += value.RoundedValue
	}
	return sum
}

func TestAdjustUserValuationsToMarketVolume(t *testing.T) {
	userBetOffsets := map[string]time.Duration{
		"alice": 0,
		"bob":   1 * time.Minute,
		"carol": 2 * time.Minute,
	}
	earliest := makeEarliestBets(adjustValuationBaseTime, userBetOffsets)

	userVals := makeUniformUserValuations(10, "alice", "bob", "carol")
	targetVolume := int64(32)
	adjusted := AdjustUserValuationsToMarketVolume(userVals, earliest, targetVolume)
	assertRoundedValues(t, adjusted, map[string]int64{"alice": 11, "bob": 11, "carol": 10})
	if sumRoundedValues(adjusted) != targetVolume {
		t.Fatalf("expected total %d, got %d", targetVolume, sumRoundedValues(adjusted))
	}

	userVals = makeUniformUserValuations(10, "alice", "bob", "carol")
	targetVolume = int64(28) // Remove 2
	adjusted = AdjustUserValuationsToMarketVolume(userVals, earliest, targetVolume)
	assertRoundedValues(t, adjusted, map[string]int64{"alice": 9, "bob": 9, "carol": 10})
}
