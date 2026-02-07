package buybetshandlers

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

type fakeBetsService struct {
	req  bets.PlaceRequest
	resp *bets.PlacedBet
	err  error
}

func (f *fakeBetsService) Place(ctx context.Context, req bets.PlaceRequest) (*bets.PlacedBet, error) {
	f.req = req
	if f.err != nil {
		return nil, f.err
	}
	return f.resp, nil
}

func (f *fakeBetsService) Sell(ctx context.Context, req bets.SellRequest) (*bets.SellResult, error) {
	return nil, nil
}

type fakeAuth struct {
	user *dusers.User
}

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

func TestPlaceBetHandler_Success(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	betsSvc := &fakeBetsService{resp: &bets.PlacedBet{Username: "alice", MarketID: 5, Amount: 120, Outcome: "YES", PlacedAt: time.Now()}}
	userSvc := &fakeAuth{user: &dusers.User{Username: "alice"}}

	payload := dto.PlaceBetRequest{MarketID: 5, Amount: 120, Outcome: "YES"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v0/bet", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := PlaceBetHandler(betsSvc, userSvc)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	if betsSvc.req.Username != "alice" || betsSvc.req.MarketID != 5 {
		t.Fatalf("unexpected service request: %+v", betsSvc.req)
	}

	var resp dto.PlaceBetResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Username != "alice" || resp.Amount != 120 || resp.MarketID != 5 {
		t.Fatalf("unexpected response body: %+v", resp)
	}
}

func TestPlaceBetHandler_ErrorMapping(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	userSvc := &fakeAuth{user: &dusers.User{Username: "alice"}}

	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"invalid outcome", bets.ErrInvalidOutcome, http.StatusBadRequest},
		{"insufficient", bets.ErrInsufficientBalance, http.StatusUnprocessableEntity},
		{"market closed", bets.ErrMarketClosed, http.StatusConflict},
		{"not found", dmarkets.ErrMarketNotFound, http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			betsSvc := &fakeBetsService{err: tc.err}
			payload := dto.PlaceBetRequest{MarketID: 1, Amount: 10, Outcome: "YES"}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPost, "/v0/bet", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := PlaceBetHandler(betsSvc, userSvc)
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestPlaceBetHandler_InvalidJSON(t *testing.T) {
	betsSvc := &fakeBetsService{}
	userSvc := &fakeAuth{user: &dusers.User{Username: "alice"}}
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	req := httptest.NewRequest(http.MethodPost, "/v0/bet", bytes.NewBufferString("{invalid"))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := PlaceBetHandler(betsSvc, userSvc)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
