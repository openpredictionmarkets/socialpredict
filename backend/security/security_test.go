package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSecurityService(t *testing.T) {
	service := NewSecurityService()

	if service.Sanitizer == nil {
		t.Error("Expected sanitizer to be initialized")
	}

	if service.Validator == nil {
		t.Error("Expected validator to be initialized")
	}

	if service.RateManager == nil {
		t.Error("Expected rate manager to be initialized")
	}

	if service.Headers.ContentTypeOptions == "" {
		t.Error("Expected security headers to be initialized")
	}
}

func TestNewCustomSecurityService(t *testing.T) {
	config := DefaultRateLimitConfig()
	service := NewCustomSecurityService(config)

	if service.Sanitizer == nil {
		t.Error("Expected sanitizer to be initialized")
	}

	if service.Validator == nil {
		t.Error("Expected validator to be initialized")
	}

	if service.RateManager == nil {
		t.Error("Expected rate manager to be initialized")
	}
}

func TestSecurityMiddleware(t *testing.T) {
	service := NewSecurityService()
	middleware := service.SecurityMiddleware()

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap handler with middleware
	wrappedHandler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	// Check that security headers are set
	if rr.Header().Get("X-Content-Type-Options") == "" {
		t.Error("Expected X-Content-Type-Options header to be set")
	}

	if rr.Header().Get("X-Frame-Options") == "" {
		t.Error("Expected X-Frame-Options header to be set")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", rr.Code)
	}
}

func TestLoginSecurityMiddleware(t *testing.T) {
	service := NewSecurityService()
	middleware := service.LoginSecurityMiddleware()

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("login success"))
	})

	// Wrap handler with middleware
	wrappedHandler := middleware(testHandler)

	req := httptest.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	// Check that security headers are set
	if rr.Header().Get("X-Content-Type-Options") == "" {
		t.Error("Expected X-Content-Type-Options header to be set")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", rr.Code)
	}
}

func TestValidateAndSanitizeUserInput(t *testing.T) {
	service := NewSecurityService()

	tests := []struct {
		name    string
		input   UserInput
		wantErr bool
	}{
		{
			name: "valid user input",
			input: UserInput{
				Username:      "testuser123",
				DisplayName:   "Test User",
				Description:   "A valid description",
				PersonalEmoji: "😀",
				PersonalLink1: "https://example.com",
				PersonalLink2: "https://example2.com",
				PersonalLink3: "https://example3.com",
				PersonalLink4: "https://example4.com",
				Password:      "StrongPass123",
			},
			wantErr: false,
		},
		{
			name: "invalid username",
			input: UserInput{
				Username:      "Test_User",
				DisplayName:   "Test User",
				Description:   "A valid description",
				PersonalEmoji: "😀",
				Password:      "StrongPass123",
			},
			wantErr: true,
		},
		{
			name: "dangerous display name",
			input: UserInput{
				Username:      "testuser123",
				DisplayName:   "Test <script>alert('xss')</script> User",
				Description:   "A valid description",
				PersonalEmoji: "😀",
				Password:      "StrongPass123",
			},
			wantErr: true,
		},
		{
			name: "weak password",
			input: UserInput{
				Username:      "testuser123",
				DisplayName:   "Test User",
				Description:   "A valid description",
				PersonalEmoji: "😀",
				Password:      "weak",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ValidateAndSanitizeUserInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndSanitizeUserInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("Expected result to be non-nil for valid input")
			}

			if !tt.wantErr && result != nil {
				if result.Username != "testuser123" {
					t.Errorf("Expected username to be sanitized correctly, got %s", result.Username)
				}
				if result.DisplayName == "" {
					t.Error("Expected display name to be sanitized and non-empty")
				}
			}
		})
	}
}

func TestValidateAndSanitizeMarketInput(t *testing.T) {
	service := NewSecurityService()

	// Use a future date for testing
	futureDate := "2026-12-31T23:59:59Z"

	tests := []struct {
		name    string
		input   MarketInput
		wantErr bool
	}{
		{
			name: "valid market input",
			input: MarketInput{
				Title:       "Will it rain tomorrow?",
				Description: "A market about weather prediction",
				EndTime:     futureDate,
			},
			wantErr: false,
		},
		{
			name: "dangerous title",
			input: MarketInput{
				Title:       "Will it rain <script>alert('xss')</script> tomorrow?",
				Description: "A market about weather prediction",
				EndTime:     futureDate,
			},
			wantErr: true,
		},
		{
			name: "empty title",
			input: MarketInput{
				Title:       "",
				Description: "A market about weather prediction",
				EndTime:     futureDate,
			},
			wantErr: true,
		},
		{
			name: "too long title",
			input: MarketInput{
				Title:       "This is a very long title that exceeds the maximum allowed length for market titles and should cause a validation error because it's way too long and contains more than 160 characters which is the limit",
				Description: "A market about weather prediction",
				EndTime:     futureDate,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ValidateAndSanitizeMarketInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndSanitizeMarketInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("Expected result to be non-nil for valid input")
			}

			if !tt.wantErr && result != nil {
				if result.Title == "" {
					t.Error("Expected title to be sanitized and non-empty")
				}
				if result.EndTime != tt.input.EndTime {
					t.Error("Expected end time to be preserved")
				}
			}
		})
	}
}

func TestValidateAndSanitizeBetInput(t *testing.T) {
	service := NewSecurityService()

	tests := []struct {
		name    string
		input   BetInput
		wantErr bool
	}{
		{
			name: "valid bet input",
			input: BetInput{
				MarketID: "12345",
				Amount:   100.50,
				Outcome:  "YES",
			},
			wantErr: false,
		},
		{
			name: "invalid outcome",
			input: BetInput{
				MarketID: "12345",
				Amount:   100.50,
				Outcome:  "MAYBE",
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			input: BetInput{
				MarketID: "12345",
				Amount:   -100.50,
				Outcome:  "YES",
			},
			wantErr: true,
		},
		{
			name: "empty market ID",
			input: BetInput{
				MarketID: "",
				Amount:   100.50,
				Outcome:  "YES",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ValidateAndSanitizeBetInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndSanitizeBetInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("Expected result to be non-nil for valid input")
			}

			if !tt.wantErr && result != nil {
				if result.MarketID != tt.input.MarketID {
					t.Error("Expected market ID to be preserved")
				}
				if result.Amount != tt.input.Amount {
					t.Error("Expected amount to be preserved")
				}
				if result.Outcome != tt.input.Outcome {
					t.Error("Expected outcome to be preserved")
				}
			}
		})
	}
}

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	if config.RateLimit.LoginRate <= 0 {
		t.Error("Expected positive login rate in config")
	}

	if config.RateLimit.GeneralRate <= 0 {
		t.Error("Expected positive general rate in config")
	}

	if config.Headers.ContentTypeOptions == "" {
		t.Error("Expected security headers to be configured")
	}
}
