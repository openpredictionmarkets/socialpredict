package main

import (
	"log"
	"net/http"
	"os"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/server"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	// Secure routes
	http.Handle("/secure", middleware.Authenticate(http.HandlerFunc(secureEndpoint)))

	err := godotenv.Load("./.env.dev")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Use environment variables
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DATABASE")
	dbPort := os.Getenv("POSTGRES_PORT") // Use DB_PORT if connecting from outside the Docker network

	// Database connection settings
	dsn := "host=" + dbHost +
		" user=" + dbUser +
		" password=" + dbPassword +
		" dbname=" + dbName +
		" port=" + dbPort +
		" sslmode=disable TimeZone=UTC"

	// Open the database connection with GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	log.Println("Successfully connected to the database.")

	// Migrate the database
	migrateDB(db)

	// Seed the admin user
	seedAdminUser(db)

	server.Start()
}

func secureEndpoint(w http.ResponseWriter, r *http.Request) {
	// This is a secure endpoint, only accessible if Authenticate middleware passes
}

func migrateDB(db *gorm.DB) {
	// Migrate the schema
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}
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
