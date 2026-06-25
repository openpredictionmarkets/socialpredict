package analytics

import (
	"context"
	"testing"
	"time"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestGormRepositoryListBetsForMarketScopesAndOrdersReplayRows(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("creator", 1000)
	bettor := modelstesting.GenerateUser("bettor", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}
	if err := db.Create(&bettor).Error; err != nil {
		t.Fatalf("seed bettor: %v", err)
	}

	market := modelstesting.GenerateMarket(1001, creator.Username)
	otherMarket := modelstesting.GenerateMarket(1002, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}
	if err := db.Create(&otherMarket).Error; err != nil {
		t.Fatalf("seed unrelated market: %v", err)
	}

	placedAt := time.Now().UTC().Truncate(time.Second)
	first := models.Bet{
		Username: "bettor",
		MarketID: uint(market.ID),
		Amount:   7,
		Outcome:  "YES",
		PlacedAt: placedAt,
	}
	unrelated := models.Bet{
		Username: "bettor",
		MarketID: uint(otherMarket.ID),
		Amount:   999,
		Outcome:  "NO",
		PlacedAt: placedAt,
	}
	second := models.Bet{
		Username: "bettor",
		MarketID: uint(market.ID),
		Amount:   9,
		Outcome:  "NO",
		PlacedAt: placedAt,
	}
	for _, bet := range []*models.Bet{&first, &unrelated, &second} {
		if err := db.Create(bet).Error; err != nil {
			t.Fatalf("seed bet: %v", err)
		}
	}

	bets, err := repo.ListBetsForMarket(ctx, uint(market.ID))
	if err != nil {
		t.Fatalf("ListBetsForMarket returned error: %v", err)
	}

	if len(bets) != 2 {
		t.Fatalf("expected 2 scoped bets, got %d: %+v", len(bets), bets)
	}
	if bets[0].MarketID != uint(market.ID) || bets[1].MarketID != uint(market.ID) {
		t.Fatalf("expected only market %d rows, got %+v", market.ID, bets)
	}
	if bets[0].Amount != 7 || bets[1].Amount != 9 {
		t.Fatalf("expected target market rows in insertion-id tie order, got %+v", bets)
	}
	if bets[0].ID >= bets[1].ID {
		t.Fatalf("expected stable id tie-break ordering, got first id %d second id %d", bets[0].ID, bets[1].ID)
	}
}
