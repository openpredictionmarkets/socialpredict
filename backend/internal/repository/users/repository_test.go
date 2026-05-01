package users

import (
	"context"
	"errors"
	"testing"
	"time"

	dusers "socialpredict/internal/domain/users"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestGormRepositoryGetByUsername(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	user := modelstesting.GenerateUser("alice", 500)
	user.PersonalEmoji = "😀"
	user.Description = "Test user"

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	got, err := repo.GetByUsername(ctx, "alice")
	if err != nil {
		t.Fatalf("GetByUsername returned error: %v", err)
	}
	if got.Username != "alice" || got.AccountBalance != user.AccountBalance || got.PersonalEmoji != "😀" {
		t.Fatalf("unexpected user data: %+v", got)
	}

	if _, err := repo.GetByUsername(ctx, "missing"); !errors.Is(err, dusers.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGormRepositoryUpdateBalance(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	user := modelstesting.GenerateUser("bob", 100)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	if err := repo.UpdateBalance(ctx, "bob", 999); err != nil {
		t.Fatalf("UpdateBalance returned error: %v", err)
	}

	var refreshed models.User
	if err := db.Where("username = ?", "bob").First(&refreshed).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if refreshed.AccountBalance != 999 {
		t.Fatalf("expected balance 999, got %d", refreshed.AccountBalance)
	}

	if err := repo.UpdateBalance(ctx, "ghost", 1); !errors.Is(err, dusers.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound for missing user, got %v", err)
	}
}

func TestGormRepositoryListUserBets(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	user := modelstesting.GenerateUser("carol", 1000)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	market := modelstesting.GenerateMarket(300, "creator")
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}

	earlier := time.Now().Add(-3 * time.Minute)
	later := time.Now().Add(-1 * time.Minute)
	first := models.Bet{
		Username: "carol",
		MarketID: uint(market.ID),
		Amount:   10,
		Outcome:  "YES",
		PlacedAt: later,
	}
	second := models.Bet{
		Username: "carol",
		MarketID: uint(market.ID),
		Amount:   5,
		Outcome:  "NO",
		PlacedAt: earlier,
	}
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("insert first bet: %v", err)
	}
	if err := db.Create(&second).Error; err != nil {
		t.Fatalf("insert second bet: %v", err)
	}

	bets, err := repo.ListUserBets(ctx, "carol")
	if err != nil {
		t.Fatalf("ListUserBets returned error: %v", err)
	}
	if len(bets) != 2 {
		t.Fatalf("expected 2 bets, got %d", len(bets))
	}

	if bets[0].PlacedAt.Before(bets[1].PlacedAt) {
		t.Fatalf("expected bets ordered descending by PlacedAt")
	}
	if bets[0].MarketID != uint(market.ID) {
		t.Fatalf("unexpected market ID in response: %+v", bets[0])
	}
}

func TestGormRepositoryIdentityAndCredentialPersistence(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	user := &dusers.User{
		Username:              "dana",
		DisplayName:           "Dana",
		Email:                 "dana@example.com",
		APIKey:                "api-dana",
		PasswordHash:          "hash-old",
		UserType:              "admin",
		InitialAccountBalance: 500,
		AccountBalance:        500,
		MustChangePassword:    true,
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if user.ID == 0 {
		t.Fatalf("expected Create to assign user ID")
	}

	for _, tt := range []struct {
		name  string
		check func(context.Context, string) (bool, error)
		value string
	}{
		{name: "username", check: repo.UsernameExists, value: user.Username},
		{name: "display name", check: repo.DisplayNameExists, value: user.DisplayName},
		{name: "email", check: repo.EmailExists, value: user.Email},
		{name: "api key", check: repo.APIKeyExists, value: user.APIKey},
	} {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := tt.check(ctx, tt.value)
			if err != nil {
				t.Fatalf("existence check returned error: %v", err)
			}
			if !exists {
				t.Fatalf("expected %s to exist", tt.name)
			}
		})
	}

	exists, err := repo.AnyUserIdentityExists(ctx, user.Username, "other", "other@example.com", "other-key")
	if err != nil {
		t.Fatalf("AnyUserIdentityExists returned error: %v", err)
	}
	if !exists {
		t.Fatalf("expected AnyUserIdentityExists to match username")
	}

	credentials, err := repo.GetCredentials(ctx, user.Username)
	if err != nil {
		t.Fatalf("GetCredentials returned error: %v", err)
	}
	if credentials.PasswordHash != "hash-old" || !credentials.MustChangePassword {
		t.Fatalf("unexpected credentials: %+v", credentials)
	}

	authUser, err := repo.FindAuthenticatedUser(ctx, user.Username)
	if err != nil {
		t.Fatalf("FindAuthenticatedUser returned error: %v", err)
	}
	if authUser.Username != user.Username || authUser.UserType != "admin" || authUser.PasswordHash != "hash-old" || !authUser.MustChangePassword {
		t.Fatalf("unexpected authenticated user: %+v", authUser)
	}

	if err := repo.UpdatePassword(ctx, user.Username, "hash-new", false); err != nil {
		t.Fatalf("UpdatePassword returned error: %v", err)
	}
	updatedCredentials, err := repo.GetCredentials(ctx, user.Username)
	if err != nil {
		t.Fatalf("GetCredentials after password update returned error: %v", err)
	}
	if updatedCredentials.PasswordHash != "hash-new" || updatedCredentials.MustChangePassword {
		t.Fatalf("unexpected updated credentials: %+v", updatedCredentials)
	}

	if _, err := repo.GetCredentials(ctx, "missing"); !errors.Is(err, dusers.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound for missing credentials, got %v", err)
	}
	if _, err := repo.FindAuthenticatedUser(ctx, "missing"); !errors.Is(err, dusers.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound for missing auth user, got %v", err)
	}
	if err := repo.UpdatePassword(ctx, "missing", "hash", false); !errors.Is(err, dusers.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound for missing password update, got %v", err)
	}
}

func TestGormRepositoryUpdateDeleteAndListUsers(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	admin := &dusers.User{
		Username:       "adminuser",
		DisplayName:    "Admin",
		Email:          "admin@example.com",
		APIKey:         "api-admin",
		PasswordHash:   "hash-admin",
		UserType:       "admin",
		AccountBalance: 1000,
	}
	regular := &dusers.User{
		Username:       "regularuser",
		DisplayName:    "Regular",
		Email:          "regular@example.com",
		APIKey:         "api-regular",
		PasswordHash:   "hash-regular",
		UserType:       "user",
		AccountBalance: 500,
	}
	if err := repo.Create(ctx, admin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	if err := repo.Create(ctx, regular); err != nil {
		t.Fatalf("create regular: %v", err)
	}

	regular.DisplayName = "Regular Updated"
	regular.Description = "updated profile"
	regular.AccountBalance = 750
	if err := repo.Update(ctx, regular); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	refreshed, err := repo.GetByUsername(ctx, regular.Username)
	if err != nil {
		t.Fatalf("GetByUsername after update returned error: %v", err)
	}
	if refreshed.DisplayName != "Regular Updated" || refreshed.Description != "updated profile" || refreshed.AccountBalance != 750 {
		t.Fatalf("unexpected refreshed user: %+v", refreshed)
	}

	admins, err := repo.List(ctx, dusers.ListFilters{UserType: "admin", Limit: 10})
	if err != nil {
		t.Fatalf("List admins returned error: %v", err)
	}
	if len(admins) != 1 || admins[0].Username != admin.Username {
		t.Fatalf("unexpected admin list: %+v", admins)
	}

	paged, err := repo.List(ctx, dusers.ListFilters{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("List paged returned error: %v", err)
	}
	if len(paged) != 1 {
		t.Fatalf("expected one paged user, got %d", len(paged))
	}

	if err := repo.Delete(ctx, regular.Username); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, err := repo.GetByUsername(ctx, regular.Username); !errors.Is(err, dusers.ErrUserNotFound) {
		t.Fatalf("expected deleted user to be missing, got %v", err)
	}
	if err := repo.Delete(ctx, regular.Username); !errors.Is(err, dusers.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound for repeated delete, got %v", err)
	}
}
