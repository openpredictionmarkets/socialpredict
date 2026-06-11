package metricshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"socialpredict/handlers"
	analytics "socialpredict/internal/domain/analytics"
	"socialpredict/internal/domain/boundary"
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/internal/domain/readmodels"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

func newAnalyticsService(t *testing.T, db *gorm.DB) *analytics.Service {
	t.Helper()
	cfg := modelstesting.GenerateEconomicConfig()
	return analytics.NewService(analytics.NewGormRepository(db), analytics.Config{
		MaximumDebtAllowed: cfg.Economics.User.MaximumDebtAllowed,
		CreateMarketCost:   cfg.Economics.MarketIncentives.CreateMarketCost,
		InitialBetFee:      cfg.Economics.Betting.BetFees.InitialBetFee,
	})
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

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var payload handlers.SuccessEnvelope[map[string]interface{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !payload.OK {
		t.Fatalf("expected ok=true, got false")
	}
	if payload.Result["moneyCreated"] == nil {
		t.Fatalf("expected moneyCreated section in response: %+v", payload)
	}
	if payload.Result["freshness"] == nil {
		t.Fatalf("expected freshness section in response: %+v", payload)
	}
}

type cachedSystemMetricsService struct {
	readModel    *analytics.SystemMetricsReadModel
	refreshCalls int
}

func (s *cachedSystemMetricsService) ComputeSystemMetrics(context.Context) (*analytics.SystemMetrics, error) {
	return &analytics.SystemMetrics{}, nil
}

func (s *cachedSystemMetricsService) GetSystemMetricsReadModel(context.Context) (*analytics.SystemMetricsReadModel, error) {
	return s.readModel, nil
}

func (s *cachedSystemMetricsService) RefreshSystemMetricsSnapshot(context.Context) (*analytics.SystemMetricsReadModel, error) {
	s.refreshCalls++
	return &analytics.SystemMetricsReadModel{
		Metrics: analytics.SystemMetrics{},
		Freshness: readmodels.NewFreshness(
			time.Now().UTC(),
			"read_model",
			analytics.SystemMetricsSnapshotTargetFreshness,
			false,
		),
	}, nil
}

func TestSystemMetricsReadModel_ServesStaleSnapshotUntilAgeExpires(t *testing.T) {
	generatedAt := time.Now().UTC().Add(-5 * time.Minute)
	svc := &cachedSystemMetricsService{
		readModel: &analytics.SystemMetricsReadModel{
			Metrics: analytics.SystemMetrics{},
			Freshness: readmodels.NewStaleFreshness(
				generatedAt,
				"read_model",
				analytics.SystemMetricsSnapshotTargetFreshness,
				false,
				"bet_accepted",
				nil,
			),
		},
	}

	_, freshness, err := systemMetricsReadModel(context.Background(), svc)
	if err != nil {
		t.Fatalf("systemMetricsReadModel returned error: %v", err)
	}
	if freshness == nil || !freshness.IsStale {
		t.Fatalf("expected stale freshness to be served, got %+v", freshness)
	}
	if svc.refreshCalls != 0 {
		t.Fatalf("expected no refresh for stale-but-young snapshot, got %d", svc.refreshCalls)
	}
}

type failingAnalyticsRepo struct{}

func (failingAnalyticsRepo) ListUsers(context.Context) ([]analytics.UserAccount, error) {
	return nil, errors.New("boom")
}

func (failingAnalyticsRepo) ListMarkets(context.Context) ([]analytics.MarketRecord, error) {
	return nil, nil
}

func (failingAnalyticsRepo) ListBetsForMarket(context.Context, uint) ([]boundary.Bet, error) {
	return nil, nil
}

func (failingAnalyticsRepo) ListBetsOrdered(context.Context) ([]boundary.Bet, error) {
	return nil, nil
}

func (failingAnalyticsRepo) UserMarketPositions(context.Context, string) ([]positionsmath.MarketPosition, error) {
	return nil, nil
}

func (failingAnalyticsRepo) UserWorkProfitResolvedMarkets(context.Context, string) ([]analytics.WorkProfitMarketRecord, error) {
	return nil, nil
}

func (failingAnalyticsRepo) CountUsersByType(context.Context, string) (int64, error) {
	return 0, nil
}

func TestGetSystemMetricsHandler_Error(t *testing.T) {
	cfg := modelstesting.GenerateEconomicConfig()
	svc := analytics.NewService(failingAnalyticsRepo{}, analytics.Config{
		MaximumDebtAllowed: cfg.Economics.User.MaximumDebtAllowed,
		CreateMarketCost:   cfg.Economics.MarketIncentives.CreateMarketCost,
		InitialBetFee:      cfg.Economics.Betting.BetFees.InitialBetFee,
	})

	handler := GetSystemMetricsHandler(svc)
	req := httptest.NewRequest("GET", "/v0/system/metrics", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != 500 {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	var payload handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal failure: %v", err)
	}
	if payload.Reason != string(handlers.ReasonInternalError) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonInternalError, payload.Reason)
	}
}
