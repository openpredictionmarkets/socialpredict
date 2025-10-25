package util

import (
	"os"
	"testing"

	"github.com/brianvoe/gofakeit"
	"socialpredict/models/modelstesting"
)

func TestGenerateUniqueApiKey(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	firstKey := GenerateUniqueApiKey(db)
	if firstKey == "" {
		t.Fatalf("expected non-empty api key")
	}

	user := modelstesting.GenerateUser("alice", 0)
	user.PrivateUser.APIKey = firstKey
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	secondKey := GenerateUniqueApiKey(db)
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

func TestGetEnvLoadsFile(t *testing.T) {
	tempDir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(origWD)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	content := []byte("FOO=bar\n")
	if err := os.WriteFile(".env.dev", content, 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	origVal, had := os.LookupEnv("FOO")
	if had {
		t.Cleanup(func() {
			os.Setenv("FOO", origVal)
		})
	} else {
		t.Cleanup(func() {
			os.Unsetenv("FOO")
		})
	}
	if err := os.Unsetenv("FOO"); err != nil {
		t.Fatalf("unsetenv: %v", err)
	}

	if err := GetEnv(); err != nil {
		t.Fatalf("GetEnv returned error: %v", err)
	}

	if got := os.Getenv("FOO"); got != "bar" {
		t.Fatalf("expected env var to be loaded, got %q", got)
	}
}

func TestGetEnvMissingFile(t *testing.T) {
	tempDir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(origWD)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origVal, had := os.LookupEnv("FOO")
	if had {
		t.Cleanup(func() {
			os.Setenv("FOO", origVal)
		})
	} else {
		t.Cleanup(func() {
			os.Unsetenv("FOO")
		})
	}
	if err := os.Unsetenv("FOO"); err != nil {
		t.Fatalf("unsetenv: %v", err)
	}

	if err := GetEnv(); err != nil {
		t.Fatalf("GetEnv should not fail when file missing: %v", err)
	}

	if got := os.Getenv("FOO"); got != "" {
		t.Fatalf("expected env var to remain empty, got %q", got)
	}
}
