package main

import (
	"net/http"
	"os"

	appenv "socialpredict/internal/app/env"
	appruntime "socialpredict/internal/app/runtime"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/logger"
	"socialpredict/migration"
	_ "socialpredict/migration/migrations" // <-- side-effect import: registers migrations via init()
	"socialpredict/seed"
	"socialpredict/server"
	"socialpredict/setup"
)

func main() {
	// Secure endpoint example
	http.Handle("/secure", authsvc.Authenticate(http.HandlerFunc(secureEndpoint)))

	readiness := appruntime.NewReadiness()

	// Load local development env overrides from .env.dev when present.
	if err := appenv.LoadDevFile(); err != nil {
		logger.Warn("startup", "development environment override not loaded", logger.Operation("LoadDevFile"), logger.Err(err))
	}

	dbCfg, err := appruntime.LoadDBConfigFromEnv()
	if err != nil {
		logger.Fatal("startup", "database configuration unavailable", err, logger.Operation("LoadDBConfigFromEnv"))
	}

	db, err := appruntime.InitDB(dbCfg, appruntime.PostgresFactory{})
	if err != nil {
		logger.Fatal("startup", "database initialization failed", err, logger.Operation("InitDB"))
	}

	const MAX_ATTEMPTS = 20
	if err := seed.EnsureDBReady(db, MAX_ATTEMPTS); err != nil {
		logger.Fatal("startup", "database readiness check failed", err, logger.Operation("EnsureDBReady"))
	}

	if err := migration.MigrateDB(db); err != nil {
		logger.Warn("startup", "database migration reported a warning", logger.Operation("MigrateDB"), logger.Err(err))
	}

	configService, err := appruntime.LoadConfigService(setup.EmbeddedSource{})
	if err != nil {
		logger.Fatal("startup", "configuration service initialization failed", err, logger.Operation("LoadConfigService"))
	}

	if err := seed.SeedUsers(db, configService); err != nil {
		logger.Fatal("startup", "user seed failed", err, logger.Operation("SeedUsers"))
	}
	if err := seed.SeedHomepage(db, "."); err != nil {
		logger.Warn("startup", "homepage seed reported a warning", logger.Operation("SeedHomepage"), logger.Err(err))
	}

	readiness.MarkReady()

	server.Start(openAPISpec, swaggerUIFS, db, configService, readiness)
}

func secureEndpoint(w http.ResponseWriter, r *http.Request) {
	// This is a secure endpoint, only accessible if Authenticate middleware passes
}
