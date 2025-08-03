package sellbetshandlers

import (
	positionsmath "socialpredict/handlers/math/positions"
	"socialpredict/models/modelstesting"
	"testing"
)

func TestCalculateSharesToSell_DustCapValidation(t *testing.T) {
	cfg := modelstesting.GenerateEconomicConfig() // Has MaxDustPerSale: 2

	tests := []struct {
		name          string
		userValue     int64
		sharesOwned   int64
		creditsToSell int64
		maxDustCap    int64
		expectError   bool
		errorType     string
		expectedDust  int64
	}{
		{
			name:          "dust within cap - allowed",
			userValue:     100,
			sharesOwned:   10,
			creditsToSell: 22, // valuePerShare=10, sharesToSell=2, actualSale=20, dust=2
			maxDustCap:    2,
			expectError:   false,
			expectedDust:  2,
		},
		{
			name:          "dust exactly at cap - allowed",
			userValue:     100,
			sharesOwned:   10,
			creditsToSell: 12, // valuePerShare=10, sharesToSell=1, actualSale=10, dust=2
			maxDustCap:    2,
			expectError:   false,
			expectedDust:  2,
		},
		{
			name:          "dust exceeds cap - rejected",
			userValue:     100,
			sharesOwned:   10,
			creditsToSell: 33, // valuePerShare=10, sharesToSell=3, actualSale=30, dust=3
			maxDustCap:    2,
			expectError:   true,
			errorType:     "ErrDustCapExceeded",
			expectedDust:  3,
		},
		{
			name:          "no dust - always allowed",
			userValue:     100,
			sharesOwned:   10,
			creditsToSell: 30, // valuePerShare=10, sharesToSell=3, actualSale=30, dust=0
			maxDustCap:    2,
			expectError:   false,
			expectedDust:  0,
		},
		{
			name:          "dust cap disabled (0) - all dust allowed",
			userValue:     100,
			sharesOwned:   10,
			creditsToSell: 99, // valuePerShare=10, sharesToSell=9, actualSale=90, dust=9
			maxDustCap:    0,  // Disabled
			expectError:   false,
			expectedDust:  9,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create test position
			userPosition := positionsmath.UserMarketPosition{
				Value: test.userValue,
			}

			// Update config with test-specific dust cap
			testCfg := *cfg
			testCfg.Economics.Betting.MaxDustPerSale = test.maxDustCap

			// Test the function
			sharesToSell, actualSaleValue, err := calculateSharesToSell(
				userPosition, test.sharesOwned, test.creditsToSell, &testCfg)

			if test.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				// Check if it's the right error type
				if dustErr, ok := err.(ErrDustCapExceeded); ok {
					if dustErr.Requested != test.expectedDust {
						t.Errorf("expected dust %d, got %d", test.expectedDust, dustErr.Requested)
					}
					if dustErr.Cap != test.maxDustCap {
						t.Errorf("expected cap %d, got %d", test.maxDustCap, dustErr.Cap)
					}
				} else {
					t.Errorf("expected ErrDustCapExceeded, got %T: %v", err, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				// Verify dust calculation
				actualDust := test.creditsToSell - actualSaleValue
				if actualDust != test.expectedDust {
					t.Errorf("expected dust %d, got %d", test.expectedDust, actualDust)
				}

				// Verify shares calculation makes sense
				expectedValuePerShare := test.userValue / test.sharesOwned
				expectedShares := test.creditsToSell / expectedValuePerShare
				if expectedShares > test.sharesOwned {
					expectedShares = test.sharesOwned
				}

				if sharesToSell != expectedShares {
					t.Errorf("expected shares to sell %d, got %d", expectedShares, sharesToSell)
				}
			}
		})
	}
}

func TestCalculateSharesToSell_EdgeCases(t *testing.T) {
	cfg := modelstesting.GenerateEconomicConfig()

	tests := []struct {
		name          string
		userValue     int64
		sharesOwned   int64
		creditsToSell int64
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "zero position value",
			userValue:     0,
			sharesOwned:   10,
			creditsToSell: 50,
			expectError:   true,
			errorMsg:      "position value is non-positive",
		},
		{
			name:          "negative position value",
			userValue:     -100,
			sharesOwned:   10,
			creditsToSell: 50,
			expectError:   true,
			errorMsg:      "position value is non-positive",
		},
		{
			name:          "credits less than value per share",
			userValue:     100,
			sharesOwned:   10, // valuePerShare = 10
			creditsToSell: 5,  // Less than 10
			expectError:   true,
			errorMsg:      "requested credit amount is less than value of one share",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPosition := positionsmath.UserMarketPosition{
				Value: test.userValue,
			}

			_, _, err := calculateSharesToSell(
				userPosition, test.sharesOwned, test.creditsToSell, cfg)

			if test.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != test.errorMsg {
					t.Errorf("expected error %q, got %q", test.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestErrDustCapExceeded_ErrorInterface(t *testing.T) {
	err := ErrDustCapExceeded{
		Cap:       2,
		Requested: 5,
	}

	expectedMsg := "dust cap exceeded: would generate 5 dust points (cap: 2)"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}

	// Test business rule identification
	if !err.IsBusinessRuleError() {
		t.Error("ErrDustCapExceeded should be identified as a business rule error")
	}
}
