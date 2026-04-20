package users

import (
	"testing"

	"github.com/brianvoe/gofakeit"

	"socialpredict/models/modelstesting"
)

func TestGenerateUniqueAPIKey(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	firstKey := GenerateUniqueAPIKey(db)
	if firstKey == "" {
		t.Fatalf("expected non-empty api key")
	}

	user := modelstesting.GenerateUser("alice", 0)
	user.PrivateUser.APIKey = firstKey
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	secondKey := GenerateUniqueAPIKey(db)
	if secondKey == "" || secondKey == firstKey {
		t.Fatalf("expected a distinct api key, got %q", secondKey)
	}
}

func TestCheckUserIsReal(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	user := modelstesting.GenerateUser("bob", 0)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	if err := CheckUserIsReal(db, "bob"); err != nil {
		t.Fatalf("expected user to be real: %v", err)
	}

	if err := CheckUserIsReal(db, "missing"); err == nil {
		t.Fatalf("expected error for missing user")
	}
}

func TestCountByField(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	user := modelstesting.GenerateUser("charlie", 0)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	if count := CountByField(db, "username", user.Username); count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func TestUniqueDisplayNameAndEmail(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	existing := modelstesting.GenerateUser("dana", 0)
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	gofakeit.Seed(1)

	displayName := UniqueDisplayName(db)
	if displayName == "" {
		t.Fatalf("expected non-empty display name")
	}
	if CountByField(db, "display_name", displayName) != 0 {
		t.Fatalf("expected generated display name to be unique")
	}

	email := UniqueEmail(db)
	if email == "" {
		t.Fatalf("expected non-empty email")
	}
	if CountByField(db, "email", email) != 0 {
		t.Fatalf("expected generated email to be unique")
	}
}
