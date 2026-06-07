package marketmath

import (
	"fmt"
	"testing"
	"time"

	"socialpredict/internal/domain/boundary"
)

type fixedSellDustCalculator struct{ dust int64 }

func (c fixedSellDustCalculator) DustForSell(sellBet boundary.Bet, allBets []boundary.Bet) int64 {
	if sellBet.Amount >= 0 {
		return 0
	}
	return c.dust
}

var dustTestBaseTime = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

func makeDustTestBet(amount int64, outcome, username string, minutes int) boundary.Bet {
	placedAt := dustTestBaseTime.Add(time.Duration(minutes) * time.Minute)
	return boundary.Bet{
		Username:  username,
		MarketID:  1,
		Amount:    amount,
		Outcome:   outcome,
		PlacedAt:  placedAt,
		CreatedAt: placedAt,
	}
}

func assertInt64Equal(t *testing.T, label string, want, got int64) {
	t.Helper()
	if got != want {
		t.Fatalf("expected %s %d, got %d", label, want, got)
	}
}

func assertDustInvariant(t *testing.T, bets []boundary.Bet) {
	t.Helper()
	baseVolume := GetMarketVolume(bets)
	volumeWithDust := GetMarketVolumeWithDust(bets)
	if volumeWithDust < baseVolume {
		t.Fatalf("currency conservation violated: volumeWithDust (%d) < baseVolume (%d)", volumeWithDust, baseVolume)
	}

	dust := volumeWithDust - baseVolume
	if dust < 0 {
		t.Fatalf("dust cannot be negative, got %d", dust)
	}
}

func TestGetMarketVolumeWithDust(t *testing.T) {
	tests := []struct {
		name             string
		bets             []boundary.Bet
		expectedBase     int64
		expectedWithDust int64
	}{
		{
			name:             "empty market",
			bets:             []boundary.Bet{},
			expectedBase:     0,
			expectedWithDust: 0,
		},
		{
			name: "only buy bets - no dust",
			bets: []boundary.Bet{
				makeDustTestBet(100, "YES", "user1", 0),
				makeDustTestBet(200, "NO", "user2", 1),
			},
			expectedBase:     300,
			expectedWithDust: 300, // No sells, so no dust
		},
		{
			name: "mixed buys and sells - with dust",
			bets: []boundary.Bet{
				makeDustTestBet(100, "YES", "user1", 0),
				makeDustTestBet(200, "NO", "user2", 1),
				makeDustTestBet(-50, "YES", "user1", 2), // Sell
				makeDustTestBet(-75, "NO", "user2", 3),  // Sell
			},
			expectedBase:     175, // 100 + 200 - 50 - 75
			expectedWithDust: 177, // Base + 2 dust (1 per sell as placeholder)
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertInt64Equal(t, "base volume", test.expectedBase, GetMarketVolume(test.bets))
			assertInt64Equal(t, "volume with dust", test.expectedWithDust, GetMarketVolumeWithDust(test.bets))
		})
	}
}

func TestCalculateDustStack(t *testing.T) {
	tests := []struct {
		name         string
		bets         []boundary.Bet
		expectedDust int64
	}{
		{
			name:         "no bets",
			bets:         []boundary.Bet{},
			expectedDust: 0,
		},
		{
			name: "only buy bets",
			bets: []boundary.Bet{
				makeDustTestBet(100, "YES", "user1", 0),
				makeDustTestBet(200, "NO", "user2", 1),
			},
			expectedDust: 0,
		},
		{
			name: "only sell bets",
			bets: []boundary.Bet{
				makeDustTestBet(-50, "YES", "user1", 0),
				makeDustTestBet(-75, "NO", "user2", 1),
			},
			expectedDust: 2, // 1 dust per sell (placeholder implementation)
		},
		{
			name: "mixed bets - chronological order matters",
			bets: []boundary.Bet{
				makeDustTestBet(100, "YES", "user1", 0), // Buy first
				makeDustTestBet(-50, "YES", "user1", 1), // Sell second
				makeDustTestBet(200, "NO", "user2", 2),  // Buy third
				makeDustTestBet(-75, "NO", "user2", 3),  // Sell fourth
			},
			expectedDust: 2, // 2 sells = 2 dust points
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertInt64Equal(t, "dust", test.expectedDust, calculateDustStack(test.bets))
		})
	}
}

func TestGetMarketDust(t *testing.T) {
	bets := []boundary.Bet{
		makeDustTestBet(100, "YES", "user1", 0),
		makeDustTestBet(-50, "YES", "user1", 1),
		makeDustTestBet(-25, "NO", "user2", 2),
	}

	assertInt64Equal(t, "dust", 2, GetMarketDust(bets))
}

func TestGetMarketDustWithCalculator(t *testing.T) {
	bets := []boundary.Bet{
		makeDustTestBet(100, "YES", "user1", 0),
		makeDustTestBet(-50, "YES", "user1", 1),
		makeDustTestBet(-25, "NO", "user2", 2),
	}

	assertInt64Equal(t, "dust", 6, GetMarketDustWithCalculator(bets, fixedSellDustCalculator{dust: 3}))
}

func TestCalculateDustForSell_Placeholder(t *testing.T) {
	sellBet := boundary.Bet{Amount: -50, Outcome: "YES", Username: "user1", MarketID: 1, PlacedAt: dustTestBaseTime}
	buyBet := boundary.Bet{Amount: 100, Outcome: "YES", Username: "user1", MarketID: 1, PlacedAt: dustTestBaseTime}

	allBets := []boundary.Bet{buyBet, sellBet}

	assertInt64Equal(t, "sell dust", 1, calculateDustForSell(sellBet, allBets))
	assertInt64Equal(t, "buy dust", 0, calculateDustForSell(buyBet, allBets))
}

func TestCurrencyConservationInvariant(t *testing.T) {
	testCases := [][]boundary.Bet{
		{}, // Empty
		{makeDustTestBet(100, "YES", "user1", 0)}, // Only buys
		{makeDustTestBet(-50, "YES", "user1", 0)}, // Only sells
		{ // Mixed
			makeDustTestBet(100, "YES", "user1", 0),
			makeDustTestBet(-50, "YES", "user1", 1),
			makeDustTestBet(200, "NO", "user2", 2),
			makeDustTestBet(-75, "NO", "user2", 3),
		},
	}

	for i, bets := range testCases {
		t.Run(fmt.Sprintf("invariant_test_%d", i), func(t *testing.T) {
			assertDustInvariant(t, bets)
		})
	}
}
