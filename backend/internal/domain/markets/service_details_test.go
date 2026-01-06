package markets_test

import (
	"context"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	marketmath "socialpredict/internal/domain/math/market"
	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestServiceGetMarketDetailsCalculatesMetrics(t *testing.T) {
	service, db, calculator := setupServiceWithDB(t)

	creator := modelstesting.GenerateUser("creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(3001, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}
	if err := db.First(&market, market.ID).Error; err != nil {
		t.Fatalf("reload market: %v", err)
	}

	bets := []models.Bet{
		modelstesting.GenerateBet(150, "YES", "alice", uint(market.ID), 0),
		modelstesting.GenerateBet(90, "NO", "bob", uint(market.ID), time.Minute),
		modelstesting.GenerateBet(-40, "YES", "alice", uint(market.ID), 2*time.Minute),
	}
	for i := range bets {
		if err := db.Create(&bets[i]).Error; err != nil {
			t.Fatalf("create bet %d: %v", i, err)
		}
	}

	overview, err := service.GetMarketDetails(context.Background(), market.ID)
	if err != nil {
		t.Fatalf("GetMarketDetails returned error: %v", err)
	}

	expectedVolume := marketmath.GetMarketVolumeWithDust(bets)
	expectedDust := marketmath.GetMarketDust(bets)
	expectedProbabilities := calculator.CalculateMarketProbabilitiesWPAM(market.CreatedAt, bets)

	if overview.TotalVolume != expectedVolume {
		t.Fatalf("total volume = %d, want %d", overview.TotalVolume, expectedVolume)
	}
	if overview.MarketDust != expectedDust {
		t.Fatalf("market dust = %d, want %d", overview.MarketDust, expectedDust)
	}
	if overview.NumUsers != 2 {
		t.Fatalf("num users = %d, want 2", overview.NumUsers)
	}
	if overview.Creator == nil || overview.Creator.Username != market.CreatorUsername {
		t.Fatalf("creator username mismatch: got %+v want %s", overview.Creator, market.CreatorUsername)
	}
	if len(overview.ProbabilityChanges) != len(expectedProbabilities) {
		t.Fatalf("probability history length = %d, want %d", len(overview.ProbabilityChanges), len(expectedProbabilities))
	}
	if overview.LastProbability != expectedProbabilities[len(expectedProbabilities)-1].Probability {
		t.Fatalf("last probability = %f, want %f", overview.LastProbability, expectedProbabilities[len(expectedProbabilities)-1].Probability)
	}
}

func TestServiceGetMarketDetails_InvalidAndMissing(t *testing.T) {
	service, _, _ := setupServiceWithDB(t)

	if _, err := service.GetMarketDetails(context.Background(), 0); err != markets.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if _, err := service.GetMarketDetails(context.Background(), 999); err != markets.ErrMarketNotFound {
		t.Fatalf("expected ErrMarketNotFound, got %v", err)
	}
}
