package main

import (
	"log"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/migration"
	"socialpredict/server"
	"socialpredict/util"
)

func main() {

	// Secure routes
	http.Handle("/secure", middleware.Authenticate(http.HandlerFunc(secureEndpoint)))

	// Load environment variables
	err := util.GetEnv()
	if err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	// Initialize the database connection
	util.InitDB()

	// Now you can safely use the database connection
	db := util.GetDB()

	// Migrate the database
	migration.MigrateDB(db)

	// Seed the admin user
	// seed.SeedUsers(db)
	// seed.SeedMarket(db)
	// seed.SeedBets(db)

	server.Start()
}

func secureEndpoint(w http.ResponseWriter, r *http.Request) {
	// This is a secure endpoint, only accessible if Authenticate middleware passes
}
