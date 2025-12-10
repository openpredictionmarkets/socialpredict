package main

import (
	"log"
	"net/http"

	authsvc "socialpredict/internal/service/auth"
	"socialpredict/migration"
	_ "socialpredict/migration/migrations" // <-- side-effect import: registers migrations via init()
	"socialpredict/seed"
	"socialpredict/server"
	"socialpredict/util"
)

func main() {
	// Secure endpoint example
	http.Handle("/secure", authsvc.Authenticate(http.HandlerFunc(secureEndpoint)))

	// Load env (.env, .env.dev)
	if err := util.GetEnv(); err != nil {
		log.Printf("env: warning loading environment: %v", err)
	}

	dbCfg, err := util.LoadDBConfigFromEnv()
	if err != nil {
		log.Fatalf("db config: %v", err)
	}

	db, err := util.InitDB(dbCfg, util.PostgresFactory{})
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

	seed.SeedUsers(db)
	if err := seed.SeedHomepage(db, "."); err != nil {
		log.Printf("seed homepage: warning: %v", err)
	}

	server.Start(openAPISpec, swaggerUIFS)
}

func secureEndpoint(w http.ResponseWriter, r *http.Request) {
	// This is a secure endpoint, only accessible if Authenticate middleware passes
}
