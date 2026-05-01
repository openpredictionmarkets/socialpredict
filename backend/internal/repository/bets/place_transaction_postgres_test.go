package bets

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	dbets "socialpredict/internal/domain/bets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGormRepositoryPlaceBetTransactionPostgres(t *testing.T) {
	dsn, ok := postgresIntegrationDSN()
	if !ok {
		t.Skip("set SOCIALPREDICT_POSTGRES_TEST_DSN or POSTGRES_TEST_DSN to run real-Postgres place-bet transaction verification")
	}

	t.Run("commitsDebitAndBetTogether", func(t *testing.T) {
		db := openPlaceBetPostgresDB(t, dsn)
		seedPlaceBetRows(t, db, "creator", "bettor", 1000, 101)

		repo := NewGormRepository(db)
		if err := applyPlacement(context.Background(), repo, 101, "bettor", 100, 7); err != nil {
			t.Fatalf("PlaceBetTransaction returned error: %v", err)
		}

		assertUserBalance(t, db, "bettor", 893)
		assertBetCount(t, db, "bettor", 101, 1)
	})

	t.Run("rollsBackDebitWhenBetCreateFails", func(t *testing.T) {
		db := openPlaceBetPostgresDB(t, dsn)
		seedPlaceBetRows(t, db, "creator", "bettor", 1000, 102)

		createErr := errors.New("forced postgres bet create failure")
		if err := db.Callback().Create().Before("gorm:create").Register("fail_postgres_bet_create", func(tx *gorm.DB) {
			if _, ok := tx.Statement.Dest.(*models.Bet); ok {
				tx.AddError(createErr)
			}
		}); err != nil {
			t.Fatalf("register create callback: %v", err)
		}

		repo := NewGormRepository(db)
		err := applyPlacement(context.Background(), repo, 102, "bettor", 100, 7)
		if !errors.Is(err, createErr) {
			t.Fatalf("expected forced create error, got %v", err)
		}

		assertUserBalance(t, db, "bettor", 1000)
		assertBetCount(t, db, "bettor", 102, 0)
	})

	t.Run("serializesOverlappingDebits", func(t *testing.T) {
		db := openPlaceBetPostgresDB(t, dsn)
		seedPlaceBetRows(t, db, "creator", "bettor", 150, 103)

		repo := NewGormRepository(db)
		firstLocked := make(chan struct{})
		releaseFirst := make(chan struct{})
		results := make(chan error, 2)

		go func() {
			results <- applyPlacementWithHook(context.Background(), repo, 103, "bettor", 100, 0, func() {
				close(firstLocked)
				<-releaseFirst
			})
		}()

		select {
		case <-firstLocked:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for first transaction to lock user row")
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- applyPlacement(context.Background(), repo, 103, "bettor", 100, 0)
		}()

		time.Sleep(100 * time.Millisecond)
		close(releaseFirst)
		wg.Wait()

		firstErr := <-results
		secondErr := <-results
		if firstErr != nil && secondErr != nil {
			t.Fatalf("expected one placement to commit, got errors %v and %v", firstErr, secondErr)
		}
		if firstErr == nil && secondErr == nil {
			t.Fatalf("expected one overlapping placement to fail after serialized balance check")
		}
		if firstErr != nil && !errors.Is(firstErr, dbets.ErrInsufficientBalance) {
			t.Fatalf("unexpected first placement error: %v", firstErr)
		}
		if secondErr != nil && !errors.Is(secondErr, dbets.ErrInsufficientBalance) {
			t.Fatalf("unexpected second placement error: %v", secondErr)
		}

		assertUserBalance(t, db, "bettor", 50)
		assertBetCount(t, db, "bettor", 103, 1)
	})
}

func applyPlacement(ctx context.Context, repo *GormRepository, marketID uint, username string, amount int64, fees int64) error {
	return applyPlacementWithHook(ctx, repo, marketID, username, amount, fees, nil)
}

func applyPlacementWithHook(ctx context.Context, repo *GormRepository, marketID uint, username string, amount int64, fees int64, afterLoad func()) error {
	return repo.PlaceBetTransaction(ctx, func(ctx context.Context, txRepo dbets.Repository, users dbets.UserService) error {
		user, hasBet, err := loadPlacementInputs(ctx, txRepo, users, marketID, username)
		if err != nil {
			return err
		}
		if afterLoad != nil {
			afterLoad()
		}
		if hasBet {
			fees = 0
		}

		totalCost := amount + fees
		if user.AccountBalance-totalCost < 0 {
			return dbets.ErrInsufficientBalance
		}
		if err := users.ApplyTransaction(ctx, username, totalCost, dusers.TransactionBuy); err != nil {
			return err
		}

		return txRepo.Create(ctx, dbets.PlaceRequest{
			Username: username,
			MarketID: marketID,
			Amount:   amount,
			Outcome:  "YES",
		}.NewBet("YES", time.Date(2026, time.April, 29, 22, 0, 0, 0, time.UTC)))
	})
}

func seedPlaceBetRows(t *testing.T, db *gorm.DB, creatorUsername string, bettorUsername string, bettorBalance int64, marketID int64) {
	t.Helper()

	creator := modelstesting.GenerateUser(creatorUsername, 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator user: %v", err)
	}

	bettor := modelstesting.GenerateUser(bettorUsername, bettorBalance)
	if err := db.Create(&bettor).Error; err != nil {
		t.Fatalf("seed bettor user: %v", err)
	}

	market := modelstesting.GenerateMarket(marketID, creatorUsername)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}
}

func openPlaceBetPostgresDB(t *testing.T, dsn string) *gorm.DB {
	t.Helper()

	adminDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres test database: %v", err)
	}
	adminSQL, err := adminDB.DB()
	if err != nil {
		t.Fatalf("postgres admin sql handle: %v", err)
	}
	t.Cleanup(func() { _ = adminSQL.Close() })

	schema := fmt.Sprintf("sp_place_bet_%d", time.Now().UnixNano())
	if err := adminDB.Exec(`CREATE SCHEMA ` + quotePostgresIdentifier(schema)).Error; err != nil {
		t.Fatalf("create isolated postgres schema: %v", err)
	}
	t.Cleanup(func() {
		_ = adminDB.Exec(`DROP SCHEMA IF EXISTS ` + quotePostgresIdentifier(schema) + ` CASCADE`).Error
	})

	db, err := gorm.Open(postgres.Open(postgresDSNWithSearchPath(dsn, schema)), &gorm.Config{})
	if err != nil {
		t.Fatalf("open isolated postgres schema: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("postgres sql handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(4)
	t.Cleanup(func() { _ = sqlDB.Close() })

	if db.Dialector.Name() != "postgres" {
		t.Fatalf("expected postgres dialector, got %q", db.Dialector.Name())
	}
	if err := db.AutoMigrate(&models.User{}, &models.Market{}, &models.Bet{}); err != nil {
		t.Fatalf("migrate place-bet tables: %v", err)
	}

	return db
}

func postgresIntegrationDSN() (string, bool) {
	for _, key := range []string{"SOCIALPREDICT_POSTGRES_TEST_DSN", "POSTGRES_TEST_DSN"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value, true
		}
	}
	return "", false
}

func postgresDSNWithSearchPath(dsn string, schema string) string {
	if parsed, err := url.Parse(dsn); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}

	return strings.TrimSpace(dsn) + " search_path=" + schema
}

func quotePostgresIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func assertUserBalance(t *testing.T, db *gorm.DB, username string, want int64) {
	t.Helper()

	var refreshed models.User
	if err := db.Where("username = ?", username).First(&refreshed).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if refreshed.AccountBalance != want {
		t.Fatalf("expected balance %d, got %d", want, refreshed.AccountBalance)
	}
}

func assertBetCount(t *testing.T, db *gorm.DB, username string, marketID uint, want int64) {
	t.Helper()

	var betCount int64
	if err := db.Model(&models.Bet{}).Where("username = ? AND market_id = ?", username, marketID).Count(&betCount).Error; err != nil {
		t.Fatalf("count bets: %v", err)
	}
	if betCount != want {
		t.Fatalf("expected %d committed bets, got %d", want, betCount)
	}
}
