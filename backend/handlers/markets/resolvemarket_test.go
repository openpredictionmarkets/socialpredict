package marketshandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
	"testing"

	"github.com/gorilla/mux"
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Set up test environment
	os.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	// Run tests
	code := m.Run()

	// Clean up
	os.Exit(code)
}

func TestResolveMarketHandler_NARefund(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	// Set the global DB for util.GetDB()
	util.DB = db

	// Create users
	creator := modelstesting.GenerateUser("creator", 0)
	bettor := modelstesting.GenerateUser("bettor", 0)
	db.Create(&creator)
	db.Create(&bettor)

	// Create market
	market := modelstesting.GenerateMarket(1, "creator")
	db.Create(&market)

	// Create bet
	bet := modelstesting.GenerateBet(100, "YES", "bettor", uint(market.ID), 0)
	db.Create(&bet)

	// Create JWT token for creator
	token := modelstesting.GenerateValidJWT("creator")

	// Create request body
	reqBody := map[string]string{"outcome": "N/A"}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/v0/market/1/resolve", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with URL vars
	router := mux.NewRouter()
	router.HandleFunc("/v0/market/{marketId}/resolve", ResolveMarketHandler).Methods("POST")
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify market is resolved
	var resolvedMarket models.Market
	db.First(&resolvedMarket, market.ID)
	if !resolvedMarket.IsResolved {
		t.Fatal("Market should be resolved")
	}
	if resolvedMarket.ResolutionResult != "N/A" {
		t.Fatalf("Expected resolution result N/A, got %s", resolvedMarket.ResolutionResult)
	}

	// Verify bettor received refund
	var updatedBettor models.User
	db.Where("username = ?", "bettor").First(&updatedBettor)
	if updatedBettor.AccountBalance != 100 {
		t.Fatalf("Expected bettor balance 100 after refund, got %d", updatedBettor.AccountBalance)
	}
}

func TestResolveMarketHandler_YESWin(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	// Set the global DB for util.GetDB()
	util.DB = db

	// Create users
	creator := modelstesting.GenerateUser("creator", 0)
	winner := modelstesting.GenerateUser("winner", 0)
	loser := modelstesting.GenerateUser("loser", 0)
	db.Create(&creator)
	db.Create(&winner)
	db.Create(&loser)

	// Create market
	market := modelstesting.GenerateMarket(2, "creator")
	db.Create(&market)

	// Create bets
	winningBet := modelstesting.GenerateBet(100, "YES", "winner", uint(market.ID), 0)
	losingBet := modelstesting.GenerateBet(100, "NO", "loser", uint(market.ID), 0)
	db.Create(&winningBet)
	db.Create(&losingBet)

	// Create JWT token for creator
	token := modelstesting.GenerateValidJWT("creator")

	// Create request body
	reqBody := map[string]string{"outcome": "YES"}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/v0/market/2/resolve", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with URL vars
	router := mux.NewRouter()
	router.HandleFunc("/v0/market/{marketId}/resolve", ResolveMarketHandler).Methods("POST")
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify market is resolved
	var resolvedMarket models.Market
	db.First(&resolvedMarket, market.ID)
	if !resolvedMarket.IsResolved {
		t.Fatal("Market should be resolved")
	}
	if resolvedMarket.ResolutionResult != "YES" {
		t.Fatalf("Expected resolution result YES, got %s", resolvedMarket.ResolutionResult)
	}

	// Verify winner got more than loser (proportional payout)
	var updatedWinner, updatedLoser models.User
	db.Where("username = ?", "winner").First(&updatedWinner)
	db.Where("username = ?", "loser").First(&updatedLoser)

	if updatedWinner.AccountBalance <= updatedLoser.AccountBalance {
		t.Fatalf("Expected winner balance (%d) to be greater than loser balance (%d)", updatedWinner.AccountBalance, updatedLoser.AccountBalance)
	}

	// The total payouts should equal the market volume (200 total bet amount)
	totalPayout := updatedWinner.AccountBalance + updatedLoser.AccountBalance
	if totalPayout != 200 {
		t.Fatalf("Expected total payout to be 200, got %d", totalPayout)
	}
}

func TestResolveMarketHandler_NOWin(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	// Set the global DB for util.GetDB()
	util.DB = db

	// Create users
	creator := modelstesting.GenerateUser("creator", 0)
	winner := modelstesting.GenerateUser("winner", 0)
	loser := modelstesting.GenerateUser("loser", 0)
	db.Create(&creator)
	db.Create(&winner)
	db.Create(&loser)

	// Create market
	market := modelstesting.GenerateMarket(3, "creator")
	db.Create(&market)

	// Create bets
	losingBet := modelstesting.GenerateBet(100, "YES", "loser", uint(market.ID), 0)
	winningBet := modelstesting.GenerateBet(100, "NO", "winner", uint(market.ID), 0)
	db.Create(&losingBet)
	db.Create(&winningBet)

	// Create JWT token for creator
	token := modelstesting.GenerateValidJWT("creator")

	// Create request body
	reqBody := map[string]string{"outcome": "NO"}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/v0/market/3/resolve", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with URL vars
	router := mux.NewRouter()
	router.HandleFunc("/v0/market/{marketId}/resolve", ResolveMarketHandler).Methods("POST")
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify market is resolved
	var resolvedMarket models.Market
	db.First(&resolvedMarket, market.ID)
	if !resolvedMarket.IsResolved {
		t.Fatal("Market should be resolved")
	}
	if resolvedMarket.ResolutionResult != "NO" {
		t.Fatalf("Expected resolution result NO, got %s", resolvedMarket.ResolutionResult)
	}

	// Verify winner got more than loser (proportional payout)
	var updatedWinner, updatedLoser models.User
	db.Where("username = ?", "winner").First(&updatedWinner)
	db.Where("username = ?", "loser").First(&updatedLoser)

	if updatedWinner.AccountBalance <= updatedLoser.AccountBalance {
		t.Fatalf("Expected winner balance (%d) to be greater than loser balance (%d)", updatedWinner.AccountBalance, updatedLoser.AccountBalance)
	}

	// The total payouts should equal the market volume (200 total bet amount)
	totalPayout := updatedWinner.AccountBalance + updatedLoser.AccountBalance
	if totalPayout != 200 {
		t.Fatalf("Expected total payout to be 200, got %d", totalPayout)
	}
}

func TestResolveMarketHandler_UnauthorizedUser(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	// Set the global DB for util.GetDB()
	util.DB = db

	// Create users
	creator := modelstesting.GenerateUser("creator", 0)
	otherUser := modelstesting.GenerateUser("other", 0)
	db.Create(&creator)
	db.Create(&otherUser)

	// Create market
	market := modelstesting.GenerateMarket(4, "creator")
	db.Create(&market)

	// Create JWT token for non-creator
	token := modelstesting.GenerateValidJWT("other")

	// Create request body
	reqBody := map[string]string{"outcome": "YES"}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/v0/market/4/resolve", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with URL vars
	router := mux.NewRouter()
	router.HandleFunc("/v0/market/{marketId}/resolve", ResolveMarketHandler).Methods("POST")
	router.ServeHTTP(w, req)

	// Check response - should be unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401, got %d", w.Code)
	}

	// Verify market is not resolved
	var market_check models.Market
	db.First(&market_check, market.ID)
	if market_check.IsResolved {
		t.Fatal("Market should not be resolved")
	}
}

func TestResolveMarketHandler_InvalidOutcome(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	// Set the global DB for util.GetDB()
	util.DB = db

	// Create user
	creator := modelstesting.GenerateUser("creator", 0)
	db.Create(&creator)

	// Create market
	market := modelstesting.GenerateMarket(5, "creator")
	db.Create(&market)

	// Create JWT token for creator
	token := modelstesting.GenerateValidJWT("creator")

	// Create request body with invalid outcome
	reqBody := map[string]string{"outcome": "MAYBE"}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/v0/market/5/resolve", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with URL vars
	router := mux.NewRouter()
	router.HandleFunc("/v0/market/{marketId}/resolve", ResolveMarketHandler).Methods("POST")
	router.ServeHTTP(w, req)

	// Check response - should be bad request
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}

	// Verify market is not resolved
	var market_check models.Market
	db.First(&market_check, market.ID)
	if market_check.IsResolved {
		t.Fatal("Market should not be resolved")
	}
}
