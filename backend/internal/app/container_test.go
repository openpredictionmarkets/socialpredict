package app

import (
	"context"
	"testing"

	dmarkets "socialpredict/internal/domain/markets"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/models/modelstesting"
)

func TestBuildApplicationWiresMarketsDependencies(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	config := modelstesting.GenerateEconomicConfig()
	if config == nil {
		t.Fatalf("expected economic config, got nil")
	}

	container := BuildApplicationWithConfigService(db, configsvc.NewStaticService(config))
	if container == nil {
		t.Fatalf("BuildApplicationWithConfigService returned nil container")
	}

	marketsService := container.GetMarketsService()
	if marketsService == nil {
		t.Fatalf("expected markets service to be initialized")
	}

	usersService := container.GetUsersService()
	if usersService == nil {
		t.Fatalf("expected users service to be initialized")
	}

	betsService := container.GetBetsService()
	if betsService == nil {
		t.Fatalf("expected bets service to be initialized")
	}

	marketsHandler := container.GetMarketsHandler()
	if marketsHandler == nil {
		t.Fatalf("expected markets handler to be initialized")
	}

	if container.GetConfigService() == nil {
		t.Fatalf("expected config service to be initialized")
	}

	if container.GetSecurityService() == nil {
		t.Fatalf("expected security service to be initialized")
	}

	if _, err := marketsService.ListMarkets(context.Background(), dmarkets.ListFilters{}); err != nil {
		t.Fatalf("ListMarkets should work against initialized repository, got error: %v", err)
	}
}

func TestBuildApplicationRetainsCompatibilityWithRawConfig(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	config := modelstesting.GenerateEconomicConfig()

	container := BuildApplication(db, config)
	if container == nil {
		t.Fatalf("BuildApplication returned nil container")
	}
	if container.GetConfigService() == nil {
		t.Fatalf("expected config service to be initialized from raw config")
	}
}
