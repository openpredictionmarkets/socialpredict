package positionsmath

import (
	"testing"
	"time"
)

func TestAdjustUserValuationsToMarketVolume(t *testing.T) {
	// Users will have identical values, but alice's bet is earliest, then bob, then carol
	userBetOffsets := map[string]time.Duration{
		"alice": 0,
		"bob":   1 * time.Minute,
		"carol": 2 * time.Minute,
	}

	earliest := make(map[string]time.Time)
	base := time.Now()
	for user, offset := range userBetOffsets {
		earliest[user] = base.Add(offset)
	}

	// All users have a rounded value of 10
	userVals := map[string]UserValuationResult{
		"alice": {Username: "alice", RoundedValue: 10},
		"bob":   {Username: "bob", RoundedValue: 10},
		"carol": {Username: "carol", RoundedValue: 10},
	}

	// Delta: need to add 2 (should go to alice then bob, since they are first by earliest bet)
	targetVolume := int64(32)
	adjusted := AdjustUserValuationsToMarketVolume(userVals, earliest, targetVolume)
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
	adjusted = AdjustUserValuationsToMarketVolume(userVals, earliest, targetVolume)
	want = map[string]int64{"alice": 9, "bob": 9, "carol": 10}
	for user, exp := range want {
		if adjusted[user].RoundedValue != exp {
			t.Errorf("user %s: want %d, got %d", user, exp, adjusted[user].RoundedValue)
		}
	}
}
