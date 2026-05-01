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

func TestGormRepositoryListHonorsStatusAndCreatorFilter(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	alice := modelstesting.GenerateUser("alice", 1000)
	bob := modelstesting.GenerateUser("bob", 1000)
	if err := db.Create(&alice).Error; err != nil {
		t.Fatalf("seed alice: %v", err)
	}
	if err := db.Create(&bob).Error; err != nil {
		t.Fatalf("seed bob: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	active := modelstesting.GenerateMarket(300, alice.Username)
	active.ResolutionDateTime = now.Add(24 * time.Hour)
	active.IsResolved = false

	closed := modelstesting.GenerateMarket(301, alice.Username)
	closed.ResolutionDateTime = now.Add(-24 * time.Hour)
	closed.IsResolved = false

	resolved := modelstesting.GenerateMarket(302, alice.Username)
	resolved.ResolutionDateTime = now.Add(-48 * time.Hour)
	resolved.FinalResolutionDateTime = now.Add(-12 * time.Hour)
	resolved.IsResolved = true
	resolved.ResolutionResult = "YES"

	bobsClosed := modelstesting.GenerateMarket(303, bob.Username)
	bobsClosed.ResolutionDateTime = now.Add(-24 * time.Hour)
	bobsClosed.IsResolved = false

	for _, market := range []any{&active, &closed, &resolved, &bobsClosed} {
		if err := db.Create(market).Error; err != nil {
			t.Fatalf("seed market: %v", err)
		}
	}

	tests := []struct {
		name       string
		filters    dmarkets.ListFilters
		wantID     int64
		wantStatus string
	}{
		{
			name:       "active for alice",
			filters:    dmarkets.ListFilters{Status: "active", CreatedBy: alice.Username, Limit: 10},
			wantID:     active.ID,
			wantStatus: "active",
		},
		{
			name:       "closed for alice",
			filters:    dmarkets.ListFilters{Status: "closed", CreatedBy: alice.Username, Limit: 10},
			wantID:     closed.ID,
			wantStatus: "closed",
		},
		{
			name:       "resolved for alice",
			filters:    dmarkets.ListFilters{Status: "resolved", CreatedBy: alice.Username, Limit: 10},
			wantID:     resolved.ID,
			wantStatus: "resolved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markets, err := repo.List(ctx, tt.filters)
			if err != nil {
				t.Fatalf("List returned error: %v", err)
			}
			if len(markets) != 1 {
				t.Fatalf("expected 1 market, got %d", len(markets))
			}
			if markets[0].ID != tt.wantID {
				t.Fatalf("expected market ID %d, got %d", tt.wantID, markets[0].ID)
			}
			if markets[0].Status != tt.wantStatus {
				t.Fatalf("expected status %q, got %q", tt.wantStatus, markets[0].Status)
			}
		})
	}
}

func TestGormRepositoryPublicSearchStatusResolveAndDelete(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("creator", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	active := modelstesting.GenerateMarket(400, creator.Username)
	active.QuestionTitle = "Will oranges rally?"
	active.Description = "Citrus market"
	active.ResolutionDateTime = now.Add(24 * time.Hour)
	active.IsResolved = false

	closed := modelstesting.GenerateMarket(401, creator.Username)
	closed.QuestionTitle = "Will apples close?"
	closed.Description = "Orchard market"
	closed.ResolutionDateTime = now.Add(-24 * time.Hour)
	closed.IsResolved = false

	resolved := modelstesting.GenerateMarket(402, creator.Username)
	resolved.QuestionTitle = "Will pears resolve?"
	resolved.Description = "Resolved orchard market"
	resolved.ResolutionDateTime = now.Add(-48 * time.Hour)
	resolved.IsResolved = true
	resolved.ResolutionResult = "NO"

	for _, market := range []*models.Market{&active, &closed, &resolved} {
		if err := db.Create(market).Error; err != nil {
			t.Fatalf("seed market %d: %v", market.ID, err)
		}
	}

	publicMarket, err := repo.GetPublicMarket(ctx, active.ID)
	if err != nil {
		t.Fatalf("GetPublicMarket returned error: %v", err)
	}
	if publicMarket.ID != active.ID || publicMarket.QuestionTitle != active.QuestionTitle || publicMarket.CreatorUsername != creator.Username {
		t.Fatalf("unexpected public market: %+v", publicMarket)
	}
	if _, err := repo.GetPublicMarket(ctx, 9999); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("expected ErrMarketNotFound for missing public market, got %v", err)
	}

	searchResults, err := repo.Search(ctx, "orchard", dmarkets.SearchFilters{Status: "closed", Limit: 10})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(searchResults) != 1 || searchResults[0].ID != closed.ID {
		t.Fatalf("unexpected closed search results: %+v", searchResults)
	}

	resolvedResults, err := repo.ListByStatus(ctx, "resolved", dmarkets.Page{Limit: 10})
	if err != nil {
		t.Fatalf("ListByStatus returned error: %v", err)
	}
	if len(resolvedResults) != 1 || resolvedResults[0].ID != resolved.ID || resolvedResults[0].Status != "resolved" {
		t.Fatalf("unexpected resolved results: %+v", resolvedResults)
	}

	if err := repo.ResolveMarket(ctx, active.ID, "YES"); err != nil {
		t.Fatalf("ResolveMarket returned error: %v", err)
	}
	refreshed, err := repo.GetByID(ctx, active.ID)
	if err != nil {
		t.Fatalf("GetByID after resolve returned error: %v", err)
	}
	if refreshed.Status != "resolved" || refreshed.ResolutionResult != "YES" {
		t.Fatalf("unexpected resolved market: %+v", refreshed)
	}
	if err := repo.ResolveMarket(ctx, 9999, "YES"); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("expected ErrMarketNotFound for missing resolve, got %v", err)
	}

	if err := repo.Delete(ctx, closed.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, err := repo.GetByID(ctx, closed.ID); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("expected deleted market to be missing, got %v", err)
	}
	if err := repo.Delete(ctx, closed.ID); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("expected ErrMarketNotFound for repeated delete, got %v", err)
	}
}
