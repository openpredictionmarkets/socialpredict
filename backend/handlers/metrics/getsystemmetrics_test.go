package metricshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	analytics "socialpredict/internal/domain/analytics"
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/setup"

	"gorm.io/gorm"
)

func newAnalyticsService(t *testing.T, db *gorm.DB) *analytics.Service {
	t.Helper()
	cfg := modelstesting.GenerateEconomicConfig()
	loader := func() *setup.EconomicConfig { return cfg }
	return analytics.NewService(analytics.NewGormRepository(db), loader)
}

func TestGetSystemMetricsHandler_Success(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)

	user := modelstesting.GenerateUser("alice", 0)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	handler := GetSystemMetricsHandler(newAnalyticsService(t, db))
	req := httptest.NewRequest("GET", "/v0/system/metrics", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload["moneyCreated"] == nil {
		t.Fatalf("expected moneyCreated section in response: %+v", payload)
	}
}

type failingAnalyticsRepo struct{}

func (failingAnalyticsRepo) ListUsers(context.Context) ([]models.User, error) {
	return nil, errors.New("boom")
}

func (failingAnalyticsRepo) ListMarkets(context.Context) ([]models.Market, error) {
	return nil, nil
}

func (failingAnalyticsRepo) ListBetsForMarket(context.Context, uint) ([]models.Bet, error) {
	return nil, nil
}

func (failingAnalyticsRepo) ListBetsOrdered(context.Context) ([]models.Bet, error) {
	return nil, nil
}

func (failingAnalyticsRepo) UserMarketPositions(context.Context, string) ([]positionsmath.MarketPosition, error) {
	return nil, nil
}

func TestGetSystemMetricsHandler_Error(t *testing.T) {
	cfg := modelstesting.GenerateEconomicConfig()
	loader := func() *setup.EconomicConfig { return cfg }
	svc := analytics.NewService(failingAnalyticsRepo{}, loader)

	handler := GetSystemMetricsHandler(svc)
	req := httptest.NewRequest("GET", "/v0/system/metrics", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != 500 {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}
