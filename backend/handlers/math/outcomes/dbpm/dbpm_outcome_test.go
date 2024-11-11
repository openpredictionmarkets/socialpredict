package dbpm

import (
	"socialpredict/logging"
	"socialpredict/models"
	"testing"
)

func TestCalculatePayoutForOutcomeDBPM(t *testing.T) {
	tests := []struct {
		name              string
		bet               models.Bet
		totalYes, totalNo int64
		betInput          string
		marketResolution  string
		expectedPayout    float64
	}{
		{
			name: "Winning YES bet",
			bet: models.Bet{
				Amount: 50,
			},
			totalYes:         100,
			totalNo:          50,
			betInput:         "YES",
			marketResolution: "YES",
			expectedPayout:   75,
		},
		{
			name: "Winning NO bet",
			bet: models.Bet{
				Amount: 30,
			},
			totalYes:         100,
			totalNo:          70,
			betInput:         "NO",
			marketResolution: "NO",
			expectedPayout:   72.85714285714285,
		},
		{
			name: "Losing YES bet",
			bet: models.Bet{
				Amount: 40,
			},
			totalYes:         100,
			totalNo:          50,
			betInput:         "YES",
			marketResolution: "NO",
			expectedPayout:   0,
		},
		{
			name: "Losing NO bet",
			bet: models.Bet{
				Amount: 25,
			},
			totalYes:         80,
			totalNo:          120,
			betInput:         "NO",
			marketResolution: "YES",
			expectedPayout:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payout := CalculatePayoutForOutcomeDBPM(tt.bet, tt.totalYes, tt.totalNo, tt.betInput, tt.marketResolution)
			logging.LogAnyType(payout, "payout")
			if payout != tt.expectedPayout {
				t.Errorf("%s: expected payout %f, got %f", tt.name, tt.expectedPayout, payout)
			}
		})
	}
}
