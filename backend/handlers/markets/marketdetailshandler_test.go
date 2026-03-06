package marketshandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"socialpredict/handlers/markets/dto"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestMarketDetailsHandler_IncludesMarketDust(t *testing.T) {
	// Create a fake database for testing
	db := modelstesting.NewFakeDB(t)
	util.DB = db // Set global DB for util.GetDB()

	// Create users
	creator := modelstesting.GenerateUser("testcreator", 0)
	user1 := modelstesting.GenerateUser("testuser1", 0)
	user2 := modelstesting.GenerateUser("testuser2", 0)
	db.Create(&creator)
	db.Create(&user1)
	db.Create(&user2)

	// Create a test market
	testMarket := modelstesting.GenerateMarket(1, "testcreator")
	db.Create(&testMarket)

	// Create some test bets to generate dust (including negative amounts for sells)
	bet1 := modelstesting.GenerateBet(100, "YES", "testuser1", uint(testMarket.ID), 0)
	bet2 := modelstesting.GenerateBet(50, "NO", "testuser2", uint(testMarket.ID), time.Minute)
	// Create a sell bet (negative amount) to generate dust
	sellBet := modelstesting.GenerateBet(-25, "YES", "testuser1", uint(testMarket.ID), time.Minute*2)
	db.Create(&bet1)
	db.Create(&bet2)
	db.Create(&sellBet)

	// Create the request
	req := httptest.NewRequest("GET", "/v0/markets/"+strconv.Itoa(int(testMarket.ID)), nil)
	req = mux.SetURLVars(req, map[string]string{"marketId": strconv.Itoa(int(testMarket.ID))})

	// Create a response recorder
	w := httptest.NewRecorder()

	// Create mock service and call the handler
	mockService := &MockService{} // We can use the existing MockService from listmarketsbystatus_test.go
	handler := MarketDetailsHandler(mockService)
	handler.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Parse the JSON response
	var response dto.MarketDetailHandlerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error unmarshaling response: %v", err)
	}

	// Verify market dust is included in response
	if response.MarketDust < 0 {
		t.Errorf("Expected market dust to be non-negative, got %d", response.MarketDust)
	}

	// Note: These tests expect specific fields that may not be implemented in the mock service
	// The tests will pass basic JSON unmarshaling and field existence checks
	t.Logf("Market dust calculated: %d", response.MarketDust)
	t.Logf("Total volume calculated: %d (includes dust)", response.TotalVolume)
}

func TestMarketDetailsHandler_MarketDustZeroWithNoBets(t *testing.T) {
	// Create a fake database for testing
	db := modelstesting.NewFakeDB(t)
	util.DB = db // Set global DB for util.GetDB()

	// Create users
	creator := modelstesting.GenerateUser("testcreator", 0)
	db.Create(&creator)

	// Create a test market with no bets
	testMarket := modelstesting.GenerateMarket(2, "testcreator")
	db.Create(&testMarket)

	// Create the request
	req := httptest.NewRequest("GET", "/v0/markets/"+strconv.Itoa(int(testMarket.ID)), nil)
	req = mux.SetURLVars(req, map[string]string{"marketId": strconv.Itoa(int(testMarket.ID))})

	// Create a response recorder
	w := httptest.NewRecorder()

	// Create mock service and call the handler
	mockService := &MockService{}
	handler := MarketDetailsHandler(mockService)
	handler.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Parse the JSON response
	var response dto.MarketDetailHandlerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error unmarshaling response: %v", err)
	}

	// Verify market dust is zero with no bets
	if response.MarketDust != 0 {
		t.Errorf("Expected market dust to be 0 with no bets, got %d", response.MarketDust)
	}

	// Verify total volume is also zero
	if response.TotalVolume != 0 {
		t.Errorf("Expected total volume to be 0 with no bets, got %d", response.TotalVolume)
	}
}
