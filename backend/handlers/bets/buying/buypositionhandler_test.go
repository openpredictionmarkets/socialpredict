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
	usermodels "socialpredict/internal/domain/users/models"
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

type fakeUsersService struct {
	user *dusers.User
}

func (f *fakeUsersService) GetPublicUser(ctx context.Context, username string) (*dusers.PublicUser, error) {
	return nil, nil
}
func (f *fakeUsersService) GetUser(ctx context.Context, username string) (*dusers.User, error) {
	return f.user, nil
}
func (f *fakeUsersService) GetPrivateProfile(ctx context.Context, username string) (*dusers.PrivateProfile, error) {
	return nil, nil
}
func (f *fakeUsersService) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error {
	return nil
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

func TestPlaceBetHandler_Success(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	betsSvc := &fakeBetsService{resp: &bets.PlacedBet{Username: "alice", MarketID: 5, Amount: 120, Outcome: "YES", PlacedAt: time.Now()}}
	userSvc := &fakeUsersService{user: &dusers.User{Username: "alice"}}

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
	userSvc := &fakeUsersService{user: &dusers.User{Username: "alice"}}

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
	userSvc := &fakeUsersService{user: &dusers.User{Username: "alice"}}
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
