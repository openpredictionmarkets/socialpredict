package marketmath

import (
	"socialpredict/models"
	"testing"
)

func assertMarketVolume(t *testing.T, bets []models.Bet, expected int64) {
	t.Helper()

	volume := GetMarketVolume(bets)
	if volume != expected {
		t.Errorf("expected %d, got %d", expected, volume)
	}
}

// TestGetMarketVolume tests that the total volume of trades is returned correctly
func TestGetMarketVolume(t *testing.T) {
	tests := []struct {
		Name     string
		Bets     []models.Bet
		Expected int64
	}{
		{
			Name:     "empty slice",
			Bets:     []models.Bet{},
			Expected: 0,
		},
		{
			Name: "positive Bets",
			Bets: []models.Bet{
				{Amount: 100},
				{Amount: 200},
				{Amount: 300},
			},
			Expected: 600,
		},
		{
			Name: "negative Bets",
			Bets: []models.Bet{
				{Amount: -100},
				{Amount: -200},
				{Amount: -300},
			},
			Expected: -600,
		},
		{
			Name: "mixed Bets",
			Bets: []models.Bet{
				{Amount: 100},
				{Amount: -50},
				{Amount: 200},
				{Amount: -150},
			},
			Expected: 100,
		},
		{
			Name: "large numbers",
			Bets: []models.Bet{
				{Amount: 9223372036854775807},
				{Amount: -9223372036854775806},
			},
			Expected: 1,
		},
		{
			Name: "infinity avoidance",
			Bets: []models.Bet{
				{Amount: 1, Outcome: "YES"},
				{Amount: -1, Outcome: "YES"},
				{Amount: 1, Outcome: "NO"},
				{Amount: -1, Outcome: "NO"},
				{Amount: 1, Outcome: "NO"},
			},
			Expected: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			assertMarketVolume(t, test.Bets, test.Expected)
		})
	}
}
