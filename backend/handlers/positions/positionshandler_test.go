package positions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	positionsmath "socialpredict/handlers/math/positions"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"github.com/gorilla/mux"
)

type mockPositionsService struct {
	positions dmarkets.MarketPositions
	err       error
}

func (m *mockPositionsService) CreateMarket(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *mockPositionsService) SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error {
	return nil
}

func (m *mockPositionsService) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *mockPositionsService) ListMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	return nil, nil
}

func (m *mockPositionsService) SearchMarkets(ctx context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.SearchResults, error) {
	return nil, nil
}

func (m *mockPositionsService) ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error {
	return nil
}

func (m *mockPositionsService) ListByStatus(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error) {
	return nil, nil
}

func (m *mockPositionsService) GetMarketLeaderboard(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
	return nil, nil
}

func (m *mockPositionsService) ProjectProbability(ctx context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	return nil, nil
}

func (m *mockPositionsService) GetMarketDetails(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error) {
	return nil, nil
}

func (m *mockPositionsService) GetMarketBets(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error) {
	return nil, nil
}

func (m *mockPositionsService) GetMarketPositions(ctx context.Context, marketID int64) (dmarkets.MarketPositions, error) {
	return m.positions, m.err
}

func (m *mockPositionsService) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (dmarkets.UserPosition, error) {
	return nil, nil
}

func TestMarketPositionsHandlerWithService_IncludesZeroPositionUsers(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)

	creator := modelstesting.GenerateUser("creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(9001, creator.Username)
	market.IsResolved = true
	market.ResolutionResult = "YES"
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	users := []string{"patrick", "jimmy", "jyron", "testuser03"}
	for _, username := range users {
		user := modelstesting.GenerateUser(username, 0)
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user %s: %v", username, err)
		}
	}

	bets := []struct {
		amount   int64
		outcome  string
		username string
		offset   time.Duration
	}{
		{50, "NO", "patrick", 0},
		{51, "NO", "jimmy", time.Second},
		{51, "NO", "jimmy", 2 * time.Second},
		{10, "YES", "jyron", 3 * time.Second},
		{30, "YES", "testuser03", 4 * time.Second},
	}

	for _, b := range bets {
		bet := modelstesting.GenerateBet(b.amount, b.outcome, b.username, uint(market.ID), b.offset)
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	marketIDStr := strconv.FormatInt(market.ID, 10)
	positionSnapshot, err := positionsmath.CalculateMarketPositions_WPAM_DBPM(db, marketIDStr)
	if err != nil {
		t.Fatalf("calculate positions: %v", err)
	}

	mockSvc := &mockPositionsService{positions: positionSnapshot}
	handler := MarketPositionsHandlerWithService(mockSvc)

	req := httptest.NewRequest("GET", "/v0/markets/positions/"+marketIDStr, nil)
	req = mux.SetURLVars(req, map[string]string{
		"marketId": marketIDStr,
	})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var positions []positionsmath.MarketPosition
	if err := json.Unmarshal(rec.Body.Bytes(), &positions); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	var locked *positionsmath.MarketPosition
	for i := range positions {
		if positions[i].Username == "testuser03" {
			locked = &positions[i]
			break
		}
	}

	if locked == nil {
		t.Fatalf("expected locked bettor to be present in handler response: %+v", positions)
	}

	if locked.YesSharesOwned != 0 || locked.NoSharesOwned != 0 || locked.Value != 0 {
		t.Fatalf("expected zero-valued position for locked bettor, got %+v", locked)
	}

	var totals models.Bet
	if err := db.Where("username = ? AND market_id = ?", "testuser03", market.ID).First(&totals).Error; err != nil {
		t.Fatalf("verify bets: %v", err)
	}
}
