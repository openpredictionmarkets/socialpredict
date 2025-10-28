package usercredit

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

type creditServiceMock struct {
	credit          int64
	err             error
	lastUsername    string
	lastMaximumDebt int64
}

func (m *creditServiceMock) GetPublicUser(context.Context, string) (*dusers.PublicUser, error) {
	return nil, nil
}

func (m *creditServiceMock) ApplyTransaction(context.Context, string, int64, string) error {
	return nil
}

func (m *creditServiceMock) GetUserCredit(_ context.Context, username string, maximumDebt int64) (int64, error) {
	m.lastUsername = username
	m.lastMaximumDebt = maximumDebt
	if m.err != nil {
		return 0, m.err
	}
	return m.credit, nil
}

func (m *creditServiceMock) GetUserPortfolio(context.Context, string) (*dusers.Portfolio, error) {
	return nil, nil
}

func TestGetUserCreditHandlerSuccess(t *testing.T) {
	mock := &creditServiceMock{credit: 750}
	handler := GetUserCreditHandler(mock, 500)

	req := httptest.NewRequest(http.MethodGet, "/v0/usercredit/alice", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body dto.UserCreditResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if body.Credit != 750 {
		t.Fatalf("expected credit 750, got %d", body.Credit)
	}
	if mock.lastUsername != "alice" || mock.lastMaximumDebt != 500 {
		t.Fatalf("unexpected parameters passed to service: username=%s maxDebt=%d", mock.lastUsername, mock.lastMaximumDebt)
	}
}

func TestGetUserCreditHandlerMethodNotAllowed(t *testing.T) {
	handler := GetUserCreditHandler(&creditServiceMock{}, 500)

	req := httptest.NewRequest(http.MethodPost, "/v0/usercredit/alice", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}

func TestGetUserCreditHandlerInternalError(t *testing.T) {
	handler := GetUserCreditHandler(&creditServiceMock{err: errors.New("boom")}, 500)

	req := httptest.NewRequest(http.MethodGet, "/v0/usercredit/alice", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}
