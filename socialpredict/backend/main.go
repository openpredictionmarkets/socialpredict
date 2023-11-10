package main

import (
	"net/http"
	"socialpredict/handlers"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/server"
)

func main() {
	http.HandleFunc("/register", handlers.Register)
	http.HandleFunc("/login", handlers.Login)

	// Secure routes
	http.Handle("/secure", middleware.Authenticate(http.HandlerFunc(secureEndpoint)))

	// Initialize the database connection
	db := // ... database initialization ...

		// Seed the admin user
		seedAdminUser(db)

	server.Start()
}

func secureEndpoint(w http.ResponseWriter, r *http.Request) {
	// This is a secure endpoint, only accessible if Authenticate middleware passes
}

func seedAdminUser(db *gorm.DB) {
	// Check if the admin user already exists
	var count int64
	db.Model(&models.User{}).Where("username = ?", "admin").Count(&count)
	if count == 0 {
		// No admin user found, create one
		adminUser := models.User{
			Username: "admin",
			Email:    "admin@example.com",
		}
		adminUser.HashPassword("securepassword") // Always use a strong, hashed password

		db.Create(&adminUser)
	}
}
