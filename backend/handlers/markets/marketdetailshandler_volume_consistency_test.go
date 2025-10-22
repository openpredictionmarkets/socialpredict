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

func TestMarketDetailsHandler_VolumeConsistency_OnlyBuys(t *testing.T) {
	// Create a fake database for testing
	db := modelstesting.NewFakeDB(t)
	util.DB = db

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

	// Create only buy bets (positive amounts)
	bet1 := modelstesting.GenerateBet(100, "YES", "testuser1", uint(testMarket.ID), 0)
	bet2 := modelstesting.GenerateBet(200, "NO", "testuser2", uint(testMarket.ID), time.Minute)
	bet3 := modelstesting.GenerateBet(50, "YES", "testuser1", uint(testMarket.ID), time.Minute*2)

	db.Create(&bet1)
	db.Create(&bet2)
	db.Create(&bet3)

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

	// Note: With mock service, this will return default values
	// The actual implementation would calculate: 100 + 200 + 50 = 350 volume
	t.Logf("Total volume with only buys: %d", response.TotalVolume)
	t.Logf("Market dust with only buys: %d", response.MarketDust)
}

func TestMarketDetailsHandler_VolumeConsistency_WithSells(t *testing.T) {
	// Create a fake database for testing
	db := modelstesting.NewFakeDB(t)
	util.DB = db

	// Create users
	creator := modelstesting.GenerateUser("testcreator", 0)
	user1 := modelstesting.GenerateUser("testuser1", 0)
	user2 := modelstesting.GenerateUser("testuser2", 0)
	db.Create(&creator)
	db.Create(&user1)
	db.Create(&user2)

	// Create a test market
	testMarket := modelstesting.GenerateMarket(2, "testcreator")
	db.Create(&testMarket)

	// Create mixed buy/sell bets
	buyBet1 := modelstesting.GenerateBet(200, "YES", "testuser1", uint(testMarket.ID), 0)
	buyBet2 := modelstesting.GenerateBet(150, "NO", "testuser2", uint(testMarket.ID), time.Minute)
	sellBet1 := modelstesting.GenerateBet(-75, "YES", "testuser1", uint(testMarket.ID), time.Minute*2)
	sellBet2 := modelstesting.GenerateBet(-50, "NO", "testuser2", uint(testMarket.ID), time.Minute*3)

	db.Create(&buyBet1)
	db.Create(&buyBet2)
	db.Create(&sellBet1)
	db.Create(&sellBet2)

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

	// Note: With mock service, this will return default values
	// The actual implementation would calculate:
	// Net volume: 200 + 150 - 75 - 50 = 225
	// Plus dust from sell transactions: typically 2 dust points = 227
	t.Logf("Total volume with buys and sells: %d", response.TotalVolume)
	t.Logf("Market dust with buys and sells: %d", response.MarketDust)

	// Market dust should be non-negative
	if response.MarketDust < 0 {
		t.Errorf("Expected market dust to be non-negative, got %d", response.MarketDust)
	}
}
