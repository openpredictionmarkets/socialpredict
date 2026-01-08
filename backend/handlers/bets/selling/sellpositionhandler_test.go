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
	usermodels "socialpredict/internal/domain/users/models"
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

type fakeUsersService struct{ user *dusers.User }

func (f *fakeUsersService) GetUser(ctx context.Context, username string) (*dusers.User, error) {
	return f.user, nil
}
func (f *fakeUsersService) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error {
	return nil
}
func (f *fakeUsersService) GetPublicUser(ctx context.Context, username string) (*dusers.PublicUser, error) {
	return nil, nil
}
func (f *fakeUsersService) GetPrivateProfile(ctx context.Context, username string) (*dusers.PrivateProfile, error) {
	return nil, nil
}
func (f *fakeUsersService) GetUserCredit(ctx context.Context, username string, maximumDebtAllowed int64) (int64, error) {
	return 0, nil
}
func (f *fakeUsersService) GetUserPortfolio(ctx context.Context, username string) (*dusers.Portfolio, error) {
	return nil, nil
}
func (f *fakeUsersService) GetUserFinancials(ctx context.Context, username string) (map[string]int64, error) {
	return nil, nil
}
func (f *fakeUsersService) ListUserMarkets(ctx context.Context, userID int64) ([]*dusers.UserMarket, error) {
	return nil, nil
}
func (f *fakeUsersService) UpdateDescription(ctx context.Context, username, description string) (*dusers.User, error) {
	return nil, nil
}
func (f *fakeUsersService) UpdateDisplayName(ctx context.Context, username, displayName string) (*dusers.User, error) {
	return nil, nil
}
func (f *fakeUsersService) UpdateEmoji(ctx context.Context, username, emoji string) (*dusers.User, error) {
	return nil, nil
}
func (f *fakeUsersService) UpdatePersonalLinks(ctx context.Context, username string, links dusers.PersonalLinks) (*dusers.User, error) {
	return nil, nil
}
func (f *fakeUsersService) ChangePassword(ctx context.Context, username, currentPassword, newPassword string) error {
	return nil
}
func (f *fakeUsersService) ValidateUserExists(ctx context.Context, username string) error { return nil }
func (f *fakeUsersService) ValidateUserBalance(ctx context.Context, username string, requiredAmount int64, maxDebt int64) error {
	return nil
}
func (f *fakeUsersService) DeductBalance(ctx context.Context, username string, amount int64) error {
	return nil
}
func (f *fakeUsersService) CreateUser(ctx context.Context, req dusers.UserCreateRequest) (*dusers.User, error) {
	return nil, nil
}
func (f *fakeUsersService) UpdateUser(ctx context.Context, username string, req dusers.UserUpdateRequest) (*dusers.User, error) {
	return nil, nil
}
func (f *fakeUsersService) DeleteUser(ctx context.Context, username string) error { return nil }
func (f *fakeUsersService) List(ctx context.Context, filters usermodels.ListFilters) ([]*dusers.User, error) {
	return nil, nil
}
func (f *fakeUsersService) ListUserBets(ctx context.Context, username string) ([]*dusers.UserBet, error) {
	return nil, nil
}
func (f *fakeUsersService) GetMarketQuestion(ctx context.Context, marketID uint) (string, error) {
	return "", nil
}
func (f *fakeUsersService) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dusers.MarketUserPosition, error) {
	return nil, nil
}
func (f *fakeUsersService) GetCredentials(ctx context.Context, username string) (*dusers.Credentials, error) {
	return nil, nil
}
func (f *fakeUsersService) UpdatePassword(ctx context.Context, username string, hashedPassword string, mustChange bool) error {
	return nil
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
	users := &fakeUsersService{user: &dusers.User{Username: "alice"}}

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
	users := &fakeUsersService{user: &dusers.User{Username: "alice"}}

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
	users := &fakeUsersService{user: &dusers.User{Username: "alice"}}

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
