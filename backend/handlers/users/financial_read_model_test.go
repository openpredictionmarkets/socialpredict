package usershandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	analytics "socialpredict/internal/domain/analytics"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

type financialReadModelServiceMock struct {
	readModel *analytics.UserFinancialMetricReadModel
	err       error
}

func (m *financialReadModelServiceMock) GetUserFinancialMetricReadModel(context.Context, string) (*analytics.UserFinancialMetricReadModel, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.readModel, nil
}

type financialReadModelAuthMock struct {
	err *authsvc.AuthError
}

func (m financialReadModelAuthMock) CurrentUser(*http.Request) (*dusers.User, *authsvc.AuthError) {
	if m.err != nil {
		return nil, m.err
	}
	return &dusers.User{Username: "viewer"}, nil
}

func (m financialReadModelAuthMock) RequireUser(r *http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.CurrentUser(r)
}

func (m financialReadModelAuthMock) RequireAdmin(*http.Request) (*dusers.User, *authsvc.AuthError) {
	return nil, &authsvc.AuthError{Kind: authsvc.ErrorKindAdminRequired, Message: "admin required"}
}

func TestGetUserFinancialReadModelHandlerRequiresLogin(t *testing.T) {
	handler := GetUserFinancialReadModelHandler(&financialReadModelServiceMock{}, financialReadModelAuthMock{
		err: &authsvc.AuthError{Kind: authsvc.ErrorKindMissingToken, Message: "missing token"},
	})
	req := httptest.NewRequest(http.MethodGet, "/v0/read/users/alice/financial-summary", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
	requireFinancialFailureReason(t, rec, handlers.ReasonInvalidToken)
}

func TestGetUserFinancialReadModelHandlerReturnsFreshness(t *testing.T) {
	generatedAt := time.Date(2026, 6, 7, 15, 0, 0, 0, time.UTC)
	snapshot := analytics.UserFinancialMetricSnapshot{
		Username:      "alice",
		GeneratedAt:   generatedAt,
		PositionCount: 2,
		Financial: analytics.FinancialSnapshot{
			AccountBalance: 500,
			AmountInPlay:   120,
		},
		Source:              "read_model",
		TransactionSafeRead: false,
	}
	handler := GetUserFinancialReadModelHandler(&financialReadModelServiceMock{
		readModel: &analytics.UserFinancialMetricReadModel{
			Snapshot:  snapshot,
			Freshness: snapshot.Freshness(),
		},
	}, financialReadModelAuthMock{})
	req := httptest.NewRequest(http.MethodGet, "/v0/read/users/alice/financial-summary", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp handlers.SuccessEnvelope[userFinancialReadModelResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Result.Username != "alice" {
		t.Fatalf("username = %q, want alice", resp.Result.Username)
	}
	if resp.Result.Financial["accountBalance"] != 500 || resp.Result.Financial["amountInPlay"] != 120 {
		t.Fatalf("unexpected financial map: %+v", resp.Result.Financial)
	}
	if resp.Result.PositionCount != 2 {
		t.Fatalf("position count = %d, want 2", resp.Result.PositionCount)
	}
	if !resp.Result.Freshness.GeneratedAt.Equal(generatedAt) {
		t.Fatalf("freshness generated at = %s, want %s", resp.Result.Freshness.GeneratedAt, generatedAt)
	}
	if resp.Result.Freshness.TransactionSafeRead {
		t.Fatalf("freshness should not be transaction safe")
	}
}
