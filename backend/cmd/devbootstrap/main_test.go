package main

import (
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestUpsertBootstrapUserCreatesLoginReadyUser(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	seed := bootstrapUser{
		username:    "testuser01",
		displayName: "Dev Test User 01",
		email:       "testuser01@example.com",
		apiKey:      "dev-testuser01-api-key",
		userType:    "REGULAR",
		emoji:       "NONE",
		description: "Development test user",
	}

	if err := upsertBootstrapUser(db, seed, defaultPassword, 500); err != nil {
		t.Fatalf("upsertBootstrapUser create returned error: %v", err)
	}

	var user models.User
	if err := db.Where("username = ?", seed.username).First(&user).Error; err != nil {
		t.Fatalf("load bootstrapped user: %v", err)
	}
	if user.MustChangePassword {
		t.Fatalf("created bootstrap user must be login-ready with must_change_password=false")
	}
	if !user.CheckPasswordHash(defaultPassword) {
		t.Fatalf("created bootstrap user password should be %q", defaultPassword)
	}
}

func TestUpsertBootstrapUserUpdatesLoginReadyUser(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	seed := bootstrapUser{
		username:    "testuser01",
		displayName: "Dev Test User 01",
		email:       "testuser01@example.com",
		apiKey:      "dev-testuser01-api-key",
		userType:    "REGULAR",
		emoji:       "NONE",
		description: "Development test user",
	}
	existing := modelstesting.GenerateUser(seed.username, 0)
	existing.MustChangePassword = true
	if err := existing.HashPassword("old-password"); err != nil {
		t.Fatalf("hash existing password: %v", err)
	}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("create existing user: %v", err)
	}

	if err := upsertBootstrapUser(db, seed, defaultPassword, 500); err != nil {
		t.Fatalf("upsertBootstrapUser update returned error: %v", err)
	}

	var user models.User
	if err := db.Where("username = ?", seed.username).First(&user).Error; err != nil {
		t.Fatalf("load bootstrapped user: %v", err)
	}
	if user.MustChangePassword {
		t.Fatalf("updated bootstrap user must be login-ready with must_change_password=false")
	}
	if !user.CheckPasswordHash(defaultPassword) {
		t.Fatalf("updated bootstrap user password should be reset to %q", defaultPassword)
	}
}
