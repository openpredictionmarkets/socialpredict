package publicuser

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
)

type portfolioServiceMock struct {
	portfolio *dusers.Portfolio
	err       error
}

func (m *portfolioServiceMock) GetPublicUser(context.Context, string) (*dusers.PublicUser, error) {
	return nil, nil
}

func (m *portfolioServiceMock) ApplyTransaction(context.Context, string, int64, string) error {
	return nil
}

func (m *portfolioServiceMock) GetUserCredit(context.Context, string, int64) (int64, error) {
	return 0, nil
}

func (m *portfolioServiceMock) GetUserPortfolio(context.Context, string) (*dusers.Portfolio, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.portfolio, nil
}

func (m *portfolioServiceMock) GetUserFinancials(context.Context, string) (map[string]int64, error) {
	return nil, nil
}

func (m *portfolioServiceMock) ListUserMarkets(context.Context, int64) ([]*dusers.UserMarket, error) {
	return nil, nil
}

func (m *portfolioServiceMock) UpdateDescription(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *portfolioServiceMock) UpdateDisplayName(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *portfolioServiceMock) UpdateEmoji(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *portfolioServiceMock) UpdatePersonalLinks(context.Context, string, dusers.PersonalLinks) (*dusers.User, error) {
	return nil, nil
}

func (m *portfolioServiceMock) ChangePassword(context.Context, string, string, string) error {
	return nil
}

func TestGetPortfolioHandlerSuccess(t *testing.T) {
	portfolio := &dusers.Portfolio{
		Items: []dusers.PortfolioItem{
			{
				MarketID:       1,
				QuestionTitle:  "Test Market",
				YesSharesOwned: 10,
				NoSharesOwned:  5,
				LastBetPlaced:  time.Now(),
			},
		},
		TotalSharesOwned: 15,
	}
	mock := &portfolioServiceMock{portfolio: portfolio}
	handler := GetPortfolioHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/v0/portfolio/alice", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body dto.PortfolioResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(body.PortfolioItems) != 1 || body.TotalSharesOwned != 15 {
		t.Fatalf("unexpected response: %+v", body)
	}
	if body.PortfolioItems[0].QuestionTitle != "Test Market" {
		t.Fatalf("expected question title 'Test Market', got %q", body.PortfolioItems[0].QuestionTitle)
	}
}

func TestGetPortfolioHandlerInvalidMethod(t *testing.T) {
	handler := GetPortfolioHandler(&portfolioServiceMock{})
	req := httptest.NewRequest(http.MethodPost, "/v0/portfolio/alice", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}

func TestGetPortfolioHandlerServiceError(t *testing.T) {
	mock := &portfolioServiceMock{err: errors.New("boom")}
	handler := GetPortfolioHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/v0/portfolio/alice", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}
