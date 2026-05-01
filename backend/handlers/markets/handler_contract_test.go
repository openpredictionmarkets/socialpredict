package marketshandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"socialpredict/handlers"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"

	"github.com/gorilla/mux"
)

type contractServiceMock struct {
	createFn       func(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error)
	listFn         func(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error)
	detailsFn      func(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error)
	searchFn       func(ctx context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.SearchResults, error)
	resolveFn      func(ctx context.Context, marketID int64, resolution string, username string) error
	listByStatusFn func(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error)
	leaderboardFn  func(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error)
	projectFn      func(ctx context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error)
}

func (m *contractServiceMock) CreateMarket(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req, creatorUsername)
	}
	return nil, nil
}

func (m *contractServiceMock) SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error {
	return nil
}

func (m *contractServiceMock) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *contractServiceMock) ListMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filters)
	}
	return nil, nil
}

func (m *contractServiceMock) GetMarketDetails(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error) {
	if m.detailsFn != nil {
		return m.detailsFn(ctx, marketID)
	}
	return nil, nil
}

func (m *contractServiceMock) SearchMarkets(ctx context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.SearchResults, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, query, filters)
	}
	return nil, nil
}

func (m *contractServiceMock) ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error {
	if m.resolveFn != nil {
		return m.resolveFn(ctx, marketID, resolution, username)
	}
	return nil
}

func (m *contractServiceMock) ListByStatus(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error) {
	if m.listByStatusFn != nil {
		return m.listByStatusFn(ctx, status, p)
	}
	return nil, nil
}

func (m *contractServiceMock) GetMarketLeaderboard(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
	if m.leaderboardFn != nil {
		return m.leaderboardFn(ctx, marketID, p)
	}
	return nil, nil
}

func (m *contractServiceMock) ProjectProbability(ctx context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	if m.projectFn != nil {
		return m.projectFn(ctx, req)
	}
	return nil, nil
}

type contractAuthMock struct {
	user *dusers.User
	err  *authsvc.AuthError
}

func (m *contractAuthMock) CurrentUser(r *http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.user, m.err
}

func (m *contractAuthMock) RequireUser(r *http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.user, m.err
}

func (m *contractAuthMock) RequireAdmin(r *http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.user, m.err
}

func TestHandlerCreateMarket_SuccessAndBusinessFailure(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	auth := &contractAuthMock{user: &dusers.User{Username: "alice"}}

	t.Run("success returns market response", func(t *testing.T) {
		service := &contractServiceMock{
			createFn: func(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error) {
				if creatorUsername != "alice" {
					t.Fatalf("expected creator alice, got %q", creatorUsername)
				}
				return &dmarkets.Market{
					ID:                 7,
					QuestionTitle:      req.QuestionTitle,
					Description:        req.Description,
					OutcomeType:        req.OutcomeType,
					ResolutionDateTime: req.ResolutionDateTime,
					CreatorUsername:    creatorUsername,
					YesLabel:           "YES",
					NoLabel:            "NO",
					Status:             "active",
					CreatedAt:          now,
					UpdatedAt:          now,
				}, nil
			},
		}

		body := bytes.NewBufferString(`{"questionTitle":"Will BTC rise?","description":"Market","outcomeType":"BINARY","resolutionDateTime":"2030-01-01T00:00:00Z"}`)
		req := httptest.NewRequest(http.MethodPost, "/v0/markets", body)
		rr := httptest.NewRecorder()

		NewHandler(service, auth).CreateMarket(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", rr.Code)
		}

		var resp dto.MarketResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode market response: %v", err)
		}
		if resp.ID != 7 || resp.CreatorUsername != "alice" || resp.QuestionTitle != "Will BTC rise?" {
			t.Fatalf("unexpected create response: %+v", resp)
		}
	})

	t.Run("insufficient balance uses stable reason", func(t *testing.T) {
		service := &contractServiceMock{
			createFn: func(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error) {
				return nil, dmarkets.ErrInsufficientBalance
			},
		}

		body := bytes.NewBufferString(`{"questionTitle":"Will BTC rise?","description":"Market","outcomeType":"BINARY","resolutionDateTime":"2030-01-01T00:00:00Z"}`)
		req := httptest.NewRequest(http.MethodPost, "/v0/markets", body)
		rr := httptest.NewRecorder()

		NewHandler(service, auth).CreateMarket(rr, req)

		assertFailureEnvelope(t, rr, http.StatusUnprocessableEntity, handlers.ReasonInsufficientBalance)
	})
}

func TestHandlerListMarkets_UsesCreatedByFilterWhenStatusProvided(t *testing.T) {
	now := time.Now().UTC()
	market := &dmarkets.Market{
		ID:                 11,
		QuestionTitle:      "Closed market",
		Description:        "desc",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(-time.Hour),
		CreatorUsername:    "alice",
		YesLabel:           "YES",
		NoLabel:            "NO",
		Status:             "closed",
		CreatedAt:          now.Add(-24 * time.Hour),
		UpdatedAt:          now,
	}

	service := &contractServiceMock{
		listFn: func(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
			want := dmarkets.ListFilters{Status: "closed", CreatedBy: "alice", Limit: 5, Offset: 2}
			if filters != want {
				t.Fatalf("expected filters %+v, got %+v", want, filters)
			}
			return []*dmarkets.Market{market}, nil
		},
		listByStatusFn: func(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error) {
			t.Fatalf("expected status+created_by request to use ListMarkets, got ListByStatus(%q, %+v)", status, p)
			return nil, nil
		},
		detailsFn: func(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error) {
			return newOverview(market), nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v0/markets?status=closed&created_by=alice&limit=5&offset=2", nil)
	rr := httptest.NewRecorder()

	NewHandler(service, nil).ListMarkets(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp dto.ListMarketsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if resp.Total != 1 || len(resp.Markets) != 1 || resp.Markets[0].Market.ID != 11 {
		t.Fatalf("unexpected list response: %+v", resp)
	}
}

func TestHandlerSearchMarkets_SupportsLegacyQAndInvalidRequestFailure(t *testing.T) {
	now := time.Now().UTC()
	market := &dmarkets.Market{
		ID:                 3,
		QuestionTitle:      "Bitcoin market",
		Description:        "desc",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(24 * time.Hour),
		CreatorUsername:    "alice",
		YesLabel:           "YES",
		NoLabel:            "NO",
		Status:             "active",
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	service := &contractServiceMock{
		searchFn: func(ctx context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.SearchResults, error) {
			if query != "bitcoin" {
				t.Fatalf("expected query bitcoin, got %q", query)
			}
			if filters.Status != "active" || filters.Limit != 4 || filters.Offset != 1 {
				t.Fatalf("unexpected search filters: %+v", filters)
			}
			return &dmarkets.SearchResults{
				PrimaryResults: []*dmarkets.Market{market},
				Query:          query,
				PrimaryStatus:  filters.Status,
				PrimaryCount:   1,
				TotalCount:     1,
			}, nil
		},
		detailsFn: func(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error) {
			return newOverview(market), nil
		},
	}

	t.Run("success with legacy q parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/markets/search?q=bitcoin&status=active&limit=4&offset=1", nil)
		rr := httptest.NewRecorder()

		NewHandler(service, nil).SearchMarkets(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		var resp dto.SearchResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode search response: %v", err)
		}
		if resp.TotalCount != 1 || len(resp.PrimaryResults) != 1 || resp.PrimaryResults[0].Market.ID != 3 {
			t.Fatalf("unexpected search response: %+v", resp)
		}
	})

	t.Run("missing query returns failure envelope", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/markets/search", nil)
		rr := httptest.NewRecorder()

		NewHandler(service, nil).SearchMarkets(rr, req)

		assertFailureEnvelope(t, rr, http.StatusBadRequest, handlers.ReasonInvalidRequest)
	})

	t.Run("suspicious query returns failure envelope", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/markets/search?query=%3Cscript%3Ealert(1)%3C%2Fscript%3E", nil)
		rr := httptest.NewRecorder()

		NewHandler(service, nil).SearchMarkets(rr, req)

		assertFailureEnvelope(t, rr, http.StatusBadRequest, handlers.ReasonInvalidRequest)
	})

	t.Run("overlong query returns failure envelope", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/markets/search?query="+strings.Repeat("a", 101), nil)
		rr := httptest.NewRecorder()

		NewHandler(service, nil).SearchMarkets(rr, req)

		assertFailureEnvelope(t, rr, http.StatusBadRequest, handlers.ReasonInvalidRequest)
	})
}

func TestHandlerGetDetails_NotFoundUsesFailureEnvelope(t *testing.T) {
	service := &contractServiceMock{
		detailsFn: func(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error) {
			return nil, dmarkets.ErrMarketNotFound
		},
	}

	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/v0/markets/99", nil), map[string]string{"id": "99"})
	rr := httptest.NewRecorder()

	NewHandler(service, nil).GetDetails(rr, req)

	assertFailureEnvelope(t, rr, http.StatusNotFound, handlers.ReasonMarketNotFound)
}

func TestHandlerResolveMarket_MapsAuthAndStateFailures(t *testing.T) {
	t.Run("password change required uses auth reason", func(t *testing.T) {
		service := &contractServiceMock{}
		auth := &contractAuthMock{
			err: &authsvc.AuthError{Kind: authsvc.ErrorKindPasswordChangeRequired, Message: "Password change required"},
		}

		req := mux.SetURLVars(httptest.NewRequest(http.MethodPost, "/v0/markets/5/resolve", bytes.NewBufferString(`{"resolution":"yes"}`)), map[string]string{"id": "5"})
		rr := httptest.NewRecorder()

		NewHandler(service, auth).ResolveMarket(rr, req)

		assertFailureEnvelope(t, rr, http.StatusForbidden, handlers.ReasonPasswordChangeRequired)
	})

	t.Run("invalid state uses market closed", func(t *testing.T) {
		service := &contractServiceMock{
			resolveFn: func(ctx context.Context, marketID int64, resolution string, username string) error {
				if marketID != 5 || resolution != "yes" || username != "alice" {
					t.Fatalf("unexpected resolve args: marketID=%d resolution=%q username=%q", marketID, resolution, username)
				}
				return dmarkets.ErrInvalidState
			},
		}
		auth := &contractAuthMock{user: &dusers.User{Username: "alice"}}

		req := mux.SetURLVars(httptest.NewRequest(http.MethodPost, "/v0/markets/5/resolve", bytes.NewBufferString(`{"resolution":"yes"}`)), map[string]string{"id": "5"})
		rr := httptest.NewRecorder()

		NewHandler(service, auth).ResolveMarket(rr, req)

		assertFailureEnvelope(t, rr, http.StatusConflict, handlers.ReasonMarketClosed)
	})
}

func TestHandlerProjectProbability_UsesQueryParams(t *testing.T) {
	service := &contractServiceMock{
		projectFn: func(ctx context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
			if req != (dmarkets.ProbabilityProjectionRequest{MarketID: 42, Amount: 15, Outcome: "YES"}) {
				t.Fatalf("unexpected projection request: %+v", req)
			}
			return &dmarkets.ProbabilityProjection{
				CurrentProbability:   0.4,
				ProjectedProbability: 0.55,
			}, nil
		},
	}

	req := mux.SetURLVars(
		httptest.NewRequest(http.MethodGet, "/v0/markets/42/projection?amount=15&outcome=YES", nil),
		map[string]string{"id": "42"},
	)
	rr := httptest.NewRecorder()

	NewHandler(service, nil).ProjectProbability(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp dto.ProbabilityProjectionResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode projection response: %v", err)
	}
	if resp.MarketID != 42 || resp.Amount != 15 || resp.Outcome != "YES" {
		t.Fatalf("unexpected projection response: %+v", resp)
	}
}

func assertFailureEnvelope(t *testing.T, rr *httptest.ResponseRecorder, wantStatus int, wantReason handlers.FailureReason) {
	t.Helper()

	if rr.Code != wantStatus {
		t.Fatalf("expected status %d, got %d body=%s", wantStatus, rr.Code, rr.Body.String())
	}

	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if resp.OK {
		t.Fatalf("expected ok=false, got %+v", resp)
	}
	if resp.Reason != string(wantReason) {
		t.Fatalf("expected reason %q, got %q", wantReason, resp.Reason)
	}
}

func newOverview(market *dmarkets.Market) *dmarkets.MarketOverview {
	return &dmarkets.MarketOverview{
		Market:  market,
		Creator: &dmarkets.CreatorSummary{Username: market.CreatorUsername},
	}
}
