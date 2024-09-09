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
	db.AutoMigrate(&models.User{})

	tx := db.Begin()
	defer func() {
		tx.Rollback()
	}()

	repo := repository.NewUserRepository(tx)

	tx.Create(&models.User{
		Username: "john_doe", DisplayName: "John Doe", Email: "john@example.com",
		UserType: "admin", AccountBalance: 100, ApiKey: "uniqueKey1-John",
	})
	tx.Create(&models.User{
		Username: "jane_doe", DisplayName: "Jane Doe", Email: "jane@example.com",
		UserType: "user", AccountBalance: 150, ApiKey: "uniqueKey2-Jane",
	})

	users, err := repo.GetAllUsers()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	for _, user := range users {
		switch user.Username {
		case "john_doe":
			if user.DisplayName != "John Doe" || user.Email != "john@example.com" || user.UserType != "admin" {
				t.Errorf("User details do not match for %s", user.Username)
			}
		case "jane_doe":
			if user.DisplayName != "Jane Doe" || user.Email != "jane@example.com" || user.UserType != "user" {
				t.Errorf("User details do not match for %s", user.Username)
			}
		default:
			t.Errorf("Unexpected user found: %s", user.Username)
		}
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
