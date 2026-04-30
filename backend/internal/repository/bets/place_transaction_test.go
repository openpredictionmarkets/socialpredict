package bets

import (
	"context"
	"errors"
	"testing"
	"time"

	dbets "socialpredict/internal/domain/bets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

func TestGormRepositoryPlaceBetTransactionRollsBackDebitWhenBetCreateFails(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}

	user := modelstesting.GenerateUser("bettor", 1000)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	market := modelstesting.GenerateMarket(1, "creator")
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}

	createErr := errors.New("forced bet create failure")
	if err := db.Callback().Create().Before("gorm:create").Register("fail_bet_create", func(tx *gorm.DB) {
		if _, ok := tx.Statement.Dest.(*models.Bet); ok {
			tx.AddError(createErr)
		}
	}); err != nil {
		t.Fatalf("register create callback: %v", err)
	}

	repo := NewGormRepository(db)
	err := repo.PlaceBetTransaction(context.Background(), func(ctx context.Context, txRepo dbets.Repository, users dbets.UserService) error {
		user, hasBet, err := loadPlacementInputs(ctx, txRepo, users, uint(market.ID), "bettor")
		if err != nil {
			return err
		}
		if hasBet {
			t.Fatalf("expected first bet fee path")
		}

		totalCost := int64(100 + 5 + 2)
		if user.AccountBalance-totalCost < -50 {
			return dbets.ErrInsufficientBalance
		}
		if err := users.ApplyTransaction(ctx, "bettor", totalCost, dusers.TransactionBuy); err != nil {
			return err
		}

		return txRepo.Create(ctx, dbets.PlaceRequest{
			Username: "bettor",
			MarketID: uint(market.ID),
			Amount:   100,
			Outcome:  "YES",
		}.NewBet("YES", time.Date(2026, time.April, 29, 22, 0, 0, 0, time.UTC)))
	})
	if !errors.Is(err, createErr) {
		t.Fatalf("expected forced create error, got %v", err)
	}

	var refreshed models.User
	if err := db.Where("username = ?", "bettor").First(&refreshed).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if refreshed.AccountBalance != 1000 {
		t.Fatalf("expected rollback to keep balance 1000, got %d", refreshed.AccountBalance)
	}

	var betCount int64
	if err := db.Model(&models.Bet{}).Where("username = ?", "bettor").Count(&betCount).Error; err != nil {
		t.Fatalf("count bets: %v", err)
	}
	if betCount != 0 {
		t.Fatalf("expected no committed bets, got %d", betCount)
	}
}

func loadPlacementInputs(ctx context.Context, repo dbets.Repository, users dbets.UserService, marketID uint, username string) (*dusers.User, bool, error) {
	user, err := users.GetUser(ctx, username)
	if err != nil {
		return nil, false, err
	}
	hasBet, err := repo.UserHasBet(ctx, marketID, username)
	if err != nil {
		return nil, false, err
	}
	return user, hasBet, nil
}
