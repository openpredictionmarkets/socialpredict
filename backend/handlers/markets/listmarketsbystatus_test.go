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
)

func TestActiveMarketsFilter(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	util.DB = db

	// Create test data
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)
	pastTime := now.Add(-24 * time.Hour)

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 1000)
	db.Create(&testUser)

	// Active market (not resolved, future resolution date)
	activeMarket := models.Market{
		ID:                 1,
		QuestionTitle:      "Active Market",
		Description:        "Test active market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: futureTime,
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}

	// Closed market (not resolved, past resolution date)
	closedMarket := models.Market{
		ID:                 2,
		QuestionTitle:      "Closed Market",
		Description:        "Test closed market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: pastTime,
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}

	// Resolved market
	resolvedMarket := models.Market{
		ID:                      3,
		QuestionTitle:           "Resolved Market",
		Description:             "Test resolved market",
		OutcomeType:             "BINARY",
		ResolutionDateTime:      pastTime,
		FinalResolutionDateTime: pastTime,
		IsResolved:              true,
		ResolutionResult:        "YES",
		InitialProbability:      0.5,
		CreatorUsername:         "testuser",
	}

	// Insert test data
	db.Create(&activeMarket)
	db.Create(&closedMarket)
	db.Create(&resolvedMarket)

	// Test ActiveMarketsFilter
	var activeResults []models.Market
	ActiveMarketsFilter(db).Find(&activeResults)
	if len(activeResults) != 1 {
		t.Errorf("Expected 1 active market, got %d", len(activeResults))
	}
	if activeResults[0].QuestionTitle != "Active Market" {
		t.Errorf("Expected 'Active Market', got %s", activeResults[0].QuestionTitle)
	}
	if activeResults[0].IsResolved {
		t.Error("Expected market to not be resolved")
	}
	if !activeResults[0].ResolutionDateTime.After(now) {
		t.Error("Expected resolution date to be in the future")
	}
}

func TestClosedMarketsFilter(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	util.DB = db

	// Create test data
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)
	pastTime := now.Add(-24 * time.Hour)

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 1000)
	db.Create(&testUser)

	// Active market
	activeMarket := models.Market{
		ID:                 1,
		QuestionTitle:      "Active Market",
		Description:        "Test active market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: futureTime,
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}

	// Closed market
	closedMarket := models.Market{
		ID:                 2,
		QuestionTitle:      "Closed Market",
		Description:        "Test closed market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: pastTime,
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}

	// Insert test data
	db.Create(&activeMarket)
	db.Create(&closedMarket)

	// Test ClosedMarketsFilter
	var closedResults []models.Market
	ClosedMarketsFilter(db).Find(&closedResults)
	if len(closedResults) != 1 {
		t.Errorf("Expected 1 closed market, got %d", len(closedResults))
	}
	if closedResults[0].QuestionTitle != "Closed Market" {
		t.Errorf("Expected 'Closed Market', got %s", closedResults[0].QuestionTitle)
	}
	if closedResults[0].IsResolved {
		t.Error("Expected market to not be resolved")
	}
	if closedResults[0].ResolutionDateTime.After(now) {
		t.Error("Expected resolution date to be in the past")
	}
}

func TestResolvedMarketsFilter(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	util.DB = db

	// Create test data
	pastTime := time.Now().Add(-24 * time.Hour)

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 1000)
	db.Create(&testUser)

	// Unresolved market
	unresolvedMarket := models.Market{
		ID:                 1,
		QuestionTitle:      "Unresolved Market",
		Description:        "Test unresolved market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: pastTime,
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}

	// Resolved market
	resolvedMarket := models.Market{
		ID:                      2,
		QuestionTitle:           "Resolved Market",
		Description:             "Test resolved market",
		OutcomeType:             "BINARY",
		ResolutionDateTime:      pastTime,
		FinalResolutionDateTime: pastTime,
		IsResolved:              true,
		ResolutionResult:        "YES",
		InitialProbability:      0.5,
		CreatorUsername:         "testuser",
	}

	// Insert test data
	db.Create(&unresolvedMarket)
	db.Create(&resolvedMarket)

	// Test ResolvedMarketsFilter
	var resolvedResults []models.Market
	ResolvedMarketsFilter(db).Find(&resolvedResults)
	if len(resolvedResults) != 1 {
		t.Errorf("Expected 1 resolved market, got %d", len(resolvedResults))
	}
	if resolvedResults[0].QuestionTitle != "Resolved Market" {
		t.Errorf("Expected 'Resolved Market', got %s", resolvedResults[0].QuestionTitle)
	}
	if !resolvedResults[0].IsResolved {
		t.Error("Expected market to be resolved")
	}
	if resolvedResults[0].ResolutionResult != "YES" {
		t.Errorf("Expected resolution result 'YES', got %s", resolvedResults[0].ResolutionResult)
	}
}

func TestListMarketsByStatus(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	util.DB = db

	// Create test user first
	testUser := modelstesting.GenerateUser("testuser", 1000)
	db.Create(&testUser)

	// Create test data
	futureTime := time.Now().Add(24 * time.Hour)

	activeMarket := models.Market{
		ID:                 1,
		QuestionTitle:      "Active Market",
		Description:        "Test active market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: futureTime,
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}

	db.Create(&activeMarket)

	// Test ListMarketsByStatus with ActiveMarketsFilter
	markets, err := ListMarketsByStatus(db, ActiveMarketsFilter)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(markets) != 1 {
		t.Errorf("Expected 1 market, got %d", len(markets))
	}
	if markets[0].QuestionTitle != "Active Market" {
		t.Errorf("Expected 'Active Market', got %s", markets[0].QuestionTitle)
	}
}

func TestListMarketsByStatusWithEmptyResults(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	util.DB = db

	// Test with no markets in database
	markets, err := ListMarketsByStatus(db, ActiveMarketsFilter)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(markets) != 0 {
		t.Errorf("Expected 0 markets, got %d", len(markets))
	}
}

func TestListActiveMarketsHandler(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	util.DB = db

	// Create test user and market data
	testUser := modelstesting.GenerateUser("testuser", 1000)
	db.Create(&testUser)

	futureTime := time.Now().Add(24 * time.Hour)

	activeMarket := models.Market{
		ID:                 1,
		QuestionTitle:      "Active Market",
		Description:        "Test active market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: futureTime,
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}
	db.Create(&activeMarket)

	// Create HTTP request
	req, err := http.NewRequest("GET", "/v0/markets/active", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler := http.HandlerFunc(ListActiveMarketsHandler)
	handler.ServeHTTP(rr, req)

	// Check response status
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Parse response
	var response ListMarketsStatusResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error unmarshaling response: %v", err)
	}

	// Verify response structure
	if response.Status != "active" {
		t.Errorf("Expected status 'active', got %s", response.Status)
	}
	if response.Count != 1 {
		t.Errorf("Expected count 1, got %d", response.Count)
	}
	if len(response.Markets) != 1 {
		t.Errorf("Expected 1 market, got %d", len(response.Markets))
	}
}

func TestListClosedMarketsHandler(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	util.DB = db

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 1000)
	db.Create(&testUser)

	pastTime := time.Now().Add(-24 * time.Hour)

	closedMarket := models.Market{
		ID:                 1,
		QuestionTitle:      "Closed Market",
		Description:        "Test closed market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: pastTime,
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}
	db.Create(&closedMarket)

	// Create HTTP request
	req, err := http.NewRequest("GET", "/v0/markets/closed", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler := http.HandlerFunc(ListClosedMarketsHandler)
	handler.ServeHTTP(rr, req)

	// Check response status
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Parse response
	var response ListMarketsStatusResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error unmarshaling response: %v", err)
	}

	// Verify response structure
	if response.Status != "closed" {
		t.Errorf("Expected status 'closed', got %s", response.Status)
	}
	if response.Count != 1 {
		t.Errorf("Expected count 1, got %d", response.Count)
	}
	if len(response.Markets) != 1 {
		t.Errorf("Expected 1 market, got %d", len(response.Markets))
	}
}

func TestListResolvedMarketsHandler(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	util.DB = db

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 1000)
	db.Create(&testUser)

	pastTime := time.Now().Add(-24 * time.Hour)

	resolvedMarket := models.Market{
		ID:                      1,
		QuestionTitle:           "Resolved Market",
		Description:             "Test resolved market",
		OutcomeType:             "BINARY",
		ResolutionDateTime:      pastTime,
		FinalResolutionDateTime: pastTime,
		IsResolved:              true,
		ResolutionResult:        "YES",
		InitialProbability:      0.5,
		CreatorUsername:         "testuser",
	}
	db.Create(&resolvedMarket)

	// Create HTTP request
	req, err := http.NewRequest("GET", "/v0/markets/resolved", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler := http.HandlerFunc(ListResolvedMarketsHandler)
	handler.ServeHTTP(rr, req)

	// Check response status
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Parse response
	var response ListMarketsStatusResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error unmarshaling response: %v", err)
	}

	// Verify response structure
	if response.Status != "resolved" {
		t.Errorf("Expected status 'resolved', got %s", response.Status)
	}
	if response.Count != 1 {
		t.Errorf("Expected count 1, got %d", response.Count)
	}
	if len(response.Markets) != 1 {
		t.Errorf("Expected 1 market, got %d", len(response.Markets))
	}
}

func TestHandlerMethodNotAllowed(t *testing.T) {
	// Test POST method on GET-only endpoint
	req, err := http.NewRequest("POST", "/v0/markets/active", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ListActiveMarketsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, status)
	}
}
