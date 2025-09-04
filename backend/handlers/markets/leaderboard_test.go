package marketshandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestMarketLeaderboardHandler_InvalidMarketId(t *testing.T) {
	// Create a request with an invalid market ID
	req, err := http.NewRequest("GET", "/v0/markets/leaderboard/invalid", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create router and add the route
	router := mux.NewRouter()
	router.HandleFunc("/v0/markets/leaderboard/{marketId}", MarketLeaderboardHandler)

	// Serve the request
	router.ServeHTTP(rr, req)

	// Check that we get a bad request status
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	// Check that the response is JSON
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Handler returned wrong content type: got %v want %v", contentType, "application/json")
	}
}

func TestMarketLeaderboardHandler_ValidFormat(t *testing.T) {
	// This test would require database setup and test data
	// For now, we'll just test that the handler responds with proper JSON format
	// In a real test environment, you'd set up test database with known data

	t.Skip("Integration test requires database setup with test data")

	// Example of what the full test would look like:
	/*
		// Setup test database with known market and bet data
		testDB := setupTestDatabase()
		defer cleanupTestDatabase(testDB)

		// Create test market and bets
		marketId := createTestMarket(testDB)
		createTestBets(testDB, marketId)

		req, err := http.NewRequest("GET", fmt.Sprintf("/v0/markets/leaderboard/%d", marketId), nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/v0/markets/leaderboard/{marketId}", MarketLeaderboardHandler)
		router.ServeHTTP(rr, req)

		// Check status
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check response format
		var leaderboard []positionsmath.UserProfitability
		err = json.Unmarshal(rr.Body.Bytes(), &leaderboard)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Verify leaderboard properties
		if len(leaderboard) == 0 {
			t.Error("Expected non-empty leaderboard")
		}

		// Check that ranks are sequential
		for i, entry := range leaderboard {
			if entry.Rank != i+1 {
				t.Errorf("Expected rank %d, got %d", i+1, entry.Rank)
			}
		}

		// Check that profits are in descending order
		for i := 1; i < len(leaderboard); i++ {
			if leaderboard[i-1].Profit < leaderboard[i].Profit {
				t.Error("Leaderboard not sorted by profit descending")
			}
		}
	*/
}

func TestMarketLeaderboardHandler_EmptyResponse(t *testing.T) {
	// Test that handler properly returns empty array for market with no positions
	t.Skip("Integration test requires database setup")
}

// Helper function that would be used in real tests
func validateLeaderboardResponse(t *testing.T, responseBody []byte) {
	var leaderboard []map[string]interface{}
	err := json.Unmarshal(responseBody, &leaderboard)
	if err != nil {
		t.Fatalf("Failed to unmarshal leaderboard response: %v", err)
	}

	// Check required fields are present
	requiredFields := []string{"username", "currentValue", "totalSpent", "profit", "position", "yesSharesOwned", "noSharesOwned", "earliestBet", "rank"}

	for i, entry := range leaderboard {
		for _, field := range requiredFields {
			if _, exists := entry[field]; !exists {
				t.Errorf("Entry %d missing required field: %s", i, field)
			}
		}
	}
}
