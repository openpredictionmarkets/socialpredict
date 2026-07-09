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

func TestGormRepositorySellBetTransactionProjectsSaleInsideTransaction(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}

	seedSellRows(t, db, "creator", "seller", 1000, 202)

	repo := NewGormRepository(db)
	err := repo.SellBetTransaction(context.Background(), func(ctx context.Context, _ dbets.Repository, markets dbets.MarketService, _ dbets.UserService) error {
		current, err := markets.GetUserPositionInMarket(ctx, 202, "seller")
		if err != nil {
			return err
		}
		projected, err := markets.ProjectUserPositionAfterBet(ctx, 202, "seller", boundary.Bet{
			Username: "seller",
			MarketID: 202,
			Amount:   -2,
			Outcome:  "YES",
			PlacedAt: time.Date(2026, time.May, 11, 22, 30, 0, 0, time.UTC),
		})
		if err != nil {
			return err
		}
		if projected.YesSharesOwned >= current.YesSharesOwned {
			t.Fatalf("projected YES shares must decrease after sale: current=%+v projected=%+v", current, projected)
		}
		if projected.NoSharesOwned > current.NoSharesOwned {
			t.Fatalf("projected sale must not create opposite-side shares: current=%+v projected=%+v", current, projected)
		}
		if projected.Value > current.Value {
			t.Fatalf("projected sale must not increase value: current=%+v projected=%+v", current, projected)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("transaction projection returned error: %v", err)
	}
}

func TestGormRepositorySellBetTransactionReadsUnlockedSellablePosition(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}

	seedSellRowsForUnlock(t, db, "creator", "seller", "follower", 1000, 203)

	repo := NewGormRepository(db)
	err := repo.SellBetTransaction(context.Background(), func(ctx context.Context, _ dbets.Repository, markets dbets.MarketService, _ dbets.UserService) error {
		seller, err := markets.GetUserSellablePositionInMarket(ctx, 203, "seller", "NO")
		if err != nil {
			return err
		}
		if seller.NoSharesOwned != 2 || seller.Value != 2 {
			t.Fatalf("expected seller unlocked DBPM-rounded value 2, got %+v", seller)
		}

		follower, err := markets.GetUserSellablePositionInMarket(ctx, 203, "follower", "NO")
		if err != nil {
			return err
		}
		if follower.NoSharesOwned != 0 || follower.Value != 0 {
			t.Fatalf("expected latest follower value to remain locked, got %+v", follower)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("transaction sellable read returned error: %v", err)
	}
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

func seedSellRowsForUnlock(t *testing.T, db *gorm.DB, creatorUsername string, sellerUsername string, followerUsername string, sellerBalance int64, marketID int64) {
	t.Helper()

	creator := modelstesting.GenerateUser(creatorUsername, 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator user: %v", err)
	}
	seller := modelstesting.GenerateUser(sellerUsername, sellerBalance)
	if err := db.Create(&seller).Error; err != nil {
		t.Fatalf("seed seller user: %v", err)
	}
	follower := modelstesting.GenerateUser(followerUsername, 1000)
	if err := db.Create(&follower).Error; err != nil {
		t.Fatalf("seed follower user: %v", err)
	}

	market := modelstesting.GenerateMarket(marketID, creatorUsername)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}

	first := modelstesting.GenerateBet(1, "NO", sellerUsername, uint(marketID), time.Second)
	second := modelstesting.GenerateBet(1, "NO", followerUsername, uint(marketID), 2*time.Second)
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("seed first buy bet: %v", err)
	}
	if err := db.Create(&second).Error; err != nil {
		t.Fatalf("seed second buy bet: %v", err)
	}
}
