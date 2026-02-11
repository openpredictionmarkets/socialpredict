package markets_test

import (
	"context"
	"sort"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	"socialpredict/internal/domain/math/probabilities/wpam"
	rmarkets "socialpredict/internal/repository/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

type fixedClock struct {
	now time.Time
}

func (f fixedClock) Now() time.Time {
	return f.now
}

func setupServiceWithDB(t *testing.T) (*markets.Service, *gorm.DB) {
	t.Helper()

	econ := modelstesting.GenerateEconomicConfig()
	wpam.SetSeeds(wpam.Seeds{
		InitialProbability:     econ.Economics.MarketCreation.InitialMarketProbability,
		InitialSubsidization:   econ.Economics.MarketCreation.InitialMarketSubsidization,
		InitialYesContribution: econ.Economics.MarketCreation.InitialMarketYes,
		InitialNoContribution:  econ.Economics.MarketCreation.InitialMarketNo,
	})

	db := modelstesting.NewFakeDB(t)
	repo := rmarkets.NewGormRepository(db)
	clock := fixedClock{now: time.Now()}
	cfg := markets.Config{}

	service := markets.NewServiceWithWallet(repo, noOpCreatorProfile{}, noOpWallet{}, clock, cfg)
	return service, db
}

func TestServiceListByStatusFiltersMarkets(t *testing.T) {
	service, db := setupServiceWithDB(t)

	now := time.Now()

	user := modelstesting.GenerateUser("testuser", 1000)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	active := models.Market{
		ID:                 1,
		QuestionTitle:      "Active Market",
		Description:        "Active",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(24 * time.Hour),
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    user.Username,
	}

	closed := models.Market{
		ID:                 2,
		QuestionTitle:      "Closed Market",
		Description:        "Closed",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(-24 * time.Hour),
		IsResolved:         false,
		InitialProbability: 0.5,
		CreatorUsername:    user.Username,
	}

	resolved := models.Market{
		ID:                      3,
		QuestionTitle:           "Resolved Market",
		Description:             "Resolved",
		OutcomeType:             "BINARY",
		ResolutionDateTime:      now.Add(-48 * time.Hour),
		FinalResolutionDateTime: now.Add(-24 * time.Hour),
		IsResolved:              true,
		ResolutionResult:        "YES",
		InitialProbability:      0.5,
		CreatorUsername:         user.Username,
	}

	for _, market := range []models.Market{active, closed, resolved} {
		if err := db.Create(&market).Error; err != nil {
			t.Fatalf("create market %s: %v", market.QuestionTitle, err)
		}
	}

	tests := []struct {
		name        string
		status      string
		expectedIDs []int64
	}{
		{
			name:        "Active Markets",
			status:      "active",
			expectedIDs: []int64{1},
		},
		{
			name:        "Closed Markets",
			status:      "closed",
			expectedIDs: []int64{2},
		},
		{
			name:        "Resolved Markets",
			status:      "resolved",
			expectedIDs: []int64{3},
		},
		{
			name:        "All Markets",
			status:      "all",
			expectedIDs: []int64{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := service.ListByStatus(context.Background(), tt.status, markets.Page{Limit: 10})
			if err != nil {
				t.Fatalf("ListByStatus returned error: %v", err)
			}

			var ids []int64
			for _, market := range results {
				ids = append(ids, market.ID)
			}
			sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

			if len(ids) != len(tt.expectedIDs) {
				t.Fatalf("expected %d markets, got %d (ids=%v)", len(tt.expectedIDs), len(ids), ids)
			}

			for i, id := range ids {
				if id != tt.expectedIDs[i] {
					t.Fatalf("expected ids %v, got %v", tt.expectedIDs, ids)
				}
			}
		})
	}
}

func TestServiceListByStatusInvalidStatus(t *testing.T) {
	service, _ := setupServiceWithDB(t)

	_, err := service.ListByStatus(context.Background(), "unknown", markets.Page{})
	if err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}

	if err != markets.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
