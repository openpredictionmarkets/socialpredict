package bets

import (
	"context"
	"errors"
	"testing"
	"time"

	dbets "socialpredict/internal/domain/bets"
	"socialpredict/internal/domain/boundary"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

func TestGormRepositorySellBetTransactionRollsBackCreditWhenSaleBetCreateFails(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}

	seedSellRows(t, db, "creator", "seller", 1000, 201)

	createErr := errors.New("forced sale bet create failure")
	if err := db.Callback().Create().Before("gorm:create").Register("fail_sale_bet_create", func(tx *gorm.DB) {
		if _, ok := tx.Statement.Dest.(*models.Bet); ok {
			tx.AddError(createErr)
		}
	}); err != nil {
		t.Fatalf("register create callback: %v", err)
	}

	repo := NewGormRepository(db)
	err := repo.SellBetTransaction(context.Background(), func(ctx context.Context, txRepo dbets.Repository, markets dbets.MarketService, users dbets.UserService) error {
		if _, err := markets.GetMarket(ctx, 201); err != nil {
			return err
		}
		if _, err := markets.GetUserPositionInMarket(ctx, 201, "seller"); err != nil {
			return err
		}
		if err := users.ApplyTransaction(ctx, "seller", 25, dusers.TransactionSale); err != nil {
			return err
		}
		return txRepo.Create(ctx, &boundary.Bet{
			Username: "seller",
			MarketID: 201,
			Amount:   -2,
			Outcome:  "YES",
			PlacedAt: time.Date(2026, time.May, 11, 22, 30, 0, 0, time.UTC),
		})
	})
	if !errors.Is(err, createErr) {
		t.Fatalf("expected forced create error, got %v", err)
	}

	assertUserBalance(t, db, "seller", 1000)
	assertBetCount(t, db, "seller", 201, 1)
}

func seedSellRows(t *testing.T, db *gorm.DB, creatorUsername string, sellerUsername string, sellerBalance int64, marketID int64) {
	t.Helper()

	creator := modelstesting.GenerateUser(creatorUsername, 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator user: %v", err)
	}

	seller := modelstesting.GenerateUser(sellerUsername, sellerBalance)
	if err := db.Create(&seller).Error; err != nil {
		t.Fatalf("seed seller user: %v", err)
	}

	market := modelstesting.GenerateMarket(marketID, creatorUsername)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}

	buy := modelstesting.GenerateBet(100, "YES", sellerUsername, uint(marketID), 0)
	if err := db.Create(&buy).Error; err != nil {
		t.Fatalf("seed buy bet: %v", err)
	}
}
