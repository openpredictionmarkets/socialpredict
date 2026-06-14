package marketshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"socialpredict/handlers"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/security"

	"github.com/gorilla/mux"
)

func TestListByStatusHandler_Smoke(t *testing.T) {
	svc := &MockService{}
	svc.ListByStatusFn = func(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error) {
		if status != "active" {
			t.Fatalf("expected status active, got %s", status)
		}
		if p.Limit != 50 {
			t.Fatalf("expected limit 50, got %d", p.Limit)
		}
		now := time.Now()
		return []*dmarkets.Market{{
			ID:                 101,
			QuestionTitle:      "Sample",
			Description:        "desc",
			OutcomeType:        "BINARY",
			ResolutionDateTime: now.Add(24 * time.Hour),
			CreatorUsername:    "creator",
			YesLabel:           "YES",
			NoLabel:            "NO",
			Status:             "active",
			CreatedAt:          now,
			UpdatedAt:          now,
		}}, nil
	}

	handler := NewHandler(svc, nil, security.NewSecurityService())
	req := httptest.NewRequest(http.MethodGet, "/v0/markets/status/active?limit=50", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/v0/markets/status/{status}", handler.ListByStatus)

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp struct {
		Markets []json.RawMessage `json:"markets"`
		Total   int               `json:"total"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("expected total 1, got %d", resp.Total)
	}
	if len(resp.Markets) != 1 {
		t.Fatalf("expected 1 market, got %d", len(resp.Markets))
	}
}

func TestMarketLeaderboardHandler_Smoke(t *testing.T) {
	svc := &MockService{}
	svc.MarketLeaderboardFn = func(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
		if marketID != 77 {
			t.Fatalf("expected marketID 77, got %d", marketID)
		}
		if p.Limit != 25 {
			t.Fatalf("expected limit 25, got %d", p.Limit)
		}
		return []*dmarkets.LeaderboardRow{{
			Username:       "alice",
			Profit:         12,
			CurrentValue:   100,
			TotalSpent:     88,
			Position:       "YES",
			YesSharesOwned: 5,
			NoSharesOwned:  0,
			Rank:           1,
		}}, nil
	}

	handler := NewHandler(svc, nil, security.NewSecurityService())
	req := httptest.NewRequest(http.MethodGet, "/v0/markets/77/leaderboard?limit=25", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/v0/markets/{id}/leaderboard", handler.MarketLeaderboard)

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp handlers.SuccessEnvelope[dto.LeaderboardResponse]
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got false")
	}
	if resp.Result.Total != 1 {
		t.Fatalf("expected total 1, got %d", resp.Result.Total)
	}
	if len(resp.Result.Leaderboard) != 1 || resp.Result.Leaderboard[0].Username != "alice" {
		t.Fatalf("unexpected leaderboard payload: %+v", resp.Result.Leaderboard)
	}
}

func TestMarketLeaderboardHandler_UsesSnapshotFreshnessWhenAvailable(t *testing.T) {
	generatedAt := time.Now().UTC().Add(-2 * time.Minute)
	svc := &readModelLeaderboardServiceMock{
		MockService: MockService{
			MarketLeaderboardFn: func(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
				t.Fatalf("raw leaderboard calculator should not be used when snapshot is available")
				return nil, nil
			},
		},
		ReadModelFn: func(ctx context.Context, marketID int64, p dmarkets.Page) (*dmarkets.MarketLeaderboardSnapshot, error) {
			if marketID != 77 {
				t.Fatalf("expected marketID 77, got %d", marketID)
			}
			if p.Limit != 25 {
				t.Fatalf("expected limit 25, got %d", p.Limit)
			}
			return &dmarkets.MarketLeaderboardSnapshot{
				MarketID:    marketID,
				GeneratedAt: generatedAt,
				Source:      "read_model",
				Rows: []*dmarkets.LeaderboardRow{{
					Username:     "snapshot_alice",
					Profit:       8,
					CurrentValue: 42,
					TotalSpent:   34,
					Position:     "YES",
					Rank:         1,
				}},
			}, nil
		},
	}

	handler := NewHandler(svc, nil, security.NewSecurityService())
	req := httptest.NewRequest(http.MethodGet, "/v0/markets/77/leaderboard?limit=25", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/v0/markets/{id}/leaderboard", handler.MarketLeaderboard)

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp handlers.SuccessEnvelope[dto.LeaderboardResponse]
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(resp.Result.Leaderboard) != 1 || resp.Result.Leaderboard[0].Username != "snapshot_alice" {
		t.Fatalf("unexpected leaderboard payload: %+v", resp.Result.Leaderboard)
	}
	if resp.Result.Freshness == nil {
		t.Fatalf("expected snapshot freshness metadata")
	}
	if !resp.Result.Freshness.GeneratedAt.Equal(generatedAt) {
		t.Fatalf("freshness generatedAt = %s, want %s", resp.Result.Freshness.GeneratedAt, generatedAt)
	}
	if resp.Result.Freshness.TargetFreshnessSeconds != int(dmarkets.MarketLeaderboardSnapshotTargetFreshness.Seconds()) {
		t.Fatalf("freshness target = %d, want %d", resp.Result.Freshness.TargetFreshnessSeconds, int(dmarkets.MarketLeaderboardSnapshotTargetFreshness.Seconds()))
	}
	if resp.Result.Freshness.TransactionSafeRead {
		t.Fatalf("leaderboard snapshot must not be marked transaction safe")
	}
}

func TestMarketLeaderboardHandler_ServesStaleSnapshotWithoutRefresh(t *testing.T) {
	generatedAt := time.Now().UTC().Add(-2 * dmarkets.MarketLeaderboardSnapshotTargetFreshness)
	svc := &readModelLeaderboardServiceMock{
		MockService: MockService{
			MarketLeaderboardFn: func(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
				t.Fatalf("raw leaderboard calculator should not be used when stale snapshot is available")
				return nil, nil
			},
		},
		ReadModelFn: func(ctx context.Context, marketID int64, p dmarkets.Page) (*dmarkets.MarketLeaderboardSnapshot, error) {
			return &dmarkets.MarketLeaderboardSnapshot{
				MarketID:    marketID,
				GeneratedAt: generatedAt,
				Source:      "read_model",
				IsStale:     true,
				StaleReason: "bet_accepted",
				Rows: []*dmarkets.LeaderboardRow{{
					Username: "stale_snapshot_alice",
					Rank:     1,
				}},
			}, nil
		},
		RefreshFn: func(ctx context.Context, marketID int64) (*dmarkets.MarketLeaderboardSnapshot, error) {
			t.Fatalf("stale leaderboard snapshot should not refresh on read")
			return nil, nil
		},
	}

	handler := NewHandler(svc, nil, security.NewSecurityService())
	req := httptest.NewRequest(http.MethodGet, "/v0/markets/77/leaderboard?limit=25", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/v0/markets/{id}/leaderboard", handler.MarketLeaderboard)

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp handlers.SuccessEnvelope[dto.LeaderboardResponse]
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(resp.Result.Leaderboard) != 1 || resp.Result.Leaderboard[0].Username != "stale_snapshot_alice" {
		t.Fatalf("unexpected leaderboard payload: %+v", resp.Result.Leaderboard)
	}
	if resp.Result.Freshness == nil || !resp.Result.Freshness.IsStale {
		t.Fatalf("expected stale freshness metadata, got %+v", resp.Result.Freshness)
	}
}

func TestMarketLeaderboardHandler_FailureEnvelope(t *testing.T) {
	svc := &MockService{}
	svc.MarketLeaderboardFn = func(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
		return nil, errors.New("boom")
	}

	handler := NewHandler(svc, nil, security.NewSecurityService())
	req := httptest.NewRequest(http.MethodGet, "/v0/markets/77/leaderboard", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/v0/markets/{id}/leaderboard", handler.MarketLeaderboard)

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal failure: %v", err)
	}
	if resp.Reason != string(handlers.ReasonInternalError) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonInternalError, resp.Reason)
	}
}

type readModelLeaderboardServiceMock struct {
	MockService
	ReadModelFn func(ctx context.Context, marketID int64, p dmarkets.Page) (*dmarkets.MarketLeaderboardSnapshot, error)
	RefreshFn   func(ctx context.Context, marketID int64) (*dmarkets.MarketLeaderboardSnapshot, error)
}

func (m *readModelLeaderboardServiceMock) GetMarketLeaderboardReadModel(ctx context.Context, marketID int64, p dmarkets.Page) (*dmarkets.MarketLeaderboardSnapshot, error) {
	if m.ReadModelFn != nil {
		return m.ReadModelFn(ctx, marketID, p)
	}
	return nil, nil
}

func (m *readModelLeaderboardServiceMock) RefreshMarketLeaderboardSnapshot(ctx context.Context, marketID int64) (*dmarkets.MarketLeaderboardSnapshot, error) {
	if m.RefreshFn != nil {
		return m.RefreshFn(ctx, marketID)
	}
	return nil, nil
}

func TestListUserOwnedMarketsHandlerRequiresLogin(t *testing.T) {
	handler := ListUserOwnedMarketsHandler(&MockService{}, &contractAuthMock{
		err: &authsvc.AuthError{Kind: authsvc.ErrorKindInvalidToken, Message: "invalid token"},
	})
	req := httptest.NewRequest(http.MethodGet, "/v0/users/alice/owned-markets", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestListUserOwnedMarketsHandlerUsesOwnedByFilter(t *testing.T) {
	now := time.Now()
	svc := &MockService{}
	svc.ListLifecycleFn = func(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
		if filters.OwnedBy != "alice" {
			t.Fatalf("OwnedBy = %q, want alice", filters.OwnedBy)
		}
		if filters.Status != dmarkets.MarketStatusAll {
			t.Fatalf("Status = %q, want all", filters.Status)
		}
		if filters.Limit != 20 || filters.Offset != 5 {
			t.Fatalf("pagination = %d/%d, want 20/5", filters.Limit, filters.Offset)
		}
		return []*dmarkets.Market{{
			ID:                 45,
			QuestionTitle:      "Owned Market",
			Description:        "desc",
			OutcomeType:        "BINARY",
			ResolutionDateTime: now.Add(24 * time.Hour),
			CreatorUsername:    "bob",
			StewardUsername:    "alice",
			YesLabel:           "YES",
			NoLabel:            "NO",
			Status:             dmarkets.MarketStatusActive,
			LifecycleStatus:    dmarkets.MarketLifecyclePublished,
			CreatedAt:          now,
			UpdatedAt:          now,
		}}, nil
	}

	handler := ListUserOwnedMarketsHandler(svc, &contractAuthMock{user: &dusers.User{Username: "viewer"}})
	req := httptest.NewRequest(http.MethodGet, "/v0/users/alice/owned-markets?limit=20&offset=5", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp handlers.SuccessEnvelope[userOwnedMarketsResponse]
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.OK || resp.Result.Total != 1 || resp.Result.Markets[0].StewardUsername != "alice" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestListUserOwnedMarketsHandlerAttachesMarketGroupMetadata(t *testing.T) {
	now := time.Now().UTC()
	group := &dmarkets.MarketGroup{
		ID:                 88,
		QuestionTitle:      "Favorite Tree",
		Description:        "Grouped favorite tree market",
		GroupType:          dmarkets.MarketGroupTypeMultipleChoiceBinary,
		LifecycleStatus:    dmarkets.MarketLifecyclePublished,
		ProposalCost:       10,
		CreatorUsername:    "testuser01",
		StewardUsername:    "testuser01",
		ApprovedBy:         "auto-approval",
		ApprovedAt:         &now,
		ResolutionDateTime: now.Add(24 * time.Hour),
		CreatedAt:          now,
		UpdatedAt:          now,
		Members: []dmarkets.MarketGroupMember{
			{MarketID: 16, AnswerLabel: "Birch", DisplayOrder: 0},
			{MarketID: 17, AnswerLabel: "Pine", DisplayOrder: 1},
			{MarketID: 18, AnswerLabel: "Oak", DisplayOrder: 2},
		},
	}
	svc := &MockService{}
	svc.ListLifecycleFn = func(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
		if filters.OwnedBy != "testuser01" {
			t.Fatalf("OwnedBy = %q, want testuser01", filters.OwnedBy)
		}
		return []*dmarkets.Market{
			{
				ID:                 16,
				QuestionTitle:      "Favorite Tree - Birch",
				Description:        "Answer choice child market",
				OutcomeType:        "BINARY",
				ResolutionDateTime: now.Add(24 * time.Hour),
				CreatorUsername:    "testuser01",
				StewardUsername:    "testuser01",
				YesLabel:           "YES",
				NoLabel:            "NO",
				Status:             dmarkets.MarketStatusActive,
				LifecycleStatus:    dmarkets.MarketLifecyclePublished,
				CreatedAt:          now,
				UpdatedAt:          now,
			},
		}, nil
	}
	svc.MarketGroupLookupFn = func(ctx context.Context, marketID int64) (*dmarkets.MarketGroup, error) {
		if marketID != 16 {
			t.Fatalf("marketID = %d, want 16", marketID)
		}
		return group, nil
	}

	handler := ListUserOwnedMarketsHandler(svc, &contractAuthMock{user: &dusers.User{Username: "viewer"}})
	req := httptest.NewRequest(http.MethodGet, "/v0/users/testuser01/owned-markets", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "testuser01"})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp handlers.SuccessEnvelope[userOwnedMarketsResponse]
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.OK || resp.Result.Total != 1 || len(resp.Result.Markets) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	market := resp.Result.Markets[0]
	if market.MarketGroup == nil {
		t.Fatalf("expected owned market response to include grouped-market metadata")
	}
	if market.MarketGroup.ID != group.ID ||
		market.MarketGroup.QuestionTitle != "Favorite Tree" ||
		market.MarketGroup.AnswerLabel != "Birch" ||
		market.MarketGroup.AnswerCount != 3 ||
		market.MarketGroup.DisplayOrder != 0 {
		t.Fatalf("unexpected market group link: %+v", market.MarketGroup)
	}
}
