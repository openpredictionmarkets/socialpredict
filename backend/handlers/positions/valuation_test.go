package positions

import (
	"testing"
)

func TestCalculateRoundedUserValuationsFromUserMarketPositions(t *testing.T) {
	userPositions := map[string]UserMarketPosition{
		"alice": {YesSharesOwned: 100, NoSharesOwned: 0},
		"bob":   {YesSharesOwned: 0, NoSharesOwned: 50},
	}

	currentProb := 0.6
	totalVolume := int64(130) // deliberately mismatched from exact float value (should be ~130 total value)

	valuations := CalculateRoundedUserValuationsFromUserMarketPositions(userPositions, currentProb, totalVolume)

	if len(valuations) != 2 {
		t.Fatalf("expected 2 valuations, got %d", len(valuations))
	}

	alice := valuations["alice"]
	bob := valuations["bob"]

	// Alice: 100 * 0.6 = 60
	// Bob: 50 * 0.4 = 20
	// Total = 80 → delta of 50 added to alice as largest floatVal
	if alice.RoundedValue+bob.RoundedValue != totalVolume {
		t.Errorf("total valuation mismatch: got %d, expected %d", alice.RoundedValue+bob.RoundedValue, totalVolume)
	}

	if alice.RoundedValue != 110 {
		t.Errorf("expected alice to be adjusted to 110, got %d", alice.RoundedValue)
	}
	if bob.RoundedValue != 20 {
		t.Errorf("expected bob to be 20, got %d", bob.RoundedValue)
	}
}
