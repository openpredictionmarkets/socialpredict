package marketmath

import (
	"socialpredict/models"
	"testing"
)

type stubVolumeCalculator struct {
	volume    int64
	endVolume int64
}

func (s stubVolumeCalculator) Volume([]models.Bet) int64 { return s.volume }
func (s stubVolumeCalculator) EndVolume([]models.Bet, int64) int64 {
	return s.endVolume
}

func assertMarketVolume(t *testing.T, bets []models.Bet, want int64) {
	t.Helper()
	if got := GetMarketVolume(bets); got != want {
		t.Fatalf("expected volume %d, got %d", want, got)
	}
}

// TestGetMarketVolume tests that the total volume of trades is returned correctly
func TestGetMarketVolume(t *testing.T) {
	tests := []struct {
		Name           string
		Bets           []models.Bet
		ExpectedVolume int64
	}{
		{
			Name:           "empty slice",
			Bets:           []models.Bet{},
			ExpectedVolume: 0,
		},
		{
			Name: "positive Bets",
			Bets: []models.Bet{
				{Amount: 100},
				{Amount: 200},
				{Amount: 300},
			},
			ExpectedVolume: 600,
		},
		{
			Name: "negative Bets",
			Bets: []models.Bet{
				{Amount: -100},
				{Amount: -200},
				{Amount: -300},
			},
			ExpectedVolume: -600,
		},
		{
			Name: "mixed Bets",
			Bets: []models.Bet{
				{Amount: 100},
				{Amount: -50},
				{Amount: 200},
				{Amount: -150},
			},
			ExpectedVolume: 100,
		},
		{
			Name: "large numbers",
			Bets: []models.Bet{
				{Amount: 9223372036854775807},
				{Amount: -9223372036854775806},
			},
			ExpectedVolume: 1,
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
			ExpectedVolume: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			assertMarketVolume(t, test.Bets, test.ExpectedVolume)
		})
	}
}

func TestGetMarketVolume_UsesInjectedCalculator(t *testing.T) {
	originalVolumeCalculator := defaultVolumeCalculator
	originalEndVolumeCalculator := defaultEndVolumeCalculator
	defaultVolumeCalculator = stubVolumeCalculator{volume: 41, endVolume: 99}
	defaultEndVolumeCalculator = stubVolumeCalculator{volume: 41, endVolume: 99}
	defer func() {
		defaultVolumeCalculator = originalVolumeCalculator
		defaultEndVolumeCalculator = originalEndVolumeCalculator
	}()

	if got := GetMarketVolume(nil); got != 41 {
		t.Fatalf("expected injected volume 41, got %d", got)
	}
	if got := GetEndMarketVolume(nil, 12); got != 99 {
		t.Fatalf("expected injected end volume 99, got %d", got)
	}
}
