package repository_test

import (
	"reflect"
	"socialpredict/models"
	"socialpredict/repository"
	"testing"
)

func TestGetAllUsers(t *testing.T) {
	mockUsers := []models.User{
		{ID: 1, Username: "alice", Email: "alice@example.com"},
		{ID: 2, Username: "bob", Email: "bob@example.com"},
	}

	mockDB := &MockDatabase{
		users: mockUsers,
	}

	userRepo := repository.NewUserRepository(mockDB)

	users, err := userRepo.GetAllUsers()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(users, mockUsers) {
		t.Errorf("Expected users %+v, got %+v", mockUsers, users)
	}
}

func TestGetUserByUsername(t *testing.T) {
	mockUsers := []models.User{
		{ID: 1, Username: "alice", Email: "alice@example.com"},
		{ID: 2, Username: "bob", Email: "bob@example.com"},
	}

	mockDB := &MockDatabase{
		users: mockUsers,
	}

	userRepo := repository.NewUserRepository(mockDB)

	user, err := userRepo.GetUserByUsername("alice")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedUser := &models.User{ID: 1, Username: "alice", Email: "alice@example.com"}
	if !reflect.DeepEqual(user, expectedUser) {
		t.Errorf("Expected user %+v, got %+v", expectedUser, user)
	}
}

func TestCountUsers(t *testing.T) {
	mockUsers := []models.User{
		{Username: "user1"},
		{Username: "user2"},
	}

	mockDB := &MockDatabase{
		users: mockUsers,
	}

	userRepo := repository.NewUserRepository(mockDB)
	count, err := userRepo.CountUsers()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if count != int64(len(mockUsers)) {
		t.Errorf("Expected count %d, got %d", len(mockUsers), count)
	}
}
