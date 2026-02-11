package positions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"github.com/gorilla/mux"
)

type mockPositionsService struct {
	positions dmarkets.MarketPositions
	err       error
}

func toDomainPositions(input []positionsmath.MarketPosition) dmarkets.MarketPositions {
	out := make(dmarkets.MarketPositions, 0, len(input))
	for _, p := range input {
		out = append(out, &dmarkets.UserPosition{
			Username:         p.Username,
			MarketID:         int64(p.MarketID),
			YesSharesOwned:   p.YesSharesOwned,
			NoSharesOwned:    p.NoSharesOwned,
			Value:            p.Value,
			TotalSpent:       p.TotalSpent,
			TotalSpentInPlay: p.TotalSpentInPlay,
			IsResolved:       p.IsResolved,
			ResolutionResult: p.ResolutionResult,
		})
	}
	return out
}

func (m *mockPositionsService) GetMarketPositions(ctx context.Context, marketID int64) (dmarkets.MarketPositions, error) {
	return m.positions, m.err
}

func (m *mockPositionsService) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	return nil, nil
}

func TestMarketPositionsHandlerWithService_IncludesZeroPositionUsers(t *testing.T) {
	modelstesting.SeedWPAMFromConfig(modelstesting.GenerateEconomicConfig())

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
	var marketModel models.Market
	if err := db.First(&marketModel, market.ID).Error; err != nil {
		t.Fatalf("reload market: %v", err)
	}

	var betsRecords []models.Bet
	if err := db.Where("market_id = ?", market.ID).Order("placed_at ASC").Find(&betsRecords).Error; err != nil {
		t.Fatalf("load bets: %v", err)
	}

	snapshot := positionsmath.MarketSnapshot{
		ID:               int64(marketModel.ID),
		CreatedAt:        marketModel.CreatedAt,
		IsResolved:       marketModel.IsResolved,
		ResolutionResult: marketModel.ResolutionResult,
	}

	positionSnapshot, err := positionsmath.CalculateMarketPositions_WPAM_DBPM(snapshot, betsRecords)
	if err != nil {
		t.Fatalf("calculate positions: %v", err)
	}

	mockSvc := &mockPositionsService{positions: toDomainPositions(positionSnapshot)}
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
