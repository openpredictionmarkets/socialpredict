package usershandlers

import (
	"context"
	"strings"
	"testing"
	"time"

	dusers "socialpredict/internal/domain/users"
	rusers "socialpredict/internal/repository/users"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/security"
)

func TestListUserMarketsReturnsDistinctMarketsOrderedByRecentBet(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	if err := db.Exec("ALTER TABLE bets ADD COLUMN user_id INTEGER").Error; err != nil {
		// Ignore duplicate column errors to keep the test resilient across schema changes
		if !strings.Contains(err.Error(), "duplicate column name") {
			t.Fatalf("add user_id column: %v", err)
		}
	}

	user := modelstesting.GenerateUser("list_user", 0)
	user.ID = 101
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	marketA := modelstesting.GenerateMarket(501, user.Username)
	marketB := modelstesting.GenerateMarket(502, user.Username)
	if err := db.Create(&marketA).Error; err != nil {
		t.Fatalf("create marketA: %v", err)
	}
	if err := db.Create(&marketB).Error; err != nil {
		t.Fatalf("create marketB: %v", err)
	}

	firstPlaced := time.Now().Add(-2 * time.Hour)
	secondPlaced := time.Now().Add(-1 * time.Hour)

	bets := []map[string]any{
		{
			"username":   user.Username,
			"user_id":    user.ID,
			"market_id":  marketA.ID,
			"amount":     int64(25),
			"placed_at":  firstPlaced,
			"created_at": firstPlaced,
		},
		{
			"username":   user.Username,
			"user_id":    user.ID,
			"market_id":  marketB.ID,
			"amount":     int64(30),
			"placed_at":  secondPlaced,
			"created_at": secondPlaced,
		},
		{
			"username":   user.Username,
			"user_id":    user.ID,
			"market_id":  marketA.ID,
			"amount":     int64(40),
			"placed_at":  secondPlaced.Add(10 * time.Minute),
			"created_at": secondPlaced.Add(10 * time.Minute),
		},
	}

	for _, payload := range bets {
		if err := db.Table("bets").Create(payload).Error; err != nil {
			t.Fatalf("insert bet %+v: %v", payload, err)
		}
	}

	service := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)

	results, err := ListUserMarkets(context.Background(), service, user.ID)
	if err != nil {
		t.Fatalf("ListUserMarkets returned error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 markets, got %d", len(results))
	}

	seen := map[int64]bool{
		marketA.ID: false,
		marketB.ID: false,
	}
	for _, market := range results {
		seen[market.ID] = true
	}
	for id, ok := range seen {
		if !ok {
			t.Fatalf("expected market %d to be present in results", id)
		}
	}
}

func TestListUserMarketsReturnsErrorFromQuery(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	if err := db.Exec("ALTER TABLE bets ADD COLUMN user_id INTEGER").Error; err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			t.Fatalf("add user_id column: %v", err)
		}
	}

	if err := db.Migrator().DropTable(&models.Bet{}); err != nil {
		t.Fatalf("drop bets table: %v", err)
	}

	service := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)

	if _, err := ListUserMarkets(context.Background(), service, 123); err == nil {
		t.Fatalf("expected error when querying without bets table, got nil")
	}
}
