package usershandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	dusers "socialpredict/internal/domain/users"
)

type financialServiceMock struct {
	snapshot map[string]int64
	err      error
}

func (m *financialServiceMock) GetPublicUser(context.Context, string) (*dusers.PublicUser, error) {
	return nil, nil
}

func (m *financialServiceMock) ApplyTransaction(context.Context, string, int64, string) error {
	return nil
}

func (m *financialServiceMock) GetUserCredit(context.Context, string, int64) (int64, error) {
	return 0, nil
}

func (m *financialServiceMock) GetUserPortfolio(context.Context, string) (*dusers.Portfolio, error) {
	return nil, nil
}

func (m *financialServiceMock) GetUserFinancials(context.Context, string) (map[string]int64, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.snapshot, nil
}

func (m *financialServiceMock) ListUserMarkets(context.Context, int64) ([]*dusers.UserMarket, error) {
	return nil, nil
}

func (m *financialServiceMock) UpdateDescription(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *financialServiceMock) UpdateDisplayName(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *financialServiceMock) UpdateEmoji(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *financialServiceMock) UpdatePersonalLinks(context.Context, string, dusers.PersonalLinks) (*dusers.User, error) {
	return nil, nil
}

func (m *financialServiceMock) ChangePassword(context.Context, string, string, string) error {
	return nil
}

func TestGetUserFinancialHandlerSuccess(t *testing.T) {
	mock := &financialServiceMock{snapshot: map[string]int64{"accountBalance": 500}}
	handler := GetUserFinancialHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/v0/users/alice/financial", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var wrapper map[string]map[string]int64
	if err := json.Unmarshal(rec.Body.Bytes(), &wrapper); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if wrapper["financial"]["accountBalance"] != 500 {
		t.Fatalf("expected accountBalance 500, got %d", wrapper["financial"]["accountBalance"])
	}
}

func TestGetUserFinancialHandlerUserNotFound(t *testing.T) {
	handler := GetUserFinancialHandler(&financialServiceMock{err: dusers.ErrUserNotFound})
	req := httptest.NewRequest(http.MethodGet, "/v0/users/missing/financial", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "missing"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestGetUserFinancialHandlerInternalError(t *testing.T) {
	handler := GetUserFinancialHandler(&financialServiceMock{err: errors.New("boom")})
	req := httptest.NewRequest(http.MethodGet, "/v0/users/alice/financial", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

func TestGetUserFinancialHandlerInvalidMethod(t *testing.T) {
	handler := GetUserFinancialHandler(&financialServiceMock{})
	req := httptest.NewRequest(http.MethodPost, "/v0/users/alice/financial", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}
