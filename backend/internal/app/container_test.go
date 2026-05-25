package app

import (
	"context"
	"testing"
	"time"

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

func TestBuildApplicationPassesGameModeToMarketsService(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	config := modelstesting.GenerateEconomicConfig()
	config.Game.Mode = configsvc.GameModeModerator
	config.Game.Moderation.MarketApprovalRequired = true

	container := BuildApplicationWithConfigService(db, configsvc.NewStaticService(config))
	if container == nil {
		t.Fatalf("BuildApplicationWithConfigService returned nil container")
	}

	user := modelstesting.GenerateUser("moderator", 1000)
	user.UserType = "MODERATOR"
	user.ModeratorStatus = "active"
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create moderator: %v", err)
	}

	market, err := container.GetMarketsService().CreateMarket(context.Background(), dmarkets.MarketCreateRequest{
		QuestionTitle:      "Will moderator mode create a proposal?",
		Description:        "Container wiring test",
		OutcomeType:        "BINARY",
		ResolutionDateTime: container.clock.Now().Add(48 * time.Hour),
	}, user.Username)
	if err != nil {
		t.Fatalf("CreateMarket returned error: %v", err)
	}
	if market.Status != dmarkets.MarketLifecycleProposed || market.LifecycleStatus != dmarkets.MarketLifecycleProposed {
		t.Fatalf("expected proposed market from container wiring, got %+v", market)
	}
}
