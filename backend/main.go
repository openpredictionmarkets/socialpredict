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

	err := util.GetEnv()
	if err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	util.InitDB()

	db := util.GetDB()

	if err := seed.EnsureDBReady(db, 20); err != nil {
		log.Fatalf("Database readiness check failed: %v", err)
	}

	migration.MigrateDB(db)

	seed.SeedUsers(db)

	server.Start()
}

func secureEndpoint(w http.ResponseWriter, r *http.Request) {
	// This is a secure endpoint, only accessible if Authenticate middleware passes
}
