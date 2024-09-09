package repository_test

import (
	"socialpredict/models"
	"socialpredict/repository"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetAllUsers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	db.AutoMigrate(&models.User{}) // Setup database

	repo := repository.NewUserRepository(db)

	// Seed the database
	db.Create(&models.User{Username: "john_doe", DisplayName: "John Doe"})
	db.Create(&models.User{Username: "jane_doe", DisplayName: "Jane Doe"})

	// Test GetAllUsers
	users, err := repo.GetAllUsers()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestGetUserByUsername(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	db.AutoMigrate(&models.User{}) // Setup database

	repo := repository.NewUserRepository(db)

	// Seed the database
	user := models.User{Username: "john_doe", DisplayName: "John Doe"}
	db.Create(&user)

	// Test GetUserByUsername
	result, err := repo.GetUserByUsername("john_doe")
	if err != nil {
		t.Fatalf("Failed to get user by username: %v", err)
	}
	if result.Username != "john_doe" {
		t.Errorf("Expected username %v, got %v", "john_doe", result.Username)
	}
}
