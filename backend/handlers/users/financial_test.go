package usershandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"socialpredict/models/modelstesting"
	"socialpredict/setup"
	"testing"

	"github.com/gorilla/mux"
)

func TestGetUserFinancialHandler_ValidUser(t *testing.T) {
	// Set up test database and user
	db := modelstesting.NewFakeDB(t)
	user := modelstesting.GenerateUser("testuser", 1000)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create request with username parameter
	req := httptest.NewRequest(http.MethodGet, "/v0/users/testuser/financial", nil)

	// Use Gorilla mux to handle path parameters
	vars := map[string]string{"username": "testuser"}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	// Create mock config loader using modelstesting helper
	mockConfigLoader := func() (*setup.EconomicConfig, error) {
		return modelstesting.GenerateEconomicConfig(), nil
	}

	// Use the testable handler with injected database
	handler := GetUserFinancialHandlerWithDB(db, mockConfigLoader)
	handler(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type application/json, got %s", contentType)
	}

	// Parse response body
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Verify response structure
	financialData, ok := response["financial"].(map[string]interface{})
	if !ok {
		t.Fatal("Response should contain 'financial' object")
	}

	// Check required fields
	requiredFields := []string{
		"accountBalance", "maximumDebtAllowed", "amountInPlay", "amountBorrowed",
		"retainedEarnings", "equity", "tradingProfits", "workProfits", "totalProfits",
		"amountInPlayActive", "totalSpent", "totalSpentInPlay", "realizedProfits",
		"potentialProfits", "realizedValue", "potentialValue",
	}

	for _, field := range requiredFields {
		if _, exists := financialData[field]; !exists {
			t.Errorf("Missing required field: %s", field)
		}
	}

	// Verify specific values for clean user
	if financialData["accountBalance"] != float64(1000) {
		t.Errorf("Expected accountBalance 1000, got %v", financialData["accountBalance"])
	}
	if financialData["amountInPlay"] != float64(0) {
		t.Errorf("Expected amountInPlay 0 for new user, got %v", financialData["amountInPlay"])
	}
}

func TestGetUserFinancialHandler_InvalidMethod(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	mockConfigLoader := func() (*setup.EconomicConfig, error) {
		return modelstesting.GenerateEconomicConfig(), nil
	}

	req := httptest.NewRequest(http.MethodPost, "/v0/users/testuser/financial", nil)
	w := httptest.NewRecorder()

	handler := GetUserFinancialHandlerWithDB(db, mockConfigLoader)
	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestGetUserFinancialHandler_MissingUsername(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	mockConfigLoader := func() (*setup.EconomicConfig, error) {
		return modelstesting.GenerateEconomicConfig(), nil
	}

	req := httptest.NewRequest(http.MethodGet, "/v0/users//financial", nil)
	w := httptest.NewRecorder()

	handler := GetUserFinancialHandlerWithDB(db, mockConfigLoader)
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetUserFinancialHandler_NonexistentUser(t *testing.T) {
	// Set up test database (no user created)
	db := modelstesting.NewFakeDB(t)
	mockConfigLoader := func() (*setup.EconomicConfig, error) {
		return modelstesting.GenerateEconomicConfig(), nil
	}

	req := httptest.NewRequest(http.MethodGet, "/v0/users/nonexistent/financial", nil)
	vars := map[string]string{"username": "nonexistent"}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	handler := GetUserFinancialHandlerWithDB(db, mockConfigLoader)
	handler(w, req)

	// The response should be successful but with zero balances
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}
