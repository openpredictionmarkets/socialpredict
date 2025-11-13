package bets

import (
	"context"
	"testing"
	"time"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestGormRepositoryCreateAndUserHasBet(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})

	repo := NewGormRepository(db)
	ctx := context.Background()

	// Seed required market and user records to satisfy foreign keys.
	user := modelstesting.GenerateUser("bettor", 1000)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	market := modelstesting.GenerateMarket(1, "creator")
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}

	bet := &models.Bet{
		Username: "bettor",
		MarketID: uint(market.ID),
		Amount:   250,
		Outcome:  "YES",
		PlacedAt: time.Now().UTC(),
	}

	if err := repo.Create(ctx, bet); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if bet.ID == 0 {
		t.Fatalf("expected bet ID to be set after Create")
	}

	hasBet, err := repo.UserHasBet(ctx, uint(market.ID), "bettor")
	if err != nil {
		t.Fatalf("UserHasBet returned error: %v", err)
	}
	if !hasBet {
		t.Fatalf("expected bettor to have a bet recorded")
	}

	hasBet, err = repo.UserHasBet(ctx, uint(market.ID), "newuser")
	if err != nil {
		t.Fatalf("UserHasBet (missing user) returned error: %v", err)
	}
	if hasBet {
		t.Fatalf("expected false for user without bets")
	}
}
