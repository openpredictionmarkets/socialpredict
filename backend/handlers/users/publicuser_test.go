package usershandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
)

type publicUserServiceMock struct {
	user *dusers.PublicUser
	err  error
}

func (m *publicUserServiceMock) GetPublicUser(_ context.Context, _ string) (*dusers.PublicUser, error) {
	return m.user, m.err
}

func (m *publicUserServiceMock) ApplyTransaction(context.Context, string, int64, string) error {
	return nil
}

func (m *publicUserServiceMock) GetUser(context.Context, string) (*dusers.User, error) {
	return nil, nil
}

func (m *publicUserServiceMock) GetUserCredit(context.Context, string, int64) (int64, error) {
	return 0, nil
}

func (m *publicUserServiceMock) GetUserPortfolio(context.Context, string) (*dusers.Portfolio, error) {
	return nil, nil
}

func (m *publicUserServiceMock) GetUserFinancials(context.Context, string) (map[string]int64, error) {
	return nil, nil
}

func (m *publicUserServiceMock) ListUserMarkets(context.Context, int64) ([]*dusers.UserMarket, error) {
	return nil, nil
}

func (m *publicUserServiceMock) UpdateDescription(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *publicUserServiceMock) UpdateDisplayName(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *publicUserServiceMock) UpdateEmoji(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *publicUserServiceMock) UpdatePersonalLinks(context.Context, string, dusers.PersonalLinks) (*dusers.User, error) {
	return nil, nil
}

func (m *publicUserServiceMock) GetPrivateProfile(context.Context, string) (*dusers.PrivateProfile, error) {
	return nil, nil
}

func (m *publicUserServiceMock) ChangePassword(context.Context, string, string, string) error {
	return nil
}

func (m *publicUserServiceMock) MustChangePassword(context.Context, string) (bool, error) {
	return false, nil
}

func TestGetPublicUserHandlerReturnsPublicUser(t *testing.T) {
	mockUser := &dusers.PublicUser{
		Username:              "alice",
		DisplayName:           "Alice",
		UserType:              "regular",
		InitialAccountBalance: 1000,
		AccountBalance:        750,
		PersonalEmoji:         "🌟",
		Description:           "hello",
		PersonalLink1:         "https://example.com",
	}
	handler := GetPublicUserHandler(&publicUserServiceMock{user: mockUser})

	req := httptest.NewRequest(http.MethodGet, "/v0/userinfo/alice", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body dto.PublicUserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if body.Username != mockUser.Username || body.DisplayName != mockUser.DisplayName {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestGetPublicUserHandlerNotFound(t *testing.T) {
	handler := GetPublicUserHandler(&publicUserServiceMock{err: dusers.ErrUserNotFound})

	req := httptest.NewRequest(http.MethodGet, "/v0/userinfo/missing", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "missing"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestGetPublicUserHandlerInternalError(t *testing.T) {
	handler := GetPublicUserHandler(&publicUserServiceMock{err: errors.New("boom")})

	req := httptest.NewRequest(http.MethodGet, "/v0/userinfo/alice", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}
