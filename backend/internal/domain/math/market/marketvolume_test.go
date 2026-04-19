package marketmath

import (
	"testing"

	"socialpredict/internal/domain/boundary"
)

type stubVolumeCalculator struct {
	volume    int64
	endVolume int64
}

func (s stubVolumeCalculator) Volume([]boundary.Bet) int64 { return s.volume }
func (s stubVolumeCalculator) EndVolume([]boundary.Bet, int64) int64 {
	return s.endVolume
}

func assertMarketVolume(t *testing.T, bets []boundary.Bet, want int64) {
	t.Helper()
	if got := GetMarketVolume(bets); got != want {
		t.Fatalf("expected volume %d, got %d", want, got)
	}
}

// TestGetMarketVolume tests that the total volume of trades is returned correctly
func TestGetMarketVolume(t *testing.T) {
	tests := []struct {
		Name           string
		Bets           []boundary.Bet
		ExpectedVolume int64
	}{
		{
			Name:           "empty slice",
			Bets:           []boundary.Bet{},
			ExpectedVolume: 0,
		},
		{
			Name: "positive Bets",
			Bets: []boundary.Bet{
				{Amount: 100},
				{Amount: 200},
				{Amount: 300},
			},
			ExpectedVolume: 600,
		},
		{
			Name: "negative Bets",
			Bets: []boundary.Bet{
				{Amount: -100},
				{Amount: -200},
				{Amount: -300},
			},
			ExpectedVolume: -600,
		},
		{
			Name: "mixed Bets",
			Bets: []boundary.Bet{
				{Amount: 100},
				{Amount: -50},
				{Amount: 200},
				{Amount: -150},
			},
			ExpectedVolume: 100,
		},
		{
			Name: "large numbers",
			Bets: []boundary.Bet{
				{Amount: 9223372036854775807},
				{Amount: -9223372036854775806},
			},
			ExpectedVolume: 1,
		},
		{
			Name: "infinity avoidance",
			Bets: []boundary.Bet{
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
