package marketshandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// TestMarketDetailsHandler_VolumeConsistencyFix verifies the fix for the logical inconsistency
// where market volume could be 0 while dust was > 0, which doesn't make mathematical sense
func TestMarketDetailsHandler_VolumeConsistencyFix(t *testing.T) {
	// Create a fake database for testing
	db := modelstesting.NewFakeDB(t)
	util.DB = db // Set global DB for util.GetDB()

	// Create users
	creator := modelstesting.GenerateUser("creator", 0)
	trader := modelstesting.GenerateUser("trader", 0)
	db.Create(&creator)
	db.Create(&trader)

	// Create a test market
	testMarket := modelstesting.GenerateMarket(1, "creator")
	db.Create(&testMarket)

	// Reproduce the scenario that caused the inconsistency:
	// User buys shares, then sells all shares back
	buyBet := modelstesting.GenerateBet(100, "YES", "trader", uint(testMarket.ID), 0)
	sellBet := modelstesting.GenerateBet(-100, "YES", "trader", uint(testMarket.ID), time.Minute)
	db.Create(&buyBet)
	db.Create(&sellBet)

	// Create the request
	req := httptest.NewRequest("GET", "/api/v0/markets/"+strconv.Itoa(int(testMarket.ID)), nil)
	req = mux.SetURLVars(req, map[string]string{"marketId": strconv.Itoa(int(testMarket.ID))})

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	MarketDetailsHandler(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Parse the JSON response
	var response MarketDetailHandlerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error unmarshaling response: %v", err)
	}

	// Log the actual values for verification
	t.Logf("Total volume (with dust): %d", response.TotalVolume)
	t.Logf("Market dust: %d", response.MarketDust)

	// Verify logical consistency: if dust > 0, then volume must be >= dust
	if response.MarketDust > 0 && response.TotalVolume < response.MarketDust {
		t.Errorf("Logical inconsistency: dust (%d) cannot be greater than total volume (%d)",
			response.MarketDust, response.TotalVolume)
	}

	// With the fix, we expect:
	// - Net betting volume: 100 - 100 = 0
	// - Dust from sell: 1
	// - Total volume (liquidity remaining): 0 + 1 = 1
	expectedVolume := int64(1) // 0 net + 1 dust
	expectedDust := int64(1)   // 1 dust from the sell

	if response.TotalVolume != expectedVolume {
		t.Errorf("Expected total volume to be %d (0 net + 1 dust), got %d", expectedVolume, response.TotalVolume)
	}

	if response.MarketDust != expectedDust {
		t.Errorf("Expected market dust to be %d, got %d", expectedDust, response.MarketDust)
	}

	// Verify the relationship: volume should equal net bets + dust
	// This ensures mathematical consistency
	netBets := int64(0) // 100 - 100 = 0
	expectedTotalVolume := netBets + response.MarketDust
	if response.TotalVolume != expectedTotalVolume {
		t.Errorf("Volume inconsistency: expected %d (net bets) + %d (dust) = %d, got %d",
			netBets, response.MarketDust, expectedTotalVolume, response.TotalVolume)
	}
}

// TestMarketDetailsHandler_NoInconsistencyWithOnlyBuys verifies behavior with only buy transactions
func TestMarketDetailsHandler_NoInconsistencyWithOnlyBuys(t *testing.T) {
	// Create a fake database for testing
	db := modelstesting.NewFakeDB(t)
	util.DB = db // Set global DB for util.GetDB()

	// Create users
	creator := modelstesting.GenerateUser("creator", 0)
	trader1 := modelstesting.GenerateUser("trader1", 0)
	trader2 := modelstesting.GenerateUser("trader2", 0)
	db.Create(&creator)
	db.Create(&trader1)
	db.Create(&trader2)

	// Create a test market
	testMarket := modelstesting.GenerateMarket(2, "creator")
	db.Create(&testMarket)

	// Create only buy bets (no sells, so no dust)
	bet1 := modelstesting.GenerateBet(100, "YES", "trader1", uint(testMarket.ID), 0)
	bet2 := modelstesting.GenerateBet(50, "NO", "trader2", uint(testMarket.ID), time.Minute)
	db.Create(&bet1)
	db.Create(&bet2)

	// Create the request
	req := httptest.NewRequest("GET", "/api/v0/markets/"+strconv.Itoa(int(testMarket.ID)), nil)
	req = mux.SetURLVars(req, map[string]string{"marketId": strconv.Itoa(int(testMarket.ID))})

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	MarketDetailsHandler(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Parse the JSON response
	var response MarketDetailHandlerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error unmarshaling response: %v", err)
	}

	// With only buys, no dust should be generated
	expectedDust := int64(0)
	expectedVolume := int64(150) // 100 + 50, no dust to add

	if response.MarketDust != expectedDust {
		t.Errorf("Expected market dust to be %d with only buys, got %d", expectedDust, response.MarketDust)
	}

	if response.TotalVolume != expectedVolume {
		t.Errorf("Expected total volume to be %d, got %d", expectedVolume, response.TotalVolume)
	}

	t.Logf("No-sell scenario - Total volume: %d, Market dust: %d", response.TotalVolume, response.MarketDust)
}
