package statshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialpredict/handlers"
	analytics "socialpredict/internal/domain/analytics"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestStatsHandlerReturnsServiceBackedConfiguration(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	regularUser := modelstesting.GenerateUser("regular-user", 100)
	regularUser.UserType = "REGULAR"
	adminUser := modelstesting.GenerateUser("admin-user", 100)
	adminUser.UserType = "ADMIN"

	for _, user := range []models.User{regularUser, adminUser} {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}

	config := modelstesting.GenerateEconomicConfig()
	config.Economics.User.InitialAccountBalance = 250
	config.Economics.User.MaximumDebtAllowed = 900
	config.Economics.Betting.BetFees.BuySharesFee = 2
	config.Economics.Betting.BetFees.SellSharesFee = 3
	repo := analytics.NewGormRepository(db)
	statsService := analytics.NewService(repo, analytics.Config{})

	req := httptest.NewRequest(http.MethodGet, "/v0/stats", nil)
	rr := httptest.NewRecorder()

	StatsHandler(statsService, economicsOnlyConfigService{
		economics: configsvc.FromSetup(config).Economics,
	}).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", rr.Code, rr.Body.String())
	}

	var response handlers.SuccessEnvelope[StatsResponse]
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !response.OK {
		t.Fatalf("expected ok=true, got false")
	}

	if response.Result.FinancialStats.TotalMoney != 250 {
		t.Fatalf("expected total money 250, got %d", response.Result.FinancialStats.TotalMoney)
	}

	if response.Result.SetupConfiguration.MaximumDebtAllowed != 900 {
		t.Fatalf("expected max debt 900, got %d", response.Result.SetupConfiguration.MaximumDebtAllowed)
	}

	if response.Result.SetupConfiguration.BuySharesFee != 2 || response.Result.SetupConfiguration.SellSharesFee != 3 {
		t.Fatalf("expected setup fees 2/3, got %d/%d", response.Result.SetupConfiguration.BuySharesFee, response.Result.SetupConfiguration.SellSharesFee)
	}
}

func TestStatsHandlerRequiresConfigService(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v0/stats", nil)
	rr := httptest.NewRecorder()

	StatsHandler(&stubStatsService{}, nil).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	var response handlers.FailureEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.OK {
		t.Fatalf("expected ok=false, got true")
	}

	if response.Reason != string(handlers.ReasonInternalError) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonInternalError, response.Reason)
	}
}

func TestStatsHandlerSanitizesStatsServiceErrors(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v0/stats", nil)
	rr := httptest.NewRecorder()

	StatsHandler(&stubStatsService{err: errors.New("database unavailable")}, configsvc.NewStaticService(modelstesting.GenerateEconomicConfig())).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	var response handlers.FailureEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Reason != string(handlers.ReasonInternalError) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonInternalError, response.Reason)
	}
}

type stubStatsService struct {
	err error
}

func (s *stubStatsService) ComputeFinancialStats(_ context.Context, _ analytics.StatsConfig) (analytics.FinancialStats, error) {
	return analytics.FinancialStats{}, s.err
}

type economicsOnlyConfigService struct {
	economics configsvc.Economics
}

func (s economicsOnlyConfigService) Current() *configsvc.AppConfig {
	panic("Current should not be called")
}

func (s economicsOnlyConfigService) Economics() configsvc.Economics {
	return s.economics
}

func (economicsOnlyConfigService) Frontend() configsvc.Frontend {
	panic("Frontend should not be called")
}

func (economicsOnlyConfigService) ChartSigFigs() int {
	panic("ChartSigFigs should not be called")
}
