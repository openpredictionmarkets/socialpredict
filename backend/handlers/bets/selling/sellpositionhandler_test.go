package sellbetshandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"socialpredict/handlers"
	"socialpredict/handlers/bets/dto"
	bets "socialpredict/internal/domain/bets"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/models/modelstesting"
)

type fakeSellService struct {
	req       bets.SellRequest
	quoteReq  bets.SellRequest
	resp      *bets.SellResult
	quoteResp *bets.SellQuoteResult
	err       error
	quoteErr  error
}

type fakeReadModelInvalidator struct {
	username string
	marketID int64
	reason   string
	calls    int
	err      error
}

func (f *fakeReadModelInvalidator) InvalidateAfterMarketTransaction(ctx context.Context, username string, marketID int64, reason string) error {
	f.username = username
	f.marketID = marketID
	f.reason = reason
	f.calls++
	return f.err
}

func (f *fakeSellService) Place(ctx context.Context, req bets.PlaceRequest) (*bets.PlacedBet, error) {
	return nil, nil
}
func (f *fakeSellService) Sell(ctx context.Context, req bets.SellRequest) (*bets.SellResult, error) {
	f.req = req
	return f.resp, f.err
}
func (f *fakeSellService) QuoteSell(ctx context.Context, req bets.SellRequest) (*bets.SellQuoteResult, error) {
	f.quoteReq = req
	return f.quoteResp, f.quoteErr
}

type fakeUsersService struct {
	user *dusers.User
	err  error
}

func (f *fakeUsersService) GetUser(ctx context.Context, username string) (*dusers.User, error) {
	if f.err != nil {
		return nil, f.err
	}
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
func (f *fakeUsersService) List(ctx context.Context, filters dusers.ListFilters) ([]*dusers.User, error) {
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
		NetProceeds:   55,
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

	var resp handlers.SuccessEnvelope[dto.SellBetResponse]
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got false")
	}
	if resp.Result.SharesSold != 3 || resp.Result.SaleValue != 60 || resp.Result.Dust != 5 || resp.Result.NetProceeds != 55 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestSellPositionHandler_InvalidatesReadModelsAfterSuccess(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	users := &fakeUsersService{user: &dusers.User{Username: "alice"}}
	svc := &fakeSellService{resp: &bets.SellResult{
		Username:      "alice",
		MarketID:      99,
		Outcome:       "YES",
		SharesSold:    2,
		SaleValue:     10,
		Dust:          1,
		NetProceeds:   9,
		TransactionAt: time.Now(),
	}}
	invalidator := &fakeReadModelInvalidator{}
	handler := SellPositionHandlerWithInvalidator(svc, users, invalidator)

	body, _ := json.Marshal(dto.SellBetRequest{MarketID: 99, Amount: 11, Outcome: "YES"})
	req := httptest.NewRequest(http.MethodPost, "/v0/sell", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", rr.Code, rr.Body.String())
	}
	if invalidator.calls != 1 || invalidator.username != "alice" || invalidator.marketID != 99 || invalidator.reason != "sale_accepted" {
		t.Fatalf("unexpected invalidation call: %+v", invalidator)
	}
}

func TestSellQuoteHandler_Success(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	quotedAt := time.Date(2026, time.June, 6, 10, 0, 0, 0, time.UTC)
	svc := &fakeSellService{quoteResp: &bets.SellQuoteResult{
		Username:         "alice",
		MarketID:         7,
		Outcome:          "YES",
		RequestedCredits: 32,
		SharesSold:       3,
		SaleValue:        30,
		Dust:             2,
		NetProceeds:      28,
		MaxDust:          2,
		ValuePerShare:    10,
		DustCapCoverage:  0.3,
		Allowed:          true,
		SuggestedAmounts: []int64{30, 31, 32, 40, 41, 42},
		Message:          "This sale can be submitted.",
		QuotedAt:         quotedAt,
	}}
	users := &fakeUsersService{user: &dusers.User{Username: "alice"}}

	body, _ := json.Marshal(dto.SellBetRequest{MarketID: 7, Amount: 32, Outcome: "YES"})
	req := httptest.NewRequest(http.MethodPost, "/v0/sell/quote", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	SellQuoteHandler(svc, users).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if svc.quoteReq.Username != "alice" || svc.quoteReq.MarketID != 7 || svc.quoteReq.Amount != 32 || svc.quoteReq.Outcome != "YES" {
		t.Fatalf("unexpected quote request: %+v", svc.quoteReq)
	}

	var resp handlers.SuccessEnvelope[dto.SellQuoteResponse]
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Result.Allowed || resp.Result.Dust != 2 || resp.Result.NetProceeds != 28 || resp.Result.MaxDust != 2 || resp.Result.ValuePerShare != 10 {
		t.Fatalf("unexpected quote response: %+v", resp.Result)
	}
	if len(resp.Result.SuggestedAmounts) != 6 {
		t.Fatalf("expected suggestions, got %+v", resp.Result.SuggestedAmounts)
	}
}

func TestSellQuoteHandler_RoundedDustQuoteReturnsAllowedPreview(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	svc := &fakeSellService{quoteResp: &bets.SellQuoteResult{
		Username:         "alice",
		MarketID:         7,
		Outcome:          "YES",
		RequestedCredits: 32,
		SharesSold:       3,
		SaleValue:        30,
		Dust:             2,
		NetProceeds:      28,
		MaxDust:          2,
		ValuePerShare:    10,
		DustCapCoverage:  0.3,
		Allowed:          true,
		SuggestedAmounts: []int64{30, 31, 32},
		Message:          "This Sale Order can be submitted. It would include a 2 credit dust fee from whole-share rounding.",
	}}
	users := &fakeUsersService{user: &dusers.User{Username: "alice"}}

	body, _ := json.Marshal(dto.SellBetRequest{MarketID: 7, Amount: 33, Outcome: "YES"})
	req := httptest.NewRequest(http.MethodPost, "/v0/sell/quote", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	SellQuoteHandler(svc, users).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp handlers.SuccessEnvelope[dto.SellQuoteResponse]
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Result.Allowed || resp.Result.DustCapExceeded || resp.Result.DustCapExceededBy != 0 || resp.Result.Dust != 2 || resp.Result.NetProceeds != 28 {
		t.Fatalf("unexpected rounded quote: %+v", resp.Result)
	}
}

func TestSellPositionHandler_ErrorMapping(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	users := &fakeUsersService{user: &dusers.User{Username: "alice"}}

	cases := []struct {
		name   string
		err    error
		want   int
		reason string
	}{
		{"bad outcome", bets.ErrInvalidOutcome, http.StatusBadRequest, string(handlers.ReasonValidationFailed)},
		{"market closed", bets.ErrMarketClosed, http.StatusConflict, string(handlers.ReasonMarketClosed)},
		{"no position", bets.ErrNoPosition, http.StatusUnprocessableEntity, string(handlers.ReasonNoPosition)},
		{"insufficient shares", bets.ErrInsufficientShares, http.StatusUnprocessableEntity, string(handlers.ReasonInsufficientShares)},
		{"dust cap", bets.ErrDustCapExceeded{Cap: 2, Requested: 3}, http.StatusUnprocessableEntity, string(handlers.ReasonDustCapExceeded)},
		{"market not found", dmarkets.ErrMarketNotFound, http.StatusNotFound, string(handlers.ReasonMarketNotFound)},
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

			var resp handlers.FailureEnvelope
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode failure envelope: %v", err)
			}
			if resp.Reason != tc.reason {
				t.Fatalf("expected reason %q, got %q", tc.reason, resp.Reason)
			}
		})
	}
}

func TestSellPositionHandler_DustCapExceededIncludesUserGuidance(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	svc := &fakeSellService{err: bets.ErrDustCapExceeded{Cap: 2, Requested: 3}}
	users := &fakeUsersService{user: &dusers.User{Username: "alice"}}

	body, _ := json.Marshal(dto.SellBetRequest{MarketID: 1, Amount: 10, Outcome: "YES"})
	req := httptest.NewRequest(http.MethodPost, "/v0/sell", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	SellPositionHandler(svc, users).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, rr.Code)
	}

	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if resp.Reason != string(handlers.ReasonDustCapExceeded) {
		t.Fatalf("expected dust cap reason, got %q", resp.Reason)
	}
	if resp.Message == "" {
		t.Fatalf("expected user-facing guidance message")
	}
	if resp.Details["dust"] != float64(3) || resp.Details["maxDust"] != float64(2) {
		t.Fatalf("expected dust details, got %+v", resp.Details)
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

	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if resp.Reason != string(handlers.ReasonInvalidRequest) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonInvalidRequest, resp.Reason)
	}
}

func TestSellPositionHandler_PasswordChangeRequired(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	svc := &fakeSellService{}
	users := &fakeUsersService{user: &dusers.User{Username: "alice", MustChangePassword: true}}

	body, _ := json.Marshal(dto.SellBetRequest{MarketID: 7, Amount: 65, Outcome: "YES"})
	req := httptest.NewRequest(http.MethodPost, "/v0/sell", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	SellPositionHandler(svc, users).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rr.Code)
	}

	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if resp.Reason != string(handlers.ReasonPasswordChangeRequired) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonPasswordChangeRequired, resp.Reason)
	}
}

func TestSellPositionHandler_UserNotFound(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	svc := &fakeSellService{}
	users := &fakeUsersService{err: dusers.ErrUserNotFound}

	body, _ := json.Marshal(dto.SellBetRequest{MarketID: 7, Amount: 65, Outcome: "YES"})
	req := httptest.NewRequest(http.MethodPost, "/v0/sell", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	SellPositionHandler(svc, users).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}

	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if resp.Reason != string(handlers.ReasonUserNotFound) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonUserNotFound, resp.Reason)
	}
}
