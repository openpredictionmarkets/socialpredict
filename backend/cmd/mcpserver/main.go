package main

import (
	"socialpredict/handlers/cms/marketdiscovery"
	"socialpredict/internal/app"
	appenv "socialpredict/internal/app/env"
	appruntime "socialpredict/internal/app/runtime"
	"socialpredict/internal/mcpserver"
	"socialpredict/logger"
	"socialpredict/migration"
	_ "socialpredict/migration/migrations"
	"socialpredict/seed"
	"socialpredict/setup"
)

func main() {
	readiness := appruntime.NewReadiness()

	if err := appenv.LoadDevFile(); err != nil {
		logger.Warn("mcpserver", "development environment override not loaded", logger.Operation("LoadDevFile"), logger.Err(err))
	}

	dbCfg, err := appruntime.LoadDBConfigFromEnv()
	if err != nil {
		logger.Fatal("mcpserver", "database configuration unavailable", err, logger.Operation("LoadDBConfigFromEnv"))
	}
	db, err := appruntime.InitDB(dbCfg, appruntime.PostgresFactory{})
	if err != nil {
		logger.Fatal("mcpserver", "database initialization failed", err, logger.Operation("InitDB"))
	}
	defer func() {
		if err := appruntime.CloseDB(db); err != nil {
			logger.Warn("mcpserver", "database shutdown reported a warning", logger.Operation("CloseDB"), logger.Err(err))
		}
	}()

	if err := seed.EnsureDBReady(db, 20); err != nil {
		logger.Fatal("mcpserver", "database readiness check failed", err, logger.Operation("EnsureDBReady"))
	}

	configService, err := appruntime.LoadConfigService(setup.EmbeddedSource{})
	if err != nil {
		logger.Fatal("mcpserver", "configuration service initialization failed", err, logger.Operation("LoadConfigService"))
	}

	startupMode, err := appruntime.LoadStartupMutationModeFromEnv()
	if err != nil {
		logger.Fatal("mcpserver", "startup mutation mode unavailable", err, logger.Operation("LoadStartupMutationModeFromEnv"))
	}
	if err := appruntime.RunStartupMutations(db, configService, startupMode, appruntime.StartupMutationHooks{
		Migrate:      migration.MigrateDB,
		Verify:       migration.VerifyApplied,
		SeedUsers:    seed.SeedUsers,
		SeedHomepage: seed.SeedHomepage,
	}); err != nil {
		logger.Fatal("mcpserver", "startup database schema incompatible", err, logger.Operation("RunStartupMutations"))
	}

	shutdownConfig, err := appruntime.LoadShutdownConfigFromEnv()
	if err != nil {
		logger.Fatal("mcpserver", "shutdown configuration unavailable", err, logger.Operation("LoadShutdownConfigFromEnv"))
	}

	container := app.BuildApplicationWithConfigService(db, configService)
	discoveryRepo := marketdiscovery.NewGormRepository(db)
	discoveryService := marketdiscovery.NewService(discoveryRepo)
	runtime := mcpserver.NewRuntime(container.GetMarketsService(), discoveryService)
	readiness.MarkReady()

	handler := mcpserver.BuildHTTPHandler(runtime, readiness, appruntime.NewServingProbe(db, readiness))
	mcpserver.StartHTTP(handler, readiness, shutdownConfig)
}
