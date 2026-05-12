package analytics_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"socialpredict/internal/app"
	"socialpredict/internal/domain/analytics"
	dbets "socialpredict/internal/domain/bets"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestResolveMarketDistributesAllBetVolumePostgres(t *testing.T) {
	dsn, ok := analyticsPostgresIntegrationDSN()
	if !ok {
		t.Skip("set SOCIALPREDICT_POSTGRES_TEST_DSN or POSTGRES_TEST_DSN to run real-Postgres market-resolution verification")
	}

	db := openAnalyticsPostgresDB(t, dsn)
	econConfig, _ := modelstesting.UseStandardTestEconomics(t)

	users := []models.User{
		modelstesting.GenerateUser("pg_creator", 0),
		modelstesting.GenerateUser("pg_no_winner", 0),
		modelstesting.GenerateUser("pg_yes_winner", 0),
		modelstesting.GenerateUser("pg_yes_second", 0),
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user %s: %v", users[i].Username, err)
		}
	}

	market := modelstesting.GenerateMarket(time.Now().UnixNano()%1_000_000, users[0].Username)
	market.IsResolved = false
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	creationFee := econConfig.Economics.MarketIncentives.CreateMarketCost
	if err := modelstesting.AdjustUserBalance(db, users[0].Username, -creationFee); err != nil {
		t.Fatalf("apply creation fee: %v", err)
	}

	container := app.BuildApplicationWithConfigService(db, configsvc.NewStaticService(econConfig))
	betsService := container.GetBetsService()
	placeBet := func(username string, amount int64, outcome string) {
		t.Helper()
		if _, err := betsService.Place(context.Background(), dbets.PlaceRequest{
			Username: username,
			MarketID: uint(market.ID),
			Amount:   amount,
			Outcome:  outcome,
		}); err != nil {
			t.Fatalf("place bet for %s: %v", username, err)
		}
	}

	placeBet(users[0].Username, 50, "NO")
	placeBet(users[1].Username, 51, "NO")
	placeBet(users[1].Username, 51, "NO")
	placeBet(users[2].Username, 10, "YES")
	placeBet(users[3].Username, 30, "YES")

	if err := container.GetMarketsService().ResolveMarket(context.Background(), int64(market.ID), "YES", market.CreatorUsername); err != nil {
		t.Fatalf("ResolveMarket: %v", err)
	}

	metricsSvc := newAnalyticsMetricsService(db, analyticsConfigFromSetup(econConfig))
	metrics := requireAnalyticsSystemMetrics(t, metricsSvc)
	if surplus := metrics.Verification.SurplusValue(); surplus != 0 {
		t.Fatalf("expected zero surplus after Postgres resolution, got %d", surplus)
	}

	repo := analytics.NewGormRepository(db)
	for _, user := range users {
		positions, err := repo.UserMarketPositions(context.Background(), user.Username)
		if err != nil {
			t.Fatalf("calculate positions for %s: %v", user.Username, err)
		}
		for _, pos := range positions {
			if pos.YesSharesOwned > 0 && pos.NoSharesOwned > 0 {
				t.Fatalf("user %s holds both YES and NO shares post-resolution", user.Username)
			}
		}
	}
}

func openAnalyticsPostgresDB(t *testing.T, dsn string) *gorm.DB {
	t.Helper()

	adminDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres test database: %v", err)
	}
	adminSQL, err := adminDB.DB()
	if err != nil {
		t.Fatalf("postgres admin sql handle: %v", err)
	}
	t.Cleanup(func() { _ = adminSQL.Close() })

	schema := fmt.Sprintf("sp_market_resolution_%d", time.Now().UnixNano())
	if err := adminDB.Exec(`CREATE SCHEMA ` + quoteAnalyticsPostgresIdentifier(schema)).Error; err != nil {
		t.Fatalf("create isolated postgres schema: %v", err)
	}
	t.Cleanup(func() {
		_ = adminDB.Exec(`DROP SCHEMA IF EXISTS ` + quoteAnalyticsPostgresIdentifier(schema) + ` CASCADE`).Error
	})

	db, err := gorm.Open(postgres.Open(analyticsPostgresDSNWithSearchPath(dsn, schema)), &gorm.Config{})
	if err != nil {
		t.Fatalf("open isolated postgres schema: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("postgres sql handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(4)
	t.Cleanup(func() { _ = sqlDB.Close() })

	if db.Dialector.Name() != "postgres" {
		t.Fatalf("expected postgres dialector, got %q", db.Dialector.Name())
	}
	if err := db.AutoMigrate(&models.User{}, &models.Market{}, &models.Bet{}); err != nil {
		t.Fatalf("migrate analytics resolution tables: %v", err)
	}

	return db
}

func analyticsPostgresIntegrationDSN() (string, bool) {
	for _, key := range []string{"SOCIALPREDICT_POSTGRES_TEST_DSN", "POSTGRES_TEST_DSN"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value, true
		}
	}
	return "", false
}

func analyticsPostgresDSNWithSearchPath(dsn string, schema string) string {
	if parsed, err := url.Parse(dsn); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}

	return strings.TrimSpace(dsn) + " search_path=" + schema
}

func quoteAnalyticsPostgresIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}
