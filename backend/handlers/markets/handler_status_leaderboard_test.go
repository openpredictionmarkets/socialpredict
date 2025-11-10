package marketshandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"

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

	handler := NewHandler(svc, nil)
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

	handler := NewHandler(svc, nil)
	req := httptest.NewRequest(http.MethodGet, "/v0/markets/77/leaderboard?limit=25", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/v0/markets/{id}/leaderboard", handler.MarketLeaderboard)

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp dto.LeaderboardResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("expected total 1, got %d", resp.Total)
	}
	if len(resp.Leaderboard) != 1 || resp.Leaderboard[0].Username != "alice" {
		t.Fatalf("unexpected leaderboard payload: %+v", resp.Leaderboard)
	}
}
