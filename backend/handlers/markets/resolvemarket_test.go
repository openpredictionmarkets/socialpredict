package marketshandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
	"testing"

	"github.com/gorilla/mux"
)

// MockResolveService for testing - implements dmarkets.ServiceInterface
type MockResolveService struct{}

func (m *MockResolveService) CreateMarket(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error) {
	return nil, nil
}
func (m *MockResolveService) SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error {
	return nil
}
func (m *MockResolveService) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	return nil, nil
}
func (m *MockResolveService) ListMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	return nil, nil
}
func (m *MockResolveService) SearchMarkets(ctx context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.SearchResults, error) {
	return &dmarkets.SearchResults{
		PrimaryResults:  []*dmarkets.Market{},
		FallbackResults: []*dmarkets.Market{},
		Query:           query,
		PrimaryStatus:   filters.Status,
		PrimaryCount:    0,
		FallbackCount:   0,
		TotalCount:      0,
		FallbackUsed:    false,
	}, nil
}
func (m *MockResolveService) ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error {
	// Mock implementation that checks authorization and valid outcomes
	if username != "creator" {
		return dmarkets.ErrUnauthorized
	}
	if resolution != "YES" && resolution != "NO" && resolution != "N/A" {
		return dmarkets.ErrInvalidInput
	}
	return nil
}
func (m *MockResolveService) ListByStatus(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error) {
	return nil, nil
}
func (m *MockResolveService) GetMarketLeaderboard(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
	return nil, nil
}
func (m *MockResolveService) ProjectProbability(ctx context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	return nil, nil
}
func (m *MockResolveService) GetMarketDetails(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error) {
	return nil, nil
}

func (m *MockResolveService) GetMarketBets(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error) {
	return []*dmarkets.BetDisplayInfo{}, nil
}

func (m *MockResolveService) GetMarketPositions(ctx context.Context, marketID int64) (dmarkets.MarketPositions, error) {
	return nil, nil
}

func (m *MockResolveService) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (dmarkets.UserPosition, error) {
	return nil, nil
}

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

	// Create request body (using "resolution" field as per DTO)
	reqBody := map[string]string{"resolution": "N/A"}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/v0/market/1/resolve", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with mock service
	mockService := &MockResolveService{}
	router := mux.NewRouter()
	router.HandleFunc("/v0/market/{marketId}/resolve", ResolveMarketHandler(mockService)).Methods("POST")
	router.ServeHTTP(w, req)

	// Check response - updated for specification compliance (NoContent instead of OK)
	if w.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Note: The actual resolution logic (updating DB, processing refunds) would be tested separately
	// This test verifies the HTTP layer works correctly with service injection
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

	// Create request body (using "resolution" field as per DTO)
	reqBody := map[string]string{"resolution": "YES"}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/v0/market/2/resolve", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with mock service
	mockService := &MockResolveService{}
	router := mux.NewRouter()
	router.HandleFunc("/v0/market/{marketId}/resolve", ResolveMarketHandler(mockService)).Methods("POST")
	router.ServeHTTP(w, req)

	// Check response - updated for specification compliance (NoContent instead of OK)
	if w.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Note: The actual payout logic would be tested in the domain service layer
	// This test verifies the HTTP layer works correctly with service injection
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

	// Create request body (using "resolution" field as per DTO)
	reqBody := map[string]string{"resolution": "NO"}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/v0/market/3/resolve", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with mock service
	mockService := &MockResolveService{}
	router := mux.NewRouter()
	router.HandleFunc("/v0/market/{marketId}/resolve", ResolveMarketHandler(mockService)).Methods("POST")
	router.ServeHTTP(w, req)

	// Check response - updated for specification compliance (NoContent instead of OK)
	if w.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Note: The actual payout logic would be tested in the domain service layer
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

	// Create request body (using "resolution" field as per DTO)
	reqBody := map[string]string{"resolution": "YES"}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/v0/market/4/resolve", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with mock service
	mockService := &MockResolveService{}
	router := mux.NewRouter()
	router.HandleFunc("/v0/market/{marketId}/resolve", ResolveMarketHandler(mockService)).Methods("POST")
	router.ServeHTTP(w, req)

	// Check response - updated for specification compliance (403 Forbidden instead of 401)
	if w.Code != http.StatusForbidden {
		t.Fatalf("Expected status 403, got %d", w.Code)
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

	// Create request body with invalid resolution (using "resolution" field as per DTO)
	reqBody := map[string]string{"resolution": "MAYBE"}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/v0/market/5/resolve", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with mock service
	mockService := &MockResolveService{}
	router := mux.NewRouter()
	router.HandleFunc("/v0/market/{marketId}/resolve", ResolveMarketHandler(mockService)).Methods("POST")
	router.ServeHTTP(w, req)

	// Check response - should be bad request
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}
}
