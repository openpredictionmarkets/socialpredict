package main

import (
	"log"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/migration"
	"socialpredict/seed"
	"socialpredict/server"
	"socialpredict/util"
)

func main() {

	http.Handle("/secure", middleware.Authenticate(http.HandlerFunc(secureEndpoint)))

	// Load .env.dev if present; non-fatal if missing
	if err := util.GetEnv(); err != nil {
		// util.GetEnv is tolerant, but log any unexpected errors
		log.Printf("Warning loading environment: %v", err)
	}

	util.InitDB()

	db := util.GetDB()

	if err := seed.EnsureDBReady(db, 20); err != nil {
		log.Fatalf("Database readiness check failed: %v", err)
	}

	migration.MigrateDB(db)

	seed.SeedUsers(db)

	// Seed homepage content - pass current working directory as repo root
	if err := seed.SeedHomepage(db, "."); err != nil {
		log.Printf("Warning: Failed to seed homepage content: %v", err)
	}

	server.Start()
}

func secureEndpoint(w http.ResponseWriter, r *http.Request) {
	// This is a secure endpoint, only accessible if Authenticate middleware passes
}
