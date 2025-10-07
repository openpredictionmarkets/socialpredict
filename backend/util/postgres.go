package util

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var err error

// InitDB initializes the database connection.
// It supports both canonical POSTGRES_* variables and legacy DB_* fallbacks.
func InitDB() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = os.Getenv("DBHOST")
	}

	dbUser := os.Getenv("POSTGRES_USER")
  if dbUser == "" {
		dbUser = os.Getenv("DB_USER")
	}

	dbPassword := os.Getenv("POSTGRES_PASSWORD")
  if dbPassword == "" {
		dbPassword = os.Getenv("DB_PASS")
	}
	if dbPassword == "" {
		dbPassword = os.Getenv("DB_PASSWORD")
	}

	dbName := os.Getenv("POSTGRES_DATABASE")
	if dbName == "" {
		dbName = os.Getenv("POSTGRES_DB")
	}
	if dbName == "" {
		dbName = os.Getenv("DB_NAME")
	}

	dbPort := os.Getenv("POSTGRES_PORT")
  if dbPort == "" {
		dbPort = os.Getenv("DB_PORT")
	}
	if dbPort == "" {
		dbPort = "5432"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	log.Println("Successfully connected to the database.")
}

// GetDB returns the database connection
func GetDB() *gorm.DB {
	return DB
}
