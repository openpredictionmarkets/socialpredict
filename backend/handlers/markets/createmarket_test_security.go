package marketshandlers

import (
	"socialpredict/security"
	"strings"
	"testing"
)

func TestMarketTitleValidation(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		expectedValid bool
		expectedError string
	}{
		{
			name:          "Valid market title",
			title:         "Will Bitcoin reach $100,000 by 2024?",
			expectedValid: true,
		},
		{
			name:          "Empty title",
			title:         "",
			expectedValid: false,
			expectedError: "exceeds 160 characters or is blank",
		},
		{
			name:          "Title at max length",
			title:         strings.Repeat("a", maxQuestionTitleLength),
			expectedValid: true,
		},
		{
			name:          "Title too long",
			title:         strings.Repeat("a", maxQuestionTitleLength+1),
			expectedValid: false,
			expectedError: "exceeds 160 characters or is blank",
		},
		{
			name:          "Title with XSS",
			title:         "Will stocks rise?<script>alert('xss')</script>",
			expectedValid: true, // Should be sanitized
		},
		{
			name:          "Title with HTML",
			title:         "Will <b>Tesla</b> stock rise?",
			expectedValid: true, // Should be sanitized
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the existing length validation function
			err := checkQuestionTitleLength(tt.title)

			if tt.expectedValid && len(tt.title) > 0 && len(tt.title) <= maxQuestionTitleLength {
				if err != nil {
					t.Errorf("Expected valid title but got error: %v", err)
				}

				// Test sanitization for valid titles
				securityService := security.NewSecurityService()
				marketInput := security.MarketInput{
					Title:       tt.title,
					Description: "Test description",
					EndTime:     "2024-12-31T23:59:59Z",
				}

				sanitized, sanitizeErr := securityService.ValidateAndSanitizeMarketInput(marketInput)
				if sanitizeErr != nil {
					t.Errorf("Sanitization failed for valid title: %v", sanitizeErr)
				}

				// Check XSS was sanitized
				if strings.Contains(tt.title, "<script>") && strings.Contains(sanitized.Title, "<script>") {
					t.Error("XSS script tag was not sanitized")
				}
			} else {
				if err == nil {
					t.Errorf("Expected error but title validation passed")
				}
			}
		})
	}
}

func TestMarketDescriptionValidation(t *testing.T) {
	tests := []struct {
		name          string
		description   string
		expectedValid bool
		expectedError string
	}{
		{
			name:          "Valid description",
			description:   "This market will resolve based on official price data from CoinMarketCap.",
			expectedValid: true,
		},
		{
			name:          "Empty description",
			description:   "",
			expectedValid: true, // Empty descriptions are allowed
		},
		{
			name:          "Long valid description",
			description:   strings.Repeat("Valid description content. ", 50),
			expectedValid: true,
		},
		{
			name:          "Description at max length",
			description:   strings.Repeat("a", 2000),
			expectedValid: true,
		},
		{
			name:          "Description too long",
			description:   strings.Repeat("a", 2001),
			expectedValid: false,
			expectedError: "exceeds 2000 characters",
		},
		{
			name:          "Description with XSS",
			description:   "Market rules:<script>alert('xss')</script>",
			expectedValid: true, // Should be sanitized
		},
		{
			name:          "Description with HTML",
			description:   "Rules: <b>Must be official data</b>",
			expectedValid: true, // Should be sanitized
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the existing length validation function
			err := checkQuestionDescriptionLength(tt.description)

			if tt.expectedValid && len(tt.description) <= 2000 {
				if err != nil {
					t.Errorf("Expected valid description but got error: %v", err)
				}

				// Test sanitization for valid descriptions
				securityService := security.NewSecurityService()
				marketInput := security.MarketInput{
					Title:       "Test title",
					Description: tt.description,
					EndTime:     "2024-12-31T23:59:59Z",
				}

				sanitized, sanitizeErr := securityService.ValidateAndSanitizeMarketInput(marketInput)
				if sanitizeErr != nil {
					t.Errorf("Sanitization failed for valid description: %v", sanitizeErr)
				}

				// Check XSS was sanitized
				if strings.Contains(tt.description, "<script>") && strings.Contains(sanitized.Description, "<script>") {
					t.Error("XSS script tag was not sanitized")
				}
			} else {
				if err == nil {
					t.Errorf("Expected error but description validation passed")
				}
			}
		})
	}
}

func TestMarketInputXSSPrevention(t *testing.T) {
	securityService := security.NewSecurityService()

	xssPayloads := []struct {
		field   string
		payload string
	}{
		{"title", "Will Bitcoin rise?<script>alert('xss')</script>"},
		{"title", "Market<img src=x onerror=alert('xss')>question"},
		{"title", "Question<svg onload=alert('xss')>?"},
		{"description", "Rules:<script>alert('xss')</script>"},
		{"description", "Description<iframe src=javascript:alert('xss')></iframe>"},
		{"description", "Text<object data=javascript:alert('xss')>"},
		{"description", "Content<embed src=javascript:alert('xss')>"},
		{"description", "Rules<link rel=stylesheet href=javascript:alert('xss')>"},
	}

	for _, test := range xssPayloads {
		t.Run("XSS_"+test.field+"_"+test.payload[:20], func(t *testing.T) {
			var marketInput security.MarketInput

			if test.field == "title" {
				marketInput = security.MarketInput{
					Title:       test.payload,
					Description: "Safe description",
					EndTime:     "2024-12-31T23:59:59Z",
				}
			} else {
				marketInput = security.MarketInput{
					Title:       "Safe title",
					Description: test.payload,
					EndTime:     "2024-12-31T23:59:59Z",
				}
			}

			sanitized, err := securityService.ValidateAndSanitizeMarketInput(marketInput)

			if err != nil {
				// Some payloads might be rejected completely, which is fine
				return
			}

			// Verify dangerous content was sanitized
			var sanitizedField string
			if test.field == "title" {
				sanitizedField = sanitized.Title
			} else {
				sanitizedField = sanitized.Description
			}

			if sanitizedField == test.payload {
				t.Errorf("Dangerous payload was not sanitized in %s: %s", test.field, test.payload)
			}

			// Verify no script tags remain
			if strings.Contains(sanitizedField, "<script>") {
				t.Errorf("Script tag remained after sanitization in %s: %s -> %s", test.field, test.payload, sanitizedField)
			}

			// Verify no javascript: protocols remain
			if strings.Contains(sanitizedField, "javascript:") {
				t.Errorf("JavaScript protocol remained after sanitization in %s: %s -> %s", test.field, test.payload, sanitizedField)
			}
		})
	}
}

func TestMarketInputValidation(t *testing.T) {
	securityService := security.NewSecurityService()

	tests := []struct {
		name          string
		input         security.MarketInput
		expectedValid bool
		expectedError string
	}{
		{
			name: "Valid market input",
			input: security.MarketInput{
				Title:       "Will Bitcoin reach $100k?",
				Description: "Market resolves based on CoinMarketCap data",
				EndTime:     "2024-12-31T23:59:59Z",
			},
			expectedValid: true,
		},
		{
			name: "Empty title",
			input: security.MarketInput{
				Title:       "",
				Description: "Valid description",
				EndTime:     "2024-12-31T23:59:59Z",
			},
			expectedValid: false,
		},
		{
			name: "Title too long",
			input: security.MarketInput{
				Title:       strings.Repeat("a", 200),
				Description: "Valid description",
				EndTime:     "2024-12-31T23:59:59Z",
			},
			expectedValid: false,
		},
		{
			name: "Description too long",
			input: security.MarketInput{
				Title:       "Valid title",
				Description: strings.Repeat("a", 3000),
				EndTime:     "2024-12-31T23:59:59Z",
			},
			expectedValid: false,
		},
		{
			name: "Valid minimal input",
			input: security.MarketInput{
				Title:       "Short?",
				Description: "",
				EndTime:     "2024-12-31T23:59:59Z",
			},
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := securityService.ValidateAndSanitizeMarketInput(tt.input)

			if tt.expectedValid {
				if err != nil {
					t.Errorf("Expected valid input but got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error but validation passed")
				}
			}
		})
	}
}

func TestMarketInputSanitization(t *testing.T) {
	securityService := security.NewSecurityService()

	input := security.MarketInput{
		Title:       "Will <b>Tesla</b> stock <script>alert('xss')</script> rise?",
		Description: "Market rules: <i>Official data only</i><script>alert('desc')</script>",
		EndTime:     "2024-12-31T23:59:59Z",
	}

	sanitized, err := securityService.ValidateAndSanitizeMarketInput(input)
	if err != nil {
		t.Fatalf("Sanitization failed: %v", err)
	}

	// Verify HTML tags are handled appropriately
	if strings.Contains(sanitized.Title, "<script>") {
		t.Error("Script tag not removed from title")
	}

	if strings.Contains(sanitized.Description, "<script>") {
		t.Error("Script tag not removed from description")
	}

	// Verify the content is still readable (important text preserved)
	if !strings.Contains(sanitized.Title, "Tesla") {
		t.Error("Important content was removed from title")
	}

	if !strings.Contains(sanitized.Description, "Official data") {
		t.Error("Important content was removed from description")
	}
}
