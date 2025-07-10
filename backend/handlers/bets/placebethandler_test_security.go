package betshandlers

import (
	"socialpredict/security"
	"testing"
)

func TestBetInputValidation(t *testing.T) {
	tests := []struct {
		name          string
		marketID      string
		amount        float64
		outcome       string
		expectedValid bool
		expectedError string
	}{
		{
			name:          "Valid bet input",
			marketID:      "123",
			amount:        100.0,
			outcome:       "YES",
			expectedValid: true,
		},
		{
			name:          "Valid NO bet",
			marketID:      "456",
			amount:        50.0,
			outcome:       "NO",
			expectedValid: true,
		},
		{
			name:          "Zero amount",
			marketID:      "123",
			amount:        0.0,
			outcome:       "YES",
			expectedValid: false,
			expectedError: "amount must be positive",
		},
		{
			name:          "Negative amount",
			marketID:      "123",
			amount:        -10.0,
			outcome:       "YES",
			expectedValid: false,
			expectedError: "amount must be positive",
		},
		{
			name:          "Invalid outcome",
			marketID:      "123",
			amount:        100.0,
			outcome:       "MAYBE",
			expectedValid: false,
			expectedError: "outcome must be YES or NO",
		},
		{
			name:          "Empty outcome",
			marketID:      "123",
			amount:        100.0,
			outcome:       "",
			expectedValid: false,
			expectedError: "outcome cannot be empty",
		},
		{
			name:          "Empty market ID",
			marketID:      "",
			amount:        100.0,
			outcome:       "YES",
			expectedValid: false,
			expectedError: "market ID cannot be empty",
		},
		{
			name:          "Very large amount",
			marketID:      "123",
			amount:        1000000.0,
			outcome:       "YES",
			expectedValid: true, // Should be allowed but may hit balance limits
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			securityService := security.NewSecurityService()

			betInput := security.BetInput{
				MarketID: tt.marketID,
				Amount:   tt.amount,
				Outcome:  tt.outcome,
			}

			_, err := securityService.ValidateAndSanitizeBetInput(betInput)

			if tt.expectedValid {
				if err != nil {
					t.Errorf("Expected valid bet input but got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error but bet validation passed")
				}
			}
		})
	}
}

func TestBetInputSanitization(t *testing.T) {
	securityService := security.NewSecurityService()

	tests := []struct {
		name     string
		input    security.BetInput
		expected security.BetInput
	}{
		{
			name: "Normal input",
			input: security.BetInput{
				MarketID: "123",
				Amount:   100.0,
				Outcome:  "YES",
			},
			expected: security.BetInput{
				MarketID: "123",
				Amount:   100.0,
				Outcome:  "YES",
			},
		},
		{
			name: "Outcome case normalization",
			input: security.BetInput{
				MarketID: "123",
				Amount:   100.0,
				Outcome:  "yes",
			},
			expected: security.BetInput{
				MarketID: "123",
				Amount:   100.0,
				Outcome:  "YES", // Should be normalized
			},
		},
		{
			name: "Market ID with whitespace",
			input: security.BetInput{
				MarketID: "  123  ",
				Amount:   100.0,
				Outcome:  "NO",
			},
			expected: security.BetInput{
				MarketID: "123", // Should be trimmed
				Amount:   100.0,
				Outcome:  "NO",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized, err := securityService.ValidateAndSanitizeBetInput(tt.input)

			if err != nil {
				t.Errorf("Sanitization failed: %v", err)
				return
			}

			if sanitized.MarketID != tt.expected.MarketID {
				t.Errorf("Expected MarketID %s, got %s", tt.expected.MarketID, sanitized.MarketID)
			}

			if sanitized.Amount != tt.expected.Amount {
				t.Errorf("Expected Amount %f, got %f", tt.expected.Amount, sanitized.Amount)
			}

			if sanitized.Outcome != tt.expected.Outcome {
				t.Errorf("Expected Outcome %s, got %s", tt.expected.Outcome, sanitized.Outcome)
			}
		})
	}
}

func TestBetInputSecurityFeatures(t *testing.T) {
	securityService := security.NewSecurityService()

	// Test various malicious inputs
	maliciousInputs := []struct {
		name         string
		input        security.BetInput
		shouldReject bool
	}{
		{
			name: "SQL injection in market ID",
			input: security.BetInput{
				MarketID: "1; DROP TABLE markets; --",
				Amount:   100.0,
				Outcome:  "YES",
			},
			shouldReject: true,
		},
		{
			name: "XSS in market ID",
			input: security.BetInput{
				MarketID: "<script>alert('xss')</script>",
				Amount:   100.0,
				Outcome:  "YES",
			},
			shouldReject: true,
		},
		{
			name: "XSS in outcome",
			input: security.BetInput{
				MarketID: "123",
				Amount:   100.0,
				Outcome:  "<script>alert('xss')</script>",
			},
			shouldReject: true,
		},
		{
			name: "NaN amount",
			input: security.BetInput{
				MarketID: "123",
				Amount:   -1.0, // Use negative as proxy for invalid number
				Outcome:  "YES",
			},
			shouldReject: true,
		},
		{
			name: "Extremely large amount",
			input: security.BetInput{
				MarketID: "123",
				Amount:   1e50, // Very large number
				Outcome:  "YES",
			},
			shouldReject: true,
		},
		{
			name: "Very long market ID",
			input: security.BetInput{
				MarketID: string(make([]byte, 1000)),
				Amount:   100.0,
				Outcome:  "YES",
			},
			shouldReject: true,
		},
	}

	for _, test := range maliciousInputs {
		t.Run(test.name, func(t *testing.T) {
			_, err := securityService.ValidateAndSanitizeBetInput(test.input)

			if test.shouldReject {
				if err == nil {
					t.Errorf("Expected malicious input to be rejected but validation passed")
				}
			} else {
				if err != nil {
					t.Errorf("Expected input to be accepted but got error: %v", err)
				}
			}
		})
	}
}

func TestBetAmountValidation(t *testing.T) {
	securityService := security.NewSecurityService()

	amountTests := []struct {
		name       string
		amount     float64
		shouldPass bool
	}{
		{"Positive small amount", 0.01, true},
		{"Positive normal amount", 100.0, true},
		{"Positive large amount", 10000.0, true},
		{"Zero amount", 0.0, false},
		{"Negative amount", -1.0, false},
		{"Very negative amount", -1000.0, false},
		{"Extremely large amount", 1e20, false}, // May be rejected as unrealistic
	}

	for _, test := range amountTests {
		t.Run(test.name, func(t *testing.T) {
			input := security.BetInput{
				MarketID: "123",
				Amount:   test.amount,
				Outcome:  "YES",
			}

			_, err := securityService.ValidateAndSanitizeBetInput(input)

			if test.shouldPass {
				if err != nil {
					t.Errorf("Expected amount %f to be valid but got error: %v", test.amount, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected amount %f to be invalid but validation passed", test.amount)
				}
			}
		})
	}
}

func TestBetOutcomeValidation(t *testing.T) {
	securityService := security.NewSecurityService()

	outcomeTests := []struct {
		name       string
		outcome    string
		shouldPass bool
		expected   string
	}{
		{"YES uppercase", "YES", true, "YES"},
		{"NO uppercase", "NO", true, "NO"},
		{"yes lowercase", "yes", true, "YES"},
		{"no lowercase", "no", true, "NO"},
		{"Yes mixed case", "Yes", true, "YES"},
		{"No mixed case", "No", true, "NO"},
		{"Invalid outcome", "MAYBE", false, ""},
		{"Empty outcome", "", false, ""},
		{"Numeric outcome", "1", false, ""},
		{"Random text", "random", false, ""},
		{"XSS attempt", "<script>alert('xss')</script>", false, ""},
	}

	for _, test := range outcomeTests {
		t.Run(test.name, func(t *testing.T) {
			input := security.BetInput{
				MarketID: "123",
				Amount:   100.0,
				Outcome:  test.outcome,
			}

			sanitized, err := securityService.ValidateAndSanitizeBetInput(input)

			if test.shouldPass {
				if err != nil {
					t.Errorf("Expected outcome '%s' to be valid but got error: %v", test.outcome, err)
				} else if sanitized.Outcome != test.expected {
					t.Errorf("Expected outcome to be normalized to '%s' but got '%s'", test.expected, sanitized.Outcome)
				}
			} else {
				if err == nil {
					t.Errorf("Expected outcome '%s' to be invalid but validation passed", test.outcome)
				}
			}
		})
	}
}
