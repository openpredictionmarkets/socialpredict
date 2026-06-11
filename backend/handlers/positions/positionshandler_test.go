package positions

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"socialpredict/handlers"
	"socialpredict/internal/domain/boundary"
	dmarkets "socialpredict/internal/domain/markets"
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"github.com/gorilla/mux"
)

type mockPositionsService struct {
	positions     dmarkets.MarketPositions
	positionsPage dmarkets.MarketPositions
	err           error
	page          *dmarkets.Page
}

func boundaryBetsFromModels(dbBets []models.Bet) []boundary.Bet {
	bets := make([]boundary.Bet, len(dbBets))
	for i, bet := range dbBets {
		bets[i] = boundary.Bet{
			ID:        uint(bet.ID),
			Username:  bet.Username,
			MarketID:  bet.MarketID,
			Amount:    bet.Amount,
			Outcome:   bet.Outcome,
			PlacedAt:  bet.PlacedAt,
			CreatedAt: bet.CreatedAt,
		}
	}
	return bets
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

func (m *mockPositionsService) GetMarketBetsPage(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.BetDisplayInfo, error) {
	return nil, nil
}

func (m *mockPositionsService) GetMarketPositions(ctx context.Context, marketID int64) (dmarkets.MarketPositions, error) {
	return m.positions, m.err
}

func (m *mockPositionsService) GetMarketPositionsPage(ctx context.Context, marketID int64, p dmarkets.Page) (dmarkets.MarketPositions, error) {
	m.page = &p
	if m.err != nil {
		return nil, m.err
	}
	return m.positionsPage, nil
}

func (m *mockPositionsService) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	return nil, nil
}

func (m *mockPositionsService) CalculateMarketVolume(ctx context.Context, marketID int64) (int64, error) {
	return 0, nil
}

func (m *mockPositionsService) GetPublicMarket(ctx context.Context, marketID int64) (*dmarkets.PublicMarket, error) {
	return &dmarkets.PublicMarket{ID: marketID}, nil
}

func TestMarketPositionsHandlerWithService_IncludesZeroPositionUsers(t *testing.T) {
	_ = modelstesting.SeedWPAMFromConfig(modelstesting.GenerateEconomicConfig())

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

	positionSnapshot, err := positionsmath.CalculateMarketPositions_WPAM_DBPM(snapshot, boundaryBetsFromModels(betsRecords))
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

	var envelope handlers.SuccessEnvelope[[]userPositionResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !envelope.OK {
		t.Fatalf("expected ok=true, got false")
	}

	positions := envelope.Result

	var locked *userPositionResponse
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

func TestMarketPositionsHandlerWithService_FailureEnvelope(t *testing.T) {
	handler := MarketPositionsHandlerWithService(&mockPositionsService{err: dmarkets.ErrMarketNotFound})

	req := httptest.NewRequest(http.MethodGet, "/v0/markets/positions/77", nil)
	req = mux.SetURLVars(req, map[string]string{"marketId": "77"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if resp.Reason != string(handlers.ReasonMarketNotFound) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonMarketNotFound, resp.Reason)
	}
}

func TestMarketPositionsHandlerWithService_UsesPaginationQuery(t *testing.T) {
	mockSvc := &mockPositionsService{
		positionsPage: dmarkets.MarketPositions{
			{Username: "alice", MarketID: 7, YesSharesOwned: 3},
		},
	}
	handler := MarketPositionsHandlerWithService(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/v0/markets/positions/7?limit=20&offset=40", nil)
	req = mux.SetURLVars(req, map[string]string{"marketId": "7"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if mockSvc.page == nil {
		t.Fatalf("expected paginated service method to be called")
	}
	if mockSvc.page.Limit != 20 || mockSvc.page.Offset != 40 {
		t.Fatalf("expected page limit=20 offset=40, got %+v", *mockSvc.page)
	}
}

func TestMarketPositionsHandlerWithService_UsesSnapshotFreshnessWhenAvailable(t *testing.T) {
	generatedAt := time.Now().UTC().Add(-3 * time.Minute)
	mockSvc := &readModelPositionsService{
		mockPositionsService: mockPositionsService{
			positionsPage: dmarkets.MarketPositions{
				{Username: "raw_alice", MarketID: 7, YesSharesOwned: 99},
			},
		},
		readModel: &dmarkets.MarketPositionsSnapshot{
			MarketID:    7,
			GeneratedAt: generatedAt,
			Source:      "read_model",
			Positions: dmarkets.MarketPositions{
				{Username: "snapshot_alice", MarketID: 7, YesSharesOwned: 3},
			},
		},
	}
	handler := MarketPositionsHandlerWithService(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/v0/markets/positions/7?limit=20&offset=0", nil)
	req = mux.SetURLVars(req, map[string]string{"marketId": "7"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if mockSvc.page != nil {
		t.Fatalf("raw paginated position calculation should not be used when snapshot is available")
	}

	var resp handlers.SuccessEnvelope[marketPositionsResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode success envelope: %v", err)
	}
	if len(resp.Result.Positions) != 1 || resp.Result.Positions[0].Username != "snapshot_alice" {
		t.Fatalf("unexpected positions payload: %+v", resp.Result.Positions)
	}
	if resp.Result.Freshness == nil {
		t.Fatalf("expected freshness metadata")
	}
	if !resp.Result.Freshness.GeneratedAt.Equal(generatedAt) {
		t.Fatalf("freshness generatedAt = %s, want %s", resp.Result.Freshness.GeneratedAt, generatedAt)
	}
	if resp.Result.Freshness.TargetFreshnessSeconds != int(dmarkets.MarketPositionsSnapshotTargetFreshness.Seconds()) {
		t.Fatalf("freshness target = %d, want %d", resp.Result.Freshness.TargetFreshnessSeconds, int(dmarkets.MarketPositionsSnapshotTargetFreshness.Seconds()))
	}
	if resp.Result.Freshness.TransactionSafeRead {
		t.Fatalf("positions snapshot must not be marked transaction safe")
	}
}

func TestMarketPositionsHandlerWithService_ServesStaleSnapshotWithoutRefresh(t *testing.T) {
	generatedAt := time.Now().UTC().Add(-2 * dmarkets.MarketPositionsSnapshotTargetFreshness)
	mockSvc := &readModelPositionsService{
		readModel: &dmarkets.MarketPositionsSnapshot{
			MarketID:    7,
			GeneratedAt: generatedAt,
			Source:      "read_model",
			IsStale:     true,
			StaleReason: "bet_accepted",
			Positions: dmarkets.MarketPositions{
				{Username: "stale_snapshot_alice", MarketID: 7, YesSharesOwned: 3},
			},
		},
	}
	handler := MarketPositionsHandlerWithService(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/v0/markets/positions/7?limit=20&offset=0", nil)
	req = mux.SetURLVars(req, map[string]string{"marketId": "7"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if mockSvc.refreshCalls != 0 {
		t.Fatalf("stale positions snapshot should not refresh on read, got %d refreshes", mockSvc.refreshCalls)
	}

	var resp handlers.SuccessEnvelope[marketPositionsResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode success envelope: %v", err)
	}
	if len(resp.Result.Positions) != 1 || resp.Result.Positions[0].Username != "stale_snapshot_alice" {
		t.Fatalf("unexpected positions payload: %+v", resp.Result.Positions)
	}
	if resp.Result.Freshness == nil || !resp.Result.Freshness.IsStale {
		t.Fatalf("expected stale freshness metadata, got %+v", resp.Result.Freshness)
	}
}

func TestMarketUserPositionHandlerWithService_FailureEnvelope(t *testing.T) {
	handler := MarketUserPositionHandlerWithService(&mockUserPositionService{
		mockPositionsService: mockPositionsService{err: errors.New("boom")},
	})

	req := httptest.NewRequest(http.MethodGet, "/v0/markets/positions/77/alice", nil)
	req = mux.SetURLVars(req, map[string]string{
		"marketId": "77",
		"username": "alice",
	})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if resp.Reason != string(handlers.ReasonInternalError) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonInternalError, resp.Reason)
	}
}

type readModelPositionsService struct {
	mockPositionsService
	readModel    *dmarkets.MarketPositionsSnapshot
	refreshCalls int
}

func (m *readModelPositionsService) GetMarketPositionsReadModel(ctx context.Context, marketID int64, p dmarkets.Page) (*dmarkets.MarketPositionsSnapshot, error) {
	if m.readModel == nil {
		return nil, nil
	}
	return &dmarkets.MarketPositionsSnapshot{
		MarketID:            m.readModel.MarketID,
		Positions:           m.readModel.Positions,
		GeneratedAt:         m.readModel.GeneratedAt,
		Source:              m.readModel.Source,
		TransactionSafeRead: m.readModel.TransactionSafeRead,
		IsStale:             m.readModel.IsStale,
		StaleReason:         m.readModel.StaleReason,
		MarkedStaleAt:       m.readModel.MarkedStaleAt,
	}, nil
}

func (m *readModelPositionsService) RefreshMarketPositionsSnapshot(ctx context.Context, marketID int64) (*dmarkets.MarketPositionsSnapshot, error) {
	m.refreshCalls++
	return m.readModel, nil
}

type mockUserPositionService struct {
	mockPositionsService
	position *dmarkets.UserPosition
}

func (m *mockUserPositionService) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.position, nil
}
