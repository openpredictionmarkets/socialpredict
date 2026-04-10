package betshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dmarkets "socialpredict/internal/domain/markets"

	"github.com/gorilla/mux"
)

// marketServiceStub satisfies dmarkets.ServiceInterface for tests.
type marketServiceStub struct {
	getMarketBetsFunc func(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error)
}

func (m marketServiceStub) CreateMarket(context.Context, dmarkets.MarketCreateRequest, string) (*dmarkets.Market, error) {
	panic("not implemented")
}
func (m marketServiceStub) SetCustomLabels(context.Context, int64, string, string) error {
	panic("not implemented")
}
func (m marketServiceStub) GetMarket(context.Context, int64) (*dmarkets.Market, error) {
	panic("not implemented")
}
func (m marketServiceStub) ListMarkets(context.Context, dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	panic("not implemented")
}
func (m marketServiceStub) SearchMarkets(context.Context, string, dmarkets.SearchFilters) (*dmarkets.SearchResults, error) {
	panic("not implemented")
}
func (m marketServiceStub) ResolveMarket(context.Context, int64, string, string) error {
	panic("not implemented")
}
func (m marketServiceStub) ListByStatus(context.Context, string, dmarkets.Page) ([]*dmarkets.Market, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetMarketLeaderboard(context.Context, int64, dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
	panic("not implemented")
}
func (m marketServiceStub) ProjectProbability(context.Context, dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetMarketDetails(context.Context, int64) (*dmarkets.MarketOverview, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetMarketBets(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error) {
	if m.getMarketBetsFunc == nil {
		panic("GetMarketBets called without stub")
	}
	return m.getMarketBetsFunc(ctx, marketID)
}
func (m marketServiceStub) GetMarketPositions(context.Context, int64) (dmarkets.MarketPositions, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetUserPositionInMarket(context.Context, int64, string) (*dmarkets.UserPosition, error) {
	panic("not implemented")
}
func (m marketServiceStub) CalculateMarketVolume(context.Context, int64) (int64, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetPublicMarket(context.Context, int64) (*dmarkets.PublicMarket, error) {
	panic("not implemented")
}

func TestMarketBetsHandlerWithService(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name           string
		method         string
		vars           map[string]string
		stub           marketServiceStub
		wantStatusCode int
		wantBodySubstr string
		verifyBody     bool
	}{
		{
			name:           "non GET method rejected",
			method:         http.MethodPost,
			vars:           map[string]string{"marketId": "1"},
			stub:           marketServiceStub{},
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "missing market id",
			method:         http.MethodGet,
			wantStatusCode: http.StatusBadRequest,
			stub:           marketServiceStub{},
			wantBodySubstr: "Market ID is required",
		},
		{
			name:           "invalid market id value",
			method:         http.MethodGet,
			vars:           map[string]string{"marketId": "abc"},
			stub:           marketServiceStub{},
			wantStatusCode: http.StatusBadRequest,
			wantBodySubstr: "Invalid market ID",
		},
		{
			name:   "market not found",
			method: http.MethodGet,
			vars:   map[string]string{"marketId": "42"},
			stub: marketServiceStub{
				getMarketBetsFunc: func(context.Context, int64) ([]*dmarkets.BetDisplayInfo, error) {
					return nil, dmarkets.ErrMarketNotFound
				},
			},
			wantStatusCode: http.StatusNotFound,
			wantBodySubstr: "Market not found",
		},
		{
			name:   "invalid input from service",
			method: http.MethodGet,
			vars:   map[string]string{"marketId": "42"},
			stub: marketServiceStub{
				getMarketBetsFunc: func(context.Context, int64) ([]*dmarkets.BetDisplayInfo, error) {
					return nil, dmarkets.ErrInvalidInput
				},
			},
			wantStatusCode: http.StatusBadRequest,
			wantBodySubstr: "Invalid market ID",
		},
		{
			name:   "internal error bubbled up",
			method: http.MethodGet,
			vars:   map[string]string{"marketId": "42"},
			stub: marketServiceStub{
				getMarketBetsFunc: func(context.Context, int64) ([]*dmarkets.BetDisplayInfo, error) {
					return nil, errors.New("boom")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBodySubstr: "Internal server error",
		},
		{
			name:   "successful response",
			method: http.MethodGet,
			vars:   map[string]string{"marketId": "7"},
			stub: marketServiceStub{
				getMarketBetsFunc: func(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error) {
					if marketID != 7 {
						t.Fatalf("expected marketID 7, got %d", marketID)
					}
					return []*dmarkets.BetDisplayInfo{
						{
							Username:    "alice",
							Outcome:     "YES",
							Amount:      100,
							Probability: 0.55,
							PlacedAt:    now,
						},
					}, nil
				},
			},
			wantStatusCode: http.StatusOK,
			verifyBody:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := MarketBetsHandlerWithService(tt.stub)
			req := httptest.NewRequest(tt.method, "/v0/markets/marketId/bets", nil)
			if tt.vars != nil {
				req = mux.SetURLVars(req, tt.vars)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Fatalf("expected status %d, got %d (body: %s)", tt.wantStatusCode, rr.Code, rr.Body.String())
			}

			if tt.verifyBody {
				var decoded []dmarkets.BetDisplayInfo
				if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if len(decoded) != 1 {
					t.Fatalf("expected 1 bet, got %d", len(decoded))
				}
				got := decoded[0]
				if got.Username != "alice" || got.Outcome != "YES" || got.Amount != 100 || got.Probability != 0.55 || !got.PlacedAt.Equal(now) {
					t.Fatalf("unexpected bet payload: %+v", got)
				}
			} else if tt.wantBodySubstr != "" && !strings.Contains(rr.Body.String(), tt.wantBodySubstr) {
				t.Fatalf("expected body to contain %q, got %q", tt.wantBodySubstr, rr.Body.String())
			}
		})
	}
}
