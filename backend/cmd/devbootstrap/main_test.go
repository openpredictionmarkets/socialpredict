package main

import (
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestUpsertBootstrapUserCreatesLoginReadyUser(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	seed := bootstrapUser{
		username:        "testuser01",
		displayName:     "Dev Test User 01",
		email:           "testuser01@example.com",
		apiKey:          "dev-testuser01-api-key",
		userType:        "MODERATOR",
		moderatorStatus: "active",
		emoji:           "NONE",
		description:     "Development test user",
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
	if user.InitialAccountBalance != 500 || user.AccountBalance != 500 {
		t.Fatalf("created bootstrap user balances = initial %d account %d, want 500/500", user.InitialAccountBalance, user.AccountBalance)
	}
	if user.UserType != "MODERATOR" || user.ModeratorStatus != "active" {
		t.Fatalf("created bootstrap owner should be active moderator, got type=%q status=%q", user.UserType, user.ModeratorStatus)
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
	if user.InitialAccountBalance != 500 || user.AccountBalance != 500 {
		t.Fatalf("updated bootstrap user balances = initial %d account %d, want 500/500", user.InitialAccountBalance, user.AccountBalance)
	}
}

func TestUpsertBootstrapMarketsCreatesPublishedTaggedMarkets(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	owner := bootstrapUser{
		username:        "testuser01",
		displayName:     "Dev Test User 01",
		email:           "testuser01@example.com",
		apiKey:          "dev-testuser01-api-key",
		userType:        "MODERATOR",
		moderatorStatus: "active",
		emoji:           "NONE",
		description:     "Development test user",
	}
	if err := upsertBootstrapUser(db, owner, defaultPassword, 500); err != nil {
		t.Fatalf("upsert owner returned error: %v", err)
	}

	if err := upsertBootstrapMarkets(db, "testuser", 0.42); err != nil {
		t.Fatalf("upsertBootstrapMarkets returned error: %v", err)
	}
	if err := upsertBootstrapMarkets(db, "testuser", 0.42); err != nil {
		t.Fatalf("upsertBootstrapMarkets should be idempotent: %v", err)
	}

	var market models.Market
	if err := db.Where("question_title = ? AND creator_username = ?", "Market A", "testuser01").First(&market).Error; err != nil {
		t.Fatalf("load Market A: %v", err)
	}
	if market.LifecycleStatus != "published" || market.IsResolved {
		t.Fatalf("Market A should be published/unresolved, got lifecycle=%q resolved=%v", market.LifecycleStatus, market.IsResolved)
	}
	if market.InitialProbability != 0.42 {
		t.Fatalf("Market A initial probability = %f, want 0.42", market.InitialProbability)
	}
	if market.StewardUsername != "testuser01" || market.ApprovedBy != "devbootstrap" || market.ApprovedAt == nil {
		t.Fatalf("Market A governance mismatch: steward=%q approvedBy=%q approvedAt=%v", market.StewardUsername, market.ApprovedBy, market.ApprovedAt)
	}
	if !market.ResolutionDateTime.After(market.CreatedAt) {
		t.Fatalf("Market A should have a future resolution date, got %v", market.ResolutionDateTime)
	}

	var tag models.MarketTag
	if err := db.Where("slug = ?", "category-a").First(&tag).Error; err != nil {
		t.Fatalf("load Category A tag: %v", err)
	}
	if tag.DisplayName != "Category A" || !tag.IsActive {
		t.Fatalf("Category A tag mismatch: %+v", tag)
	}

	var assignments int64
	if err := db.Model(&models.MarketTagAssignment{}).Where("market_id = ? AND tag_id = ?", market.ID, tag.ID).Count(&assignments).Error; err != nil {
		t.Fatalf("count tag assignments: %v", err)
	}
	if assignments != 1 {
		t.Fatalf("expected exactly one Market A/Category A assignment, got %d", assignments)
	}
}
