package markets

import (
	"context"
	"errors"
	"testing"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestGormRepositoryCreateAndGetByID(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	market := &dmarkets.Market{
		QuestionTitle:           "Test market",
		Description:             "Description",
		OutcomeType:             "BINARY",
		ResolutionDateTime:      now.Add(24 * time.Hour),
		FinalResolutionDateTime: now.Add(48 * time.Hour),
		ResolutionResult:        "",
		CreatorUsername:         "creator",
		YesLabel:                "YES",
		NoLabel:                 "NO",
		Status:                  "active",
		CreatedAt:               now,
		UpdatedAt:               now,
		InitialProbability:      0.5,
		UTCOffset:               -5,
	}

	if err := repo.Create(ctx, market); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if market.ID == 0 {
		t.Fatalf("expected market ID to be set")
	}

	got, err := repo.GetByID(ctx, market.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if got.QuestionTitle != market.QuestionTitle || got.CreatorUsername != market.CreatorUsername || got.YesLabel != "YES" || got.InitialProbability != 0.5 {
		t.Fatalf("unexpected market data: %+v", got)
	}

	if _, err := repo.GetByID(ctx, market.ID+999); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("expected ErrMarketNotFound, got %v", err)
	}
}

func TestGormRepositoryUpdateLabels(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	seed := modelstesting.GenerateMarket(100, "creator")
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}

	if err := repo.UpdateLabels(ctx, seed.ID, "Moon", "Sun"); err != nil {
		t.Fatalf("UpdateLabels returned error: %v", err)
	}

	var refreshed models.Market
	if err := db.First(&refreshed, seed.ID).Error; err != nil {
		t.Fatalf("reload market: %v", err)
	}
	if refreshed.YesLabel != "Moon" || refreshed.NoLabel != "Sun" {
		t.Fatalf("labels not updated: %+v", refreshed)
	}

	if err := repo.UpdateLabels(ctx, seed.ID+1, "A", "B"); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("expected ErrMarketNotFound for missing market, got %v", err)
	}
}

func TestGormRepositoryListBetsForMarket(t *testing.T) {
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

	market := modelstesting.GenerateMarket(200, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}

	first := models.Bet{
		Username: "bettor",
		MarketID: uint(market.ID),
		Amount:   10,
		Outcome:  "YES",
		PlacedAt: time.Now().Add(-2 * time.Minute),
	}
	second := models.Bet{
		Username: "bettor",
		MarketID: uint(market.ID),
		Amount:   15,
		Outcome:  "NO",
		PlacedAt: time.Now().Add(-1 * time.Minute),
	}
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("insert first bet: %v", err)
	}
	if err := db.Create(&second).Error; err != nil {
		t.Fatalf("insert second bet: %v", err)
	}

	bets, err := repo.ListBetsForMarket(ctx, market.ID)
	if err != nil {
		t.Fatalf("ListBetsForMarket returned error: %v", err)
	}

	if len(bets) != 2 {
		t.Fatalf("expected 2 bets, got %d", len(bets))
	}
	if bets[0].Username != "bettor" || bets[0].Amount != 10 || bets[0].Outcome != "YES" {
		t.Fatalf("unexpected first bet: %+v", bets[0])
	}
	if !bets[0].PlacedAt.Before(bets[1].PlacedAt) {
		t.Fatalf("bets not ordered ascending by PlacedAt")
	}
}
