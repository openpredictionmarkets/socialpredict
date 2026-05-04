package marketshandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"socialpredict/handlers"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/security"
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
			err := ValidateMarketResolutionTime(tt.resolutionTime, config)

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
			err := ValidateMarketResolutionTime(resolutionTime, config)

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

func TestCreateMarketHandlerWithService_InvalidInputUsesFailureEnvelope(t *testing.T) {
	auth := &contractAuthMock{user: &dusers.User{Username: "alice"}}

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantReason handlers.FailureReason
	}{
		{
			name:       "invalid JSON uses invalid request",
			body:       `{"questionTitle":`,
			wantStatus: http.StatusBadRequest,
			wantReason: handlers.ReasonInvalidRequest,
		},
		{
			name:       "sanitizer rejection uses invalid request without parser string",
			body:       `{"questionTitle":"Will BTC rise?<script>alert(1)</script>","description":"Market","outcomeType":"BINARY","resolutionDateTime":"2030-01-01T00:00:00Z"}`,
			wantStatus: http.StatusBadRequest,
			wantReason: handlers.ReasonInvalidRequest,
		},
		{
			name:       "missing security service uses internal error",
			body:       `{"questionTitle":"Will BTC rise?","description":"Market","outcomeType":"BINARY","resolutionDateTime":"2030-01-01T00:00:00Z"}`,
			wantStatus: http.StatusInternalServerError,
			wantReason: handlers.ReasonInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &searchServiceMock{
				createFn: func(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error) {
					t.Fatalf("service should not be called for invalid boundary input")
					return nil, nil
				},
			}
			securityService := security.NewSecurityService()
			if tt.wantReason == handlers.ReasonInternalError {
				securityService = nil
			}

			handler := CreateMarketHandlerWithService(service, auth, nil, securityService)
			req := httptest.NewRequest(http.MethodPost, "/v0/markets", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			handler(rr, req)

			assertFailureEnvelope(t, rr, tt.wantStatus, tt.wantReason)
			assertNoLegacyErrorText(t, rr)
		})
	}
}

func TestCreateMarketHandlerWithService_DomainValidationUsesStableReason(t *testing.T) {
	auth := &contractAuthMock{user: &dusers.User{Username: "alice"}}
	service := &searchServiceMock{
		createFn: func(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error) {
			return nil, dmarkets.ErrInvalidResolutionTime
		},
	}

	handler := CreateMarketHandlerWithService(service, auth, nil, security.NewSecurityService())
	req := httptest.NewRequest(http.MethodPost, "/v0/markets", bytes.NewBufferString(
		`{"questionTitle":"Will BTC rise?","description":"Market","outcomeType":"BINARY","resolutionDateTime":"2030-01-01T00:00:00Z"}`,
	))
	rr := httptest.NewRecorder()

	handler(rr, req)

	assertFailureEnvelope(t, rr, http.StatusBadRequest, handlers.ReasonValidationFailed)
	assertNoLegacyErrorText(t, rr)
}

func assertNoLegacyErrorText(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	var response handlers.FailureEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if strings.Contains(rr.Body.String(), "Invalid market data") ||
		strings.Contains(rr.Body.String(), "<script>") ||
		strings.Contains(rr.Body.String(), "market resolution time") {
		t.Fatalf("failure body leaked legacy or parser text: %s", rr.Body.String())
	}
}
