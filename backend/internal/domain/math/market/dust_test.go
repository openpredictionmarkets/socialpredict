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
	return makeDustTestBetWithDust(amount, 0, outcome, username, minutes)
}

func makeDustTestBetWithDust(amount int64, dust int64, outcome, username string, minutes int) boundary.Bet {
	placedAt := dustTestBaseTime.Add(time.Duration(minutes) * time.Minute)
	return boundary.Bet{
		Username:     username,
		MarketID:     1,
		Amount:       amount,
		Dust:         dust,
		DustRecorded: true,
		Outcome:      outcome,
		PlacedAt:     placedAt,
		CreatedAt:    placedAt,
	}
}

func makeLegacyDustTestBet(amount int64, outcome, username string, minutes int) boundary.Bet {
	bet := makeDustTestBetWithDust(amount, 0, outcome, username, minutes)
	bet.DustRecorded = false
	return bet
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
				makeDustTestBetWithDust(-50, 3, "YES", "user1", 2), // Sell
				makeDustTestBetWithDust(-75, 2, "NO", "user2", 3),  // Sell
			},
			expectedBase:     175, // 100 + 200 - 50 - 75
			expectedWithDust: 180, // Base + exact recorded sale dust
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
				makeDustTestBetWithDust(-50, 3, "YES", "user1", 0),
				makeDustTestBetWithDust(-75, 2, "NO", "user2", 1),
			},
			expectedDust: 5,
		},
		{
			name: "mixed bets - chronological order matters",
			bets: []boundary.Bet{
				makeDustTestBet(100, "YES", "user1", 0),            // Buy first
				makeDustTestBetWithDust(-50, 3, "YES", "user1", 1), // Sell second
				makeDustTestBet(200, "NO", "user2", 2),             // Buy third
				makeDustTestBetWithDust(-75, 2, "NO", "user2", 3),  // Sell fourth
			},
			expectedDust: 5,
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
		makeDustTestBetWithDust(-50, 3, "YES", "user1", 1),
		makeDustTestBetWithDust(-25, 2, "NO", "user2", 2),
	}

	assertInt64Equal(t, "dust", 5, GetMarketDust(bets))
}

func TestGetMarketDustFallsBackForLegacySellBets(t *testing.T) {
	bets := []boundary.Bet{
		makeDustTestBet(100, "YES", "user1", 0),
		makeLegacyDustTestBet(-50, "YES", "user1", 1),
		makeLegacyDustTestBet(-25, "NO", "user2", 2),
	}

	assertInt64Equal(t, "dust", 2, GetMarketDust(bets))
}

func TestGetMarketDustDoesNotFallbackForRecordedZeroDustSell(t *testing.T) {
	bets := []boundary.Bet{
		makeDustTestBet(10, "YES", "user1", 0),
		makeDustTestBet(1, "YES", "user1", 1),
		makeDustTestBetWithDust(-9, 0, "YES", "user1", 2),
	}

	assertInt64Equal(t, "dust", 0, GetMarketDust(bets))
	assertInt64Equal(t, "volume with dust", 2, GetMarketVolumeWithDust(bets))
}

func TestMarketDustScenarioMatrix(t *testing.T) {
	tests := []struct {
		name               string
		bets               []boundary.Bet
		wantBaseVolume     int64
		wantMarketDust     int64
		wantVolumeWithDust int64
	}{
		{
			name:               "no bets",
			bets:               []boundary.Bet{},
			wantBaseVolume:     0,
			wantMarketDust:     0,
			wantVolumeWithDust: 0,
		},
		{
			name: "only buys produce no dust",
			bets: []boundary.Bet{
				makeDustTestBet(10, "YES", "alice", 0),
				makeDustTestBet(5, "NO", "bob", 1),
			},
			wantBaseVolume:     15,
			wantMarketDust:     0,
			wantVolumeWithDust: 15,
		},
		{
			name: "recorded zero-dust sale does not fallback",
			bets: []boundary.Bet{
				makeDustTestBet(10, "YES", "alice", 0),
				makeDustTestBet(1, "YES", "alice", 1),
				makeDustTestBetWithDust(-9, 0, "YES", "alice", 2),
			},
			wantBaseVolume:     2,
			wantMarketDust:     0,
			wantVolumeWithDust: 2,
		},
		{
			name: "single recorded one-dust sale adds one",
			bets: []boundary.Bet{
				makeDustTestBet(10, "YES", "alice", 0),
				makeDustTestBetWithDust(-9, 1, "YES", "alice", 1),
			},
			wantBaseVolume:     1,
			wantMarketDust:     1,
			wantVolumeWithDust: 2,
		},
		{
			name: "single recorded two-dust sale adds two",
			bets: []boundary.Bet{
				makeDustTestBet(10, "YES", "alice", 0),
				makeDustTestBetWithDust(-8, 2, "YES", "alice", 1),
			},
			wantBaseVolume:     2,
			wantMarketDust:     2,
			wantVolumeWithDust: 4,
		},
		{
			name: "mixed users and outcomes sum recorded sale dust",
			bets: []boundary.Bet{
				makeDustTestBet(100, "YES", "alice", 0),
				makeDustTestBet(70, "NO", "bob", 1),
				makeDustTestBetWithDust(-40, 2, "YES", "alice", 2),
				makeDustTestBetWithDust(-30, 1, "NO", "bob", 3),
			},
			wantBaseVolume:     100,
			wantMarketDust:     3,
			wantVolumeWithDust: 103,
		},
		{
			name: "multiple sells by one user include zero and nonzero dust",
			bets: []boundary.Bet{
				makeDustTestBet(50, "YES", "alice", 0),
				makeDustTestBetWithDust(-10, 0, "YES", "alice", 1),
				makeDustTestBetWithDust(-15, 2, "YES", "alice", 2),
			},
			wantBaseVolume:     25,
			wantMarketDust:     2,
			wantVolumeWithDust: 27,
		},
		{
			name: "legacy null-dust sells still use fallback",
			bets: []boundary.Bet{
				makeDustTestBet(50, "YES", "alice", 0),
				makeLegacyDustTestBet(-10, "YES", "alice", 1),
				makeLegacyDustTestBet(-5, "YES", "alice", 2),
			},
			wantBaseVolume:     35,
			wantMarketDust:     2,
			wantVolumeWithDust: 37,
		},
		{
			name: "legacy and recorded sells combine",
			bets: []boundary.Bet{
				makeDustTestBet(80, "YES", "alice", 0),
				makeLegacyDustTestBet(-10, "YES", "alice", 1),
				makeDustTestBetWithDust(-20, 3, "YES", "alice", 2),
			},
			wantBaseVolume:     50,
			wantMarketDust:     4,
			wantVolumeWithDust: 54,
		},
		{
			name: "recorded zero-dust sale after nonzero sale keeps total stable",
			bets: []boundary.Bet{
				makeDustTestBet(40, "YES", "alice", 0),
				makeDustTestBetWithDust(-8, 1, "YES", "alice", 1),
				makeDustTestBetWithDust(-7, 0, "YES", "alice", 2),
				makeDustTestBet(15, "NO", "bob", 3),
			},
			wantBaseVolume:     40,
			wantMarketDust:     1,
			wantVolumeWithDust: 41,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertInt64Equal(t, "base volume", tc.wantBaseVolume, GetMarketVolume(tc.bets))
			assertInt64Equal(t, "market dust", tc.wantMarketDust, GetMarketDust(tc.bets))
			assertInt64Equal(t, "volume with dust", tc.wantVolumeWithDust, GetMarketVolumeWithDust(tc.bets))
		})
	}
}

func TestGetMarketDustWithCalculator(t *testing.T) {
	bets := []boundary.Bet{
		makeDustTestBet(100, "YES", "user1", 0),
		makeLegacyDustTestBet(-50, "YES", "user1", 1),
		makeLegacyDustTestBet(-25, "NO", "user2", 2),
	}

	assertInt64Equal(t, "dust", 6, GetMarketDustWithCalculator(bets, fixedSellDustCalculator{dust: 3}))
}

func TestCalculateDustForSell(t *testing.T) {
	sellBet := boundary.Bet{Amount: -50, Dust: 2, Outcome: "YES", Username: "user1", MarketID: 1, PlacedAt: dustTestBaseTime}
	buyBet := boundary.Bet{Amount: 100, Outcome: "YES", Username: "user1", MarketID: 1, PlacedAt: dustTestBaseTime}

	allBets := []boundary.Bet{buyBet, sellBet}

	assertInt64Equal(t, "sell dust", 2, calculateDustForSell(sellBet, allBets))
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
