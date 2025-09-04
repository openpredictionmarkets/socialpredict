package marketshandlers

import (
	"socialpredict/setup"
	"strings"
	"testing"
	"time"
)

// TestCheckQuestionTitleLength_invalid tests the question titles that should generate an error
func TestCheckQuestionTitleLength_invalid(t *testing.T) {
	tests := []struct {
		testname      string
		questionTitle string
	}{
		{
			testname:      "TitleExceedsLength",
			questionTitle: strings.Repeat("a", maxQuestionTitleLength+1),
		},
		{
			testname:      "EmptyTitle",
			questionTitle: "",
		},
	}
	for _, test := range tests {
		err := checkQuestionTitleLength(test.questionTitle)
		if err == nil {
			t.Errorf("Expected error in test %s", test.testname)
		}
	}
}

func TestCheckQuestionTitleLength_valid(t *testing.T) {
	tests := []struct {
		testname      string
		questionTitle string
	}{
		{
			testname:      "Single character title",
			questionTitle: "a",
		},
		{
			testname:      "Max length title",
			questionTitle: strings.Repeat("a", maxQuestionTitleLength),
		},
	}
	for _, test := range tests {
		err := checkQuestionTitleLength(test.questionTitle)
		if err != nil {
			t.Errorf("Unexpected error in test %s", test.testname)
		}
	}
}

// TestValidateMarketResolutionTime tests the business logic validation for market resolution times
func TestValidateMarketResolutionTime(t *testing.T) {
	// Create test config with 1.0 hour minimum future time
	config := &setup.EconomicConfig{
		Economics: setup.Economics{
			MarketCreation: setup.MarketCreation{
				MinimumFutureHours: 1.0,
			},
		},
	}

	tests := []struct {
		name           string
		resolutionTime time.Time
		expectedError  bool
		errorContains  string
	}{
		{
			name:           "Market resolving in past should be rejected",
			resolutionTime: time.Now().Add(-24 * time.Hour),
			expectedError:  true,
			errorContains:  "must be at least 1.0 hours in the future",
		},
		{
			name:           "Market resolving exactly now should be rejected",
			resolutionTime: time.Now(),
			expectedError:  true,
			errorContains:  "must be at least 1.0 hours in the future",
		},
		{
			name:           "Market resolving in 30 minutes should be rejected",
			resolutionTime: time.Now().Add(30 * time.Minute),
			expectedError:  true,
			errorContains:  "must be at least 1.0 hours in the future",
		},
		{
			name:           "Market resolving in exactly 1 hour should be rejected",
			resolutionTime: time.Now().Add(1 * time.Hour),
			expectedError:  true,
			errorContains:  "must be at least 1.0 hours in the future",
		},
		{
			name:           "Market resolving in 1.1 hours should be accepted",
			resolutionTime: time.Now().Add(66 * time.Minute),
			expectedError:  false,
		},
		{
			name:           "Market resolving in 24 hours should be accepted",
			resolutionTime: time.Now().Add(24 * time.Hour),
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMarketResolutionTime(tt.resolutionTime, config)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but validation passed")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s' but got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected validation to pass but got error: %v", err)
				}
			}
		})
	}
}

// TestValidateMarketResolutionTimeCustomConfig tests the business logic with different configurations
func TestValidateMarketResolutionTimeCustomConfig(t *testing.T) {
	tests := []struct {
		name               string
		minimumFutureHours float64
		testTime           time.Duration
		expectedError      bool
	}{
		{
			name:               "0.5 hour minimum - 20 minutes should fail",
			minimumFutureHours: 0.5,
			testTime:           20 * time.Minute,
			expectedError:      true,
		},
		{
			name:               "0.5 hour minimum - 40 minutes should pass",
			minimumFutureHours: 0.5,
			testTime:           40 * time.Minute,
			expectedError:      false,
		},
		{
			name:               "2.0 hour minimum - 1.5 hours should fail",
			minimumFutureHours: 2.0,
			testTime:           90 * time.Minute,
			expectedError:      true,
		},
		{
			name:               "2.0 hour minimum - 2.5 hours should pass",
			minimumFutureHours: 2.0,
			testTime:           150 * time.Minute,
			expectedError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &setup.EconomicConfig{
				Economics: setup.Economics{
					MarketCreation: setup.MarketCreation{
						MinimumFutureHours: tt.minimumFutureHours,
					},
				},
			}

			resolutionTime := time.Now().Add(tt.testTime)
			err := validateMarketResolutionTime(resolutionTime, config)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but validation passed")
				}
			} else {
				if err != nil {
					t.Errorf("Expected validation to pass but got error: %v", err)
				}
			}
		})
	}
}
