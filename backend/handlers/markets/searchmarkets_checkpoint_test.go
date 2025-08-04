package marketshandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSearchMarketsCheckpointRequirements tests the exact scenarios described in CHECKPOINT20250803-03.md
func TestSearchMarketsCheckpointRequirements(t *testing.T) {
	// Setup test database
	db := modelstesting.NewFakeDB(t)
	util.DB = db // Set global DB for util.GetDB()

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 0)
	db.Create(&testUser)

	// Create test markets as described in checkpoint
	now := time.Now()

	// Bitcoin markets with different statuses
	bitcoinActiveMarket := models.Market{
		ID:                 1,
		QuestionTitle:      "Will Bitcoin reach $100k by end of year?",
		Description:        "A market about bitcoin price predictions",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(30 * 24 * time.Hour), // Active (future)
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}

	bitcoinClosedMarket := models.Market{
		ID:                 2,
		QuestionTitle:      "Bitcoin market prediction",
		Description:        "Another bitcoin market that is now closed",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(-1 * time.Hour), // Closed (past, not resolved)
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}

	bitcoinResolvedMarket := models.Market{
		ID:                      3,
		QuestionTitle:           "Will Bitcoin overtake gold market cap?",
		Description:             "Historical bitcoin vs gold market",
		OutcomeType:             "BINARY",
		ResolutionDateTime:      now.Add(-2 * time.Hour),
		FinalResolutionDateTime: now.Add(-1 * time.Hour),
		IsResolved:              true,
		ResolutionResult:        "YES",
		InitialProbability:      0.5,
		CreatorUsername:         "testuser",
	}

	// Non-bitcoin market for control
	stockMarket := models.Market{
		ID:                 4,
		QuestionTitle:      "Stock market crash prediction",
		Description:        "Will stocks crash this year?",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(15 * 24 * time.Hour), // Active
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}

	// Insert test markets
	db.Create(&bitcoinActiveMarket)
	db.Create(&bitcoinClosedMarket)
	db.Create(&bitcoinResolvedMarket)
	db.Create(&stockMarket)

	// Test cases exactly as described in checkpoint
	testCases := []struct {
		name               string
		url                string
		expectedStatusCode int
		expectedMinResults int
		expectedMaxResults int
		description        string
	}{
		{
			name:               "Test Case 1 - Keyword Only",
			url:                "/v0/markets/search?query=bitcoin",
			expectedStatusCode: http.StatusOK,
			expectedMinResults: 3, // Should find all 3 bitcoin markets
			expectedMaxResults: 3,
			description:        "Should return all markets with 'bitcoin' in title or description, regardless of status",
		},
		{
			name:               "Test Case 2 - Keyword and Active Status",
			url:                "/v0/markets/search?query=bitcoin&status=active",
			expectedStatusCode: http.StatusOK,
			expectedMinResults: 1, // At least the active bitcoin market
			expectedMaxResults: 3, // Could include fallback results
			description:        "Should return markets with 'bitcoin' that are active (isResolved=false, ResolutionDateTime > now)",
		},
		{
			name:               "Test Case 3 - Keyword and Closed Status",
			url:                "/v0/markets/search?query=bitcoin&status=closed",
			expectedStatusCode: http.StatusOK,
			expectedMinResults: 1, // At least the closed bitcoin market
			expectedMaxResults: 3, // Could include fallback results
			description:        "Should return markets with 'bitcoin' that are closed (isResolved=false, ResolutionDateTime <= now)",
		},
		{
			name:               "Test Case 4 - Keyword and Resolved Status",
			url:                "/v0/markets/search?query=bitcoin&status=resolved",
			expectedStatusCode: http.StatusOK,
			expectedMinResults: 1, // At least the resolved bitcoin market
			expectedMaxResults: 3, // Could include fallback results
			description:        "Should return markets with 'bitcoin' that are resolved (isResolved=true)",
		},
		{
			name:               "Test Case 5 - All Status",
			url:                "/v0/markets/search?query=bitcoin&status=all",
			expectedStatusCode: http.StatusOK,
			expectedMinResults: 3, // Should find all 3 bitcoin markets
			expectedMaxResults: 3,
			description:        "Should behave identically to Test Case 1, returning all matching markets",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()

			// Call the handler
			SearchMarketsHandler(w, req)

			// Verify status code
			assert.Equal(t, tc.expectedStatusCode, w.Code,
				"Status code mismatch for %s: %s", tc.name, tc.description)

			if tc.expectedStatusCode == http.StatusOK {
				// Parse response
				var response SearchMarketsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Failed to parse response for %s", tc.name)

				// Verify result count is within expected range
				totalResults := response.TotalCount
				assert.GreaterOrEqual(t, totalResults, tc.expectedMinResults,
					"Too few results for %s: %s. Expected at least %d, got %d",
					tc.name, tc.description, tc.expectedMinResults, totalResults)

				assert.LessOrEqual(t, totalResults, tc.expectedMaxResults,
					"Too many results for %s: %s. Expected at most %d, got %d",
					tc.name, tc.description, tc.expectedMaxResults, totalResults)

				// Verify query matches
				assert.Equal(t, "bitcoin", response.Query,
					"Query mismatch for %s", tc.name)

				// Log detailed results for inspection
				t.Logf("%s: Found %d total results (%d primary + %d fallback)",
					tc.name, response.TotalCount, response.PrimaryCount, response.FallbackCount)

				if response.FallbackUsed {
					t.Logf("  Fallback was used for %s", tc.name)
				}
			}
		})
	}
}

// TestSearchMarketsStatusFiltering specifically tests the database filtering logic
func TestSearchMarketsStatusFiltering(t *testing.T) {
	// Setup test database
	db := modelstesting.NewFakeDB(t)
	util.DB = db

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 0)
	db.Create(&testUser)

	now := time.Now()

	// Create markets with precise timing for testing
	activeMarket := models.Market{
		ID:                 1,
		QuestionTitle:      "Active Test Market",
		Description:        "Test market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(1 * time.Hour), // 1 hour in future = active
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}

	closedMarket := models.Market{
		ID:                 2,
		QuestionTitle:      "Closed Test Market",
		Description:        "Test market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(-1 * time.Hour), // 1 hour in past = closed
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}

	resolvedMarket := models.Market{
		ID:                      3,
		QuestionTitle:           "Resolved Test Market",
		Description:             "Test market",
		OutcomeType:             "BINARY",
		ResolutionDateTime:      now.Add(-2 * time.Hour),
		FinalResolutionDateTime: now.Add(-1 * time.Hour),
		IsResolved:              true,
		ResolutionResult:        "YES",
		InitialProbability:      0.5,
		CreatorUsername:         "testuser",
	}

	db.Create(&activeMarket)
	db.Create(&closedMarket)
	db.Create(&resolvedMarket)

	tests := []struct {
		name        string
		status      string
		expectedIDs []int64
		description string
	}{
		{
			name:        "Active filter",
			status:      "active",
			expectedIDs: []int64{1}, // Only active market
			description: "Should return only markets where isResolved=false AND resolutionDateTime > now",
		},
		{
			name:        "Closed filter",
			status:      "closed",
			expectedIDs: []int64{2}, // Only closed market
			description: "Should return only markets where isResolved=false AND resolutionDateTime <= now",
		},
		{
			name:        "Resolved filter",
			status:      "resolved",
			expectedIDs: []int64{3}, // Only resolved market
			description: "Should return only markets where isResolved=true",
		},
		{
			name:        "All filter",
			status:      "all",
			expectedIDs: []int64{1, 2, 3}, // All markets
			description: "Should return all markets regardless of status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SearchMarkets(db, "Test", tt.status, 10)
			assert.NoError(t, err, "Search failed for %s: %s", tt.name, tt.description)

			// Extract IDs from primary results
			var foundIDs []int64
			for _, marketOverview := range result.PrimaryResults {
				foundIDs = append(foundIDs, marketOverview.Market.ID)
			}

			// Verify we found the expected markets
			assert.ElementsMatch(t, tt.expectedIDs, foundIDs,
				"Market ID mismatch for %s: %s. Expected %v, got %v",
				tt.name, tt.description, tt.expectedIDs, foundIDs)
		})
	}
}
