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
