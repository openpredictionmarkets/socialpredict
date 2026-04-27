package main

import (
	"log"
	"net/http"

	appenv "socialpredict/internal/app/env"
	appruntime "socialpredict/internal/app/runtime"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/migration"
	_ "socialpredict/migration/migrations" // <-- side-effect import: registers migrations via init()
	"socialpredict/seed"
	"socialpredict/server"
	"socialpredict/setup"
)

func main() {
	// Secure endpoint example
	http.Handle("/secure", authsvc.Authenticate(http.HandlerFunc(secureEndpoint)))

	// Load local development env overrides from .env.dev when present.
	if err := appenv.LoadDevFile(); err != nil {
		log.Printf("env: warning loading environment: %v", err)
	}

	dbCfg, err := appruntime.LoadDBConfigFromEnv()
	if err != nil {
		log.Fatalf("db config: %v", err)
	}

	db, err := appruntime.InitDB(dbCfg, appruntime.PostgresFactory{})
	if err != nil {
		log.Fatalf("db init: %v", err)
	}

	const MAX_ATTEMPTS = 20
	if err := seed.EnsureDBReady(db, MAX_ATTEMPTS); err != nil {
		log.Fatalf("database readiness check failed: %v", err)
	}

	if err := migration.MigrateDB(db); err != nil {
		log.Printf("migration: warning: %v", err)
	}

	configService, err := appruntime.LoadConfigService(setup.EmbeddedSource{})
	if err != nil {
		log.Fatalf("config init: %v", err)
	}

	if err := seed.SeedUsers(db, configService); err != nil {
		log.Fatalf("seed users: %v", err)
	}
	if err := seed.SeedHomepage(db, "."); err != nil {
		log.Printf("seed homepage: warning: %v", err)
	}

	server.Start(openAPISpec, swaggerUIFS, db, configService)
}

func secureEndpoint(w http.ResponseWriter, r *http.Request) {
	// This is a secure endpoint, only accessible if Authenticate middleware passes
}
