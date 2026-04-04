package marketmath

import (
	"fmt"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"testing"
	"time"
)

type marketVolumeDustExpectation struct {
	base     int64
	withDust int64
}

func assertMarketVolumeWithDust(t *testing.T, bets []models.Bet, expected marketVolumeDustExpectation) {
	t.Helper()

	baseVolume := GetMarketVolume(bets)
	if baseVolume != expected.base {
		t.Errorf("expected base volume %d, got %d", expected.base, baseVolume)
	}

	volumeWithDust := GetMarketVolumeWithDust(bets)
	if volumeWithDust != expected.withDust {
		t.Errorf("expected volume with dust %d, got %d", expected.withDust, volumeWithDust)
	}
}

func assertDustAmount(t *testing.T, bets []models.Bet, expectedDust int64) {
	t.Helper()

	dust := calculateDustStack(bets)
	if dust != expectedDust {
		t.Errorf("expected dust %d, got %d", expectedDust, dust)
	}
}

func TestGetMarketVolumeWithDust(t *testing.T) {
	tests := []struct {
		name     string
		bets     []models.Bet
		expected marketVolumeDustExpectation
	}{
		{
			name:     "empty market",
			bets:     []models.Bet{},
			expected: marketVolumeDustExpectation{base: 0, withDust: 0},
		},
		{
			name: "only buy bets - no dust",
			bets: []models.Bet{
				modelstesting.GenerateBet(100, "YES", "user1", 1, 0),
				modelstesting.GenerateBet(200, "NO", "user2", 1, time.Minute),
			},
			expected: marketVolumeDustExpectation{base: 300, withDust: 300},
		},
		{
			name: "mixed buys and sells - with dust",
			bets: []models.Bet{
				modelstesting.GenerateBet(100, "YES", "user1", 1, 0),
				modelstesting.GenerateBet(200, "NO", "user2", 1, time.Minute),
				modelstesting.GenerateBet(-50, "YES", "user1", 1, 2*time.Minute),
				modelstesting.GenerateBet(-75, "NO", "user2", 1, 3*time.Minute),
			},
			expected: marketVolumeDustExpectation{base: 175, withDust: 177},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertMarketVolumeWithDust(t, test.bets, test.expected)
		})
	}
}

func TestCalculateDustStack(t *testing.T) {
	tests := []struct {
		name         string
		bets         []models.Bet
		expectedDust int64
	}{
		{
			name:         "no bets",
			bets:         []models.Bet{},
			expectedDust: 0,
		},
		{
			name: "only buy bets",
			bets: []models.Bet{
				modelstesting.GenerateBet(100, "YES", "user1", 1, 0),
				modelstesting.GenerateBet(200, "NO", "user2", 1, time.Minute),
			},
			expectedDust: 0,
		},
		{
			name: "only sell bets",
			bets: []models.Bet{
				modelstesting.GenerateBet(-50, "YES", "user1", 1, 0),
				modelstesting.GenerateBet(-75, "NO", "user2", 1, time.Minute),
			},
			expectedDust: 2, // 1 dust per sell (placeholder implementation)
		},
		{
			name: "mixed bets - chronological order matters",
			bets: []models.Bet{
				modelstesting.GenerateBet(100, "YES", "user1", 1, 0),            // Buy first
				modelstesting.GenerateBet(-50, "YES", "user1", 1, time.Minute),  // Sell second
				modelstesting.GenerateBet(200, "NO", "user2", 1, 2*time.Minute), // Buy third
				modelstesting.GenerateBet(-75, "NO", "user2", 1, 3*time.Minute), // Sell fourth
			},
			expectedDust: 2, // 2 sells = 2 dust points
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertDustAmount(t, test.bets, test.expectedDust)
		})
	}
}

func TestGetMarketDust(t *testing.T) {
	bets := []models.Bet{
		modelstesting.GenerateBet(100, "YES", "user1", 1, 0),
		modelstesting.GenerateBet(-50, "YES", "user1", 1, time.Minute),
		modelstesting.GenerateBet(-25, "NO", "user2", 1, 2*time.Minute),
	}

	dust := GetMarketDust(bets)
	expectedDust := int64(2)

	if dust != expectedDust {
		t.Errorf("expected dust %d, got %d", expectedDust, dust)
	}
}

func TestCalculateDustForSell_Placeholder(t *testing.T) {
	sellBet := modelstesting.GenerateBet(-50, "YES", "user1", 1, 0)
	buyBet := modelstesting.GenerateBet(100, "YES", "user1", 1, 0)
	allBets := []models.Bet{buyBet, sellBet}
	tests := []struct {
		name     string
		bet      models.Bet
		expected int64
	}{
		{name: "sell bet", bet: sellBet, expected: 1},
		{name: "buy bet", bet: buyBet, expected: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dust := calculateDustForSell(test.bet, allBets)
			if dust != test.expected {
				t.Errorf("expected %d dust, got %d", test.expected, dust)
			}
		})
	}
}

func TestCurrencyConservationInvariant(t *testing.T) {
	testCases := [][]models.Bet{
		{},
		{modelstesting.GenerateBet(100, "YES", "user1", 1, 0)},
		{modelstesting.GenerateBet(-50, "YES", "user1", 1, 0)},
		{
			modelstesting.GenerateBet(100, "YES", "user1", 1, 0),
			modelstesting.GenerateBet(-50, "YES", "user1", 1, time.Minute),
			modelstesting.GenerateBet(200, "NO", "user2", 1, 2*time.Minute),
			modelstesting.GenerateBet(-75, "NO", "user2", 1, 3*time.Minute),
		},
	}

	for i, bets := range testCases {
		t.Run(fmt.Sprintf("invariant_test_%d", i), func(t *testing.T) {
			baseVolume := GetMarketVolume(bets)
			volumeWithDust := GetMarketVolumeWithDust(bets)

			if volumeWithDust < baseVolume {
				t.Errorf("currency conservation violated: volumeWithDust (%d) < baseVolume (%d)",
					volumeWithDust, baseVolume)
			}

			dust := volumeWithDust - baseVolume
			if dust < 0 {
				t.Errorf("dust cannot be negative, got %d", dust)
			}
		})
	}
}
