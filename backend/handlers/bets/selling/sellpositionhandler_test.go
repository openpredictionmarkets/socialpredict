package sellbetshandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"socialpredict/handlers/bets/dto"
	bets "socialpredict/internal/domain/bets"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/models/modelstesting"
)

type fakeSellService struct {
	req  bets.SellRequest
	resp *bets.SellResult
	err  error
}

func (f *fakeSellService) Place(ctx context.Context, req bets.PlaceRequest) (*bets.PlacedBet, error) {
	return nil, nil
}
func (f *fakeSellService) Sell(ctx context.Context, req bets.SellRequest) (*bets.SellResult, error) {
	f.req = req
	return f.resp, f.err
}

type fakeAuth struct{ user *dusers.User }

func (f *fakeAuth) CurrentUser(r *http.Request) (*dusers.User, *authsvc.HTTPError) {
	return f.user, nil
}
func (f *fakeAuth) RequireUser(r *http.Request) (*dusers.User, *authsvc.HTTPError) {
	return f.user, nil
}
func (f *fakeAuth) RequireAdmin(r *http.Request) (*dusers.User, *authsvc.HTTPError) {
	return f.user, nil
}
func (f *fakeAuth) ChangePassword(ctx context.Context, username, currentPassword, newPassword string) error {
	return nil
}
func (f *fakeAuth) MustChangePassword(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func TestSellPositionHandler_Success(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	svc := &fakeSellService{resp: &bets.SellResult{
		Username:      "alice",
		MarketID:      7,
		SharesSold:    3,
		SaleValue:     60,
		Dust:          5,
		Outcome:       "YES",
		TransactionAt: time.Now(),
	}}
	users := &fakeAuth{user: &dusers.User{Username: "alice"}}

	body, _ := json.Marshal(dto.SellBetRequest{MarketID: 7, Amount: 65, Outcome: "YES"})
	req := httptest.NewRequest(http.MethodPost, "/v0/sell", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := SellPositionHandler(svc, users)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	if svc.req.Username != "alice" || svc.req.MarketID != 7 || svc.req.Amount != 65 {
		t.Fatalf("unexpected request payload: %+v", svc.req)
	}

	var resp dto.SellBetResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.SharesSold != 3 || resp.SaleValue != 60 || resp.Dust != 5 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestSellPositionHandler_ErrorMapping(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	users := &fakeAuth{user: &dusers.User{Username: "alice"}}

	cases := []struct {
		name string
		err  error
		want int
	}{
		{"bad outcome", bets.ErrInvalidOutcome, http.StatusBadRequest},
		{"market closed", bets.ErrMarketClosed, http.StatusConflict},
		{"no position", bets.ErrNoPosition, http.StatusUnprocessableEntity},
		{"dust cap", bets.ErrDustCapExceeded{Cap: 2, Requested: 3}, http.StatusUnprocessableEntity},
		{"market not found", dmarkets.ErrMarketNotFound, http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeSellService{err: tc.err}
			body, _ := json.Marshal(dto.SellBetRequest{MarketID: 1, Amount: 10, Outcome: "YES"})
			req := httptest.NewRequest(http.MethodPost, "/v0/sell", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := SellPositionHandler(svc, users)
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.want {
				t.Fatalf("expected status %d, got %d", tc.want, rr.Code)
			}
		})
	}
}

func TestSellPositionHandler_InvalidJSON(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	svc := &fakeSellService{}
	users := &fakeAuth{user: &dusers.User{Username: "alice"}}

	req := httptest.NewRequest(http.MethodPost, "/v0/sell", bytes.NewBufferString("{"))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := SellPositionHandler(svc, users)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
