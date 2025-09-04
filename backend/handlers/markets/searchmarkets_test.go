package marketshandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSearchMarketsHandler(t *testing.T) {
	// Setup test database
	db := modelstesting.NewFakeDB(t)
	util.DB = db // Set global DB for util.GetDB()

	// Create test markets
	testMarkets := []models.Market{
		{
			ID:                 1,
			QuestionTitle:      "Will Bitcoin reach $100k?",
			Description:        "Test market about Bitcoin",
			OutcomeType:        "BINARY",
			ResolutionDateTime: time.Now().Add(24 * time.Hour),
			IsResolved:         false,
			InitialProbability: 0.5,
			CreatorUsername:    "testuser",
		},
		{
			ID:                      2,
			QuestionTitle:           "Will Ethereum overtake Bitcoin?",
			Description:             "Another crypto market",
			OutcomeType:             "BINARY",
			ResolutionDateTime:      time.Now().Add(-1 * time.Hour), // Closed
			FinalResolutionDateTime: time.Now(),
			IsResolved:              true,
			ResolutionResult:        "YES",
			InitialProbability:      0.5,
			CreatorUsername:         "testuser",
		},
		{
			ID:                 3,
			QuestionTitle:      "Will the stock market crash?",
			Description:        "Market about stocks",
			OutcomeType:        "BINARY",
			ResolutionDateTime: time.Now().Add(48 * time.Hour),
			IsResolved:         false,
			InitialProbability: 0.5,
			CreatorUsername:    "testuser",
		},
	}

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 0)
	db.Create(&testUser)

	// Insert test markets
	for _, market := range testMarkets {
		db.Create(&market)
	}

	tests := []struct {
		name           string
		query          string
		status         string
		limit          string
		expectedStatus int
		expectedCount  int
		searchTerm     string
	}{
		{
			name:           "Search Bitcoin in all markets",
			query:          "Bitcoin",
			status:         "all",
			limit:          "10",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // Should find both Bitcoin markets
			searchTerm:     "Bitcoin",
		},
		{
			name:           "Search active markets only",
			query:          "Bitcoin",
			status:         "active",
			limit:          "10",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // Active + fallback resolved Bitcoin market
			searchTerm:     "Bitcoin",
		},
		{
			name:           "Search resolved markets only",
			query:          "Ethereum",
			status:         "resolved",
			limit:          "10",
			expectedStatus: http.StatusOK,
			expectedCount:  1, // Only resolved Ethereum market
			searchTerm:     "Ethereum",
		},
		{
			name:           "Search with no results",
			query:          "NonexistentTerm",
			status:         "all",
			limit:          "10",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			searchTerm:     "NonexistentTerm",
		},
		{
			name:           "Empty query should fail",
			query:          "",
			status:         "all",
			limit:          "10",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			searchTerm:     "",
		},
		{
			name:           "Case insensitive search",
			query:          "bitcoin",
			status:         "all",
			limit:          "10",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // Should find Bitcoin markets regardless of case
			searchTerm:     "bitcoin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/v0/markets/search", nil)
			q := req.URL.Query()
			if tt.query != "" {
				q.Add("query", tt.query)
			}
			if tt.status != "" {
				q.Add("status", tt.status)
			}
			if tt.limit != "" {
				q.Add("limit", tt.limit)
			}
			req.URL.RawQuery = q.Encode()

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			SearchMarketsHandler(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Parse response
				var response SearchMarketsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Check query matches
				if tt.query != "" {
					assert.Equal(t, tt.searchTerm, response.Query)
				}

				// Check total count
				assert.Equal(t, tt.expectedCount, response.TotalCount)

				// If we expect results, verify structure
				if tt.expectedCount > 0 {
					assert.LessOrEqual(t, response.PrimaryCount, tt.expectedCount)
					assert.GreaterOrEqual(t, len(response.PrimaryResults), 0)
				}
			}
		})
	}
}

func TestSearchMarketsFunction(t *testing.T) {
	// Setup test database
	db := modelstesting.NewFakeDB(t)
	util.DB = db // Set global DB for util.GetDB()

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 0)
	db.Create(&testUser)

	// Create test markets with different statuses
	activeMarket := models.Market{
		ID:                 1,
		QuestionTitle:      "Active Bitcoin Market",
		Description:        "Test active market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: time.Now().Add(24 * time.Hour),
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}

	resolvedMarket := models.Market{
		ID:                      2,
		QuestionTitle:           "Resolved Bitcoin Market",
		Description:             "Test resolved market",
		OutcomeType:             "BINARY",
		ResolutionDateTime:      time.Now().Add(-1 * time.Hour),
		FinalResolutionDateTime: time.Now(),
		IsResolved:              true,
		ResolutionResult:        "YES",
		InitialProbability:      0.5,
		CreatorUsername:         "testuser",
	}

	db.Create(&activeMarket)
	db.Create(&resolvedMarket)

	tests := []struct {
		name                 string
		query                string
		status               string
		limit                int
		expectedPrimary      int
		expectedFallback     int
		expectedFallbackUsed bool
	}{
		{
			name:                 "Search all Bitcoin markets",
			query:                "Bitcoin",
			status:               "all",
			limit:                10,
			expectedPrimary:      2,
			expectedFallback:     0,
			expectedFallbackUsed: false,
		},
		{
			name:                 "Search active Bitcoin markets with fallback",
			query:                "Bitcoin",
			status:               "active",
			limit:                10,
			expectedPrimary:      1,
			expectedFallback:     1, // Should get the resolved one as fallback
			expectedFallbackUsed: true,
		},
		{
			name:                 "Search resolved Bitcoin markets with fallback",
			query:                "Bitcoin",
			status:               "resolved",
			limit:                10,
			expectedPrimary:      1,
			expectedFallback:     1, // Should get the active one as fallback
			expectedFallbackUsed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SearchMarkets(db, tt.query, tt.status, tt.limit)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			assert.Equal(t, tt.query, result.Query)
			assert.Equal(t, tt.expectedPrimary, result.PrimaryCount)
			assert.Equal(t, tt.expectedFallback, result.FallbackCount)
			assert.Equal(t, tt.expectedFallbackUsed, result.FallbackUsed)
			assert.Equal(t, tt.expectedPrimary+tt.expectedFallback, result.TotalCount)
		})
	}
}

func TestSearchMarketsWithInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		method string
		query  string
		status int
	}{
		{
			name:   "Invalid HTTP method",
			method: http.MethodPost,
			query:  "test",
			status: http.StatusMethodNotAllowed,
		},
		{
			name:   "Missing query parameter",
			method: http.MethodGet,
			query:  "",
			status: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/v0/markets/search", nil)
			if tt.query != "" {
				q := req.URL.Query()
				q.Add("query", tt.query)
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()
			SearchMarketsHandler(w, req)
			assert.Equal(t, tt.status, w.Code)
		})
	}
}

func TestSearchMarketsLimitParameter(t *testing.T) {
	// Setup test database
	db := modelstesting.NewFakeDB(t)
	util.DB = db // Set global DB for util.GetDB()

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 0)
	db.Create(&testUser)

	// Create multiple test markets
	for i := 1; i <= 10; i++ {
		market := models.Market{
			ID:                 int64(i),
			QuestionTitle:      "Test Market " + strconv.Itoa(i),
			Description:        "Test description",
			OutcomeType:        "BINARY",
			ResolutionDateTime: time.Now().Add(24 * time.Hour),
			IsResolved:         false,
			InitialProbability: 0.5,
			CreatorUsername:    "testuser",
		}
		db.Create(&market)
	}

	tests := []struct {
		name          string
		limit         string
		expectedCount int
	}{
		{
			name:          "Default limit",
			limit:         "",
			expectedCount: 10, // Should get all 10 markets
		},
		{
			name:          "Limit to 5",
			limit:         "5",
			expectedCount: 5,
		},
		{
			name:          "Limit to 15 (more than available)",
			limit:         "15",
			expectedCount: 10, // Should get all 10 available
		},
		{
			name:          "Invalid limit (too high)",
			limit:         "100",
			expectedCount: 10, // Should default to reasonable limit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v0/markets/search", nil)
			q := req.URL.Query()
			q.Add("query", "Test")
			if tt.limit != "" {
				q.Add("limit", tt.limit)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			SearchMarketsHandler(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response SearchMarketsResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.LessOrEqual(t, response.TotalCount, tt.expectedCount)
		})
	}
}

func TestSearchMarketsCaseInsensitiveComprehensive(t *testing.T) {
	// Setup test database
	db := modelstesting.NewFakeDB(t)
	util.DB = db // Set global DB for util.GetDB()

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 0)
	db.Create(&testUser)

	// Create test markets with various case patterns
	testMarkets := []models.Market{
		{
			ID:                 1,
			QuestionTitle:      "BITCOIN Price Prediction",
			Description:        "Market about bitcoin prices",
			OutcomeType:        "BINARY",
			ResolutionDateTime: time.Now().Add(24 * time.Hour),
			IsResolved:         false,
			InitialProbability: 0.5,
			CreatorUsername:    "testuser",
		},
		{
			ID:                 2,
			QuestionTitle:      "Will Bitcoin reach new highs?",
			Description:        "Another BTC market",
			OutcomeType:        "BINARY",
			ResolutionDateTime: time.Now().Add(24 * time.Hour),
			IsResolved:         false,
			InitialProbability: 0.5,
			CreatorUsername:    "testuser",
		},
		{
			ID:                 3,
			QuestionTitle:      "ethereum vs bitcoin",
			Description:        "Comparing ETH and BTC",
			OutcomeType:        "BINARY",
			ResolutionDateTime: time.Now().Add(24 * time.Hour),
			IsResolved:         false,
			InitialProbability: 0.5,
			CreatorUsername:    "testuser",
		},
	}

	for _, market := range testMarkets {
		db.Create(&market)
	}

	tests := []struct {
		name          string
		searchQuery   string
		expectedCount int
		description   string
	}{
		{
			name:          "Lowercase search",
			searchQuery:   "bitcoin",
			expectedCount: 3, // Should find all 3 markets
			description:   "Should find BITCOIN, Bitcoin, and bitcoin variations",
		},
		{
			name:          "Uppercase search",
			searchQuery:   "BITCOIN",
			expectedCount: 3, // Should find all 3 markets
			description:   "Should find all bitcoin variations regardless of case",
		},
		{
			name:          "Mixed case search",
			searchQuery:   "BitCoin",
			expectedCount: 3, // Should find all 3 markets
			description:   "Should find all bitcoin variations with mixed case",
		},
		{
			name:          "Partial match lowercase",
			searchQuery:   "btc",
			expectedCount: 2, // Should find markets with BTC
			description:   "Should find BTC matches case-insensitively",
		},
		{
			name:          "Description search",
			searchQuery:   "eth",
			expectedCount: 1, // Should find the ethereum market
			description:   "Should search in descriptions case-insensitively",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SearchMarkets(db, tt.searchQuery, "all", 10)
			assert.NoError(t, err, tt.description)
			assert.Equal(t, tt.expectedCount, result.TotalCount,
				"For query '%s': %s. Expected %d, got %d",
				tt.searchQuery, tt.description, tt.expectedCount, result.TotalCount)
		})
	}
}

func TestSearchMarketsFallbackThreshold(t *testing.T) {
	// Setup test database
	db := modelstesting.NewFakeDB(t)
	util.DB = db // Set global DB for util.GetDB()

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 0)
	db.Create(&testUser)

	// Create markets with specific statuses for testing fallback
	markets := []struct {
		id          int64
		title       string
		isActive    bool
		isResolved  bool
		description string
	}{
		// Active markets with "crypto"
		{1, "Crypto Market 1", true, false, "Active crypto market"},
		{2, "Crypto Market 2", true, false, "Another active crypto market"},
		{3, "Crypto Market 3", true, false, "Third active crypto market"},

		// Resolved markets with "crypto"
		{4, "Crypto Market 4", false, true, "Resolved crypto market"},
		{5, "Crypto Market 5", false, true, "Another resolved crypto market"},
		{6, "Crypto Market 6", false, true, "Third resolved crypto market"},
		{7, "Crypto Market 7", false, true, "Fourth resolved crypto market"},

		// Non-crypto markets
		{8, "Stock Market", true, false, "Regular stock market"},
		{9, "Weather Prediction", false, true, "Weather market"},
	}

	for _, m := range markets {
		market := models.Market{
			ID:                 m.id,
			QuestionTitle:      m.title,
			Description:        m.description,
			OutcomeType:        "BINARY",
			InitialProbability: 0.5,
			CreatorUsername:    "testuser",
		}

		if m.isActive {
			market.ResolutionDateTime = time.Now().Add(24 * time.Hour)
			market.IsResolved = false
		} else if m.isResolved {
			market.ResolutionDateTime = time.Now().Add(-1 * time.Hour)
			market.FinalResolutionDateTime = time.Now()
			market.IsResolved = true
			market.ResolutionResult = "YES"
		} else {
			// Closed but not resolved
			market.ResolutionDateTime = time.Now().Add(-1 * time.Hour)
			market.IsResolved = false
		}

		db.Create(&market)
	}

	tests := []struct {
		name                 string
		query                string
		status               string
		expectedPrimary      int
		expectedFallback     int
		expectedFallbackUsed bool
		description          string
	}{
		{
			name:                 "Active crypto search - fallback triggered",
			query:                "crypto",
			status:               "active",
			expectedPrimary:      3, // 3 active crypto markets
			expectedFallback:     4, // 4 resolved crypto markets
			expectedFallbackUsed: true,
			description:          "Should find 3 active + 4 resolved crypto markets as fallback",
		},
		{
			name:                 "Resolved crypto search - no fallback needed",
			query:                "crypto",
			status:               "resolved",
			expectedPrimary:      4, // 4 resolved crypto markets
			expectedFallback:     3, // 3 active crypto markets as fallback
			expectedFallbackUsed: true,
			description:          "Should find 4 resolved + 3 active crypto markets as fallback",
		},
		{
			name:                 "All crypto search - no fallback",
			query:                "crypto",
			status:               "all",
			expectedPrimary:      7, // All 7 crypto markets
			expectedFallback:     0,
			expectedFallbackUsed: false,
			description:          "Should find all 7 crypto markets with no fallback needed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SearchMarkets(db, tt.query, tt.status, 10)
			assert.NoError(t, err, tt.description)
			assert.Equal(t, tt.expectedPrimary, result.PrimaryCount,
				"Primary count mismatch for %s", tt.description)
			assert.Equal(t, tt.expectedFallback, result.FallbackCount,
				"Fallback count mismatch for %s", tt.description)
			assert.Equal(t, tt.expectedFallbackUsed, result.FallbackUsed,
				"Fallback used mismatch for %s", tt.description)
			assert.Equal(t, tt.expectedPrimary+tt.expectedFallback, result.TotalCount,
				"Total count mismatch for %s", tt.description)
		})
	}
}

func TestSearchMarketsEdgeCases(t *testing.T) {
	// Setup test database
	db := modelstesting.NewFakeDB(t)
	util.DB = db // Set global DB for util.GetDB()

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 0)
	db.Create(&testUser)

	// Create a market with special characters
	market := models.Market{
		ID:                 1,
		QuestionTitle:      "Market with @#$%^&*() special chars",
		Description:        "Description with números 123 and símbolos!",
		OutcomeType:        "BINARY",
		ResolutionDateTime: time.Now().Add(24 * time.Hour),
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}
	db.Create(&market)

	tests := []struct {
		name        string
		query       string
		expectCount int
		description string
	}{
		{
			name:        "Special characters search",
			query:       "@#$",
			expectCount: 1,
			description: "Should find market with special characters",
		},
		{
			name:        "Numbers search",
			query:       "123",
			expectCount: 1,
			description: "Should find market with numbers",
		},
		{
			name:        "Single character search",
			query:       "M",
			expectCount: 1,
			description: "Should find market with single character match",
		},
		{
			name:        "Empty-like query",
			query:       "   ",
			expectCount: 0,
			description: "Whitespace-only query should return no results",
		},
		{
			name:        "Very long query",
			query:       "this is a very long search query that probably won't match anything but should not break the system or cause any errors",
			expectCount: 0,
			description: "Very long query should be handled gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SearchMarkets(db, tt.query, "all", 10)
			if tt.query == "   " {
				// Whitespace query should return an error or empty results
				if err == nil {
					assert.Equal(t, 0, result.TotalCount, tt.description)
				}
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.expectCount, result.TotalCount, tt.description)
			}
		})
	}
}
