package main

import (
	"net/http"

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
	defer func() {
		if err := appruntime.CloseDB(db); err != nil {
			logger.Warn("startup", "database shutdown reported a warning", logger.Operation("CloseDB"), logger.Err(err))
		}
	}()

	const MAX_ATTEMPTS = 20
	if err := seed.EnsureDBReady(db, MAX_ATTEMPTS); err != nil {
		logger.Fatal("startup", "database readiness check failed", err, logger.Operation("EnsureDBReady"))
	}

	configService, err := appruntime.LoadConfigService(setup.EmbeddedSource{})
	if err != nil {
		logger.Fatal("startup", "configuration service initialization failed", err, logger.Operation("LoadConfigService"))
	}

	securityConfig, err := appruntime.LoadSecurityConfigFromEnv()
	if err != nil {
		logger.Fatal("startup", "security configuration unavailable", err, logger.Operation("LoadSecurityConfigFromEnv"))
	}
	authsvc.ConfigureJWTSigningKey(securityConfig.JWTSigningKey)

	startupMode, err := appruntime.LoadStartupMutationModeFromEnv()
	if err != nil {
		logger.Fatal("startup", "startup mutation mode unavailable", err, logger.Operation("LoadStartupMutationModeFromEnv"))
	}

	shutdownConfig, err := appruntime.LoadShutdownConfigFromEnv()
	if err != nil {
		logger.Fatal("startup", "shutdown configuration unavailable", err, logger.Operation("LoadShutdownConfigFromEnv"))
	}

	if startupMode.Writer {
		logger.Info("startup", "startup writer enabled for database migrations and seeds", logger.Operation("StartupMutationMode"))
	} else {
		logger.Info("startup", "startup writer disabled; verifying database schema before serving", logger.Operation("StartupMutationMode"))
	}
	if err := runStartupMutations(db, configService, startupMode, startupMutationHooks{
		migrate:      migration.MigrateDB,
		verify:       migration.VerifyApplied,
		seedUsers:    seed.SeedUsers,
		seedHomepage: seed.SeedHomepage,
	}); err != nil {
		logger.Fatal("startup", "startup database mutation check failed", err, logger.Operation("RunStartupMutations"))
	}

	readiness.MarkReady()

	server.Start(openAPISpec, swaggerUIFS, db, configService, readiness, securityConfig, shutdownConfig)
}

func secureEndpoint(w http.ResponseWriter, r *http.Request) {
	// This is a secure endpoint, only accessible if Authenticate middleware passes
}
