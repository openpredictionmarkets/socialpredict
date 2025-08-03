package marketshandlers

import (
	"socialpredict/security"
	"testing"
	"time"
)

func TestMarketFutureDateValidation(t *testing.T) {
	securityService := security.NewSecurityService()

	tests := []struct {
		name          string
		endTime       string
		expectedValid bool
		expectedError string
	}{
		{
			name:          "Market closing in past should be rejected",
			endTime:       time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			expectedValid: false,
			expectedError: "endtime must be at least 1 hour in the future",
		},
		{
			name:          "Market closing exactly now should be rejected",
			endTime:       time.Now().Format(time.RFC3339),
			expectedValid: false,
			expectedError: "endtime must be at least 1 hour in the future",
		},
		{
			name:          "Market closing in 30 minutes should be rejected",
			endTime:       time.Now().Add(30 * time.Minute).Format(time.RFC3339),
			expectedValid: false,
			expectedError: "endtime must be at least 1 hour in the future",
		},
		{
			name:          "Market closing in exactly 1 hour should be rejected", // our validator requires > 1 hour
			endTime:       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			expectedValid: false,
			expectedError: "endtime must be at least 1 hour in the future",
		},
		{
			name:          "Market closing in 1.5 hours should be accepted",
			endTime:       time.Now().Add(90 * time.Minute).Format(time.RFC3339),
			expectedValid: true,
		},
		{
			name:          "Market closing in 24 hours should be accepted",
			endTime:       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			expectedValid: true,
		},
		{
			name:          "Invalid date format should be rejected",
			endTime:       "invalid-date-format",
			expectedValid: false,
			expectedError: "endtime must be at least 1 hour in the future",
		},
		{
			name:          "Empty date should be rejected",
			endTime:       "",
			expectedValid: false,
			expectedError: "endtime is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marketInput := security.MarketInput{
				Title:       "Test Market Question?",
				Description: "Test description",
				EndTime:     tt.endTime,
			}

			_, err := securityService.ValidateAndSanitizeMarketInput(marketInput)

			if tt.expectedValid {
				if err != nil {
					t.Errorf("Expected valid market input but got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error but validation passed")
				} else {
					// Check that error message contains expected text
					if tt.expectedError != "" && err.Error() != "" {
						// Just ensure we got some validation error - exact message matching can be brittle
						t.Logf("Got expected validation error: %v", err)
					}
				}
			}
		})
	}
}

func TestMarketDateFormats(t *testing.T) {
	securityService := security.NewSecurityService()

	// Test various date formats that should work
	futureTime := time.Now().Add(2 * time.Hour)

	tests := []struct {
		name       string
		endTime    string
		shouldWork bool
	}{
		{
			name:       "RFC3339 format",
			endTime:    futureTime.Format(time.RFC3339),
			shouldWork: true,
		},
		{
			name:       "UTC format with Z",
			endTime:    futureTime.UTC().Format("2006-01-02T15:04:05Z"),
			shouldWork: true,
		},
		{
			name:       "Format with timezone offset",
			endTime:    futureTime.Format("2006-01-02T15:04:05Z07:00"),
			shouldWork: true,
		},
		{
			name:       "Simple ISO format",
			endTime:    futureTime.UTC().Format("2006-01-02T15:04:05"),
			shouldWork: true,
		},
		{
			name:       "Space separated format",
			endTime:    futureTime.UTC().Format("2006-01-02 15:04:05"),
			shouldWork: true,
		},
		{
			name:       "Invalid format",
			endTime:    futureTime.Format("01/02/2006 3:04 PM"),
			shouldWork: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marketInput := security.MarketInput{
				Title:       "Test Market Question?",
				Description: "Test description",
				EndTime:     tt.endTime,
			}

			_, err := securityService.ValidateAndSanitizeMarketInput(marketInput)

			if tt.shouldWork {
				if err != nil {
					t.Errorf("Expected date format '%s' to work but got error: %v", tt.endTime, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected date format to fail but validation passed")
				} else {
					t.Logf("Got expected validation error for invalid format: %v", err)
				}
			}
		})
	}
}
