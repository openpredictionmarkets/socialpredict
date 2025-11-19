package markets_test

import (
	"context"
	"sort"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

func seedSearchMarkets(t *testing.T, db *gorm.DB, username string) {
	t.Helper()

	now := time.Now()

	markets := []models.Market{
		{
			ID:                 1,
			QuestionTitle:      "Will Bitcoin reach $100k by end of year?",
			Description:        "Bitcoin price prediction",
			OutcomeType:        "BINARY",
			ResolutionDateTime: now.Add(48 * time.Hour),
			IsResolved:         false,
			InitialProbability: 0.5,
			CreatorUsername:    username,
		},
		{
			ID:                 2,
			QuestionTitle:      "Bitcoin market prediction",
			Description:        "Closed bitcoin market",
			OutcomeType:        "BINARY",
			ResolutionDateTime: now.Add(-1 * time.Hour),
			IsResolved:         false,
			InitialProbability: 0.5,
			CreatorUsername:    username,
		},
		{
			ID:                      3,
			QuestionTitle:           "Will Bitcoin overtake gold market cap?",
			Description:             "Resolved bitcoin market",
			OutcomeType:             "BINARY",
			ResolutionDateTime:      now.Add(-24 * time.Hour),
			FinalResolutionDateTime: now.Add(-12 * time.Hour),
			IsResolved:              true,
			ResolutionResult:        "YES",
			InitialProbability:      0.5,
			CreatorUsername:         username,
		},
		{
			ID:                 4,
			QuestionTitle:      "Stock market crash prediction",
			Description:        "Market about stocks",
			OutcomeType:        "BINARY",
			ResolutionDateTime: now.Add(24 * time.Hour),
			IsResolved:         false,
			InitialProbability: 0.5,
			CreatorUsername:    username,
		},
	}

	for _, market := range markets {
		if err := db.Create(&market).Error; err != nil {
			t.Fatalf("seed market %d: %v", market.ID, err)
		}
	}
}

func TestServiceSearchMarketsFiltersByStatus(t *testing.T) {
	service, db := setupServiceWithDB(t)

	user := modelstesting.GenerateUser("testuser", 0)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	seedSearchMarkets(t, db, user.Username)

	tests := []struct {
		name             string
		status           string
		expectedIDs      []int64
		expectedTotal    int
		expectedFallback bool
	}{
		{
			name:          "Keyword only",
			status:        "",
			expectedIDs:   []int64{1, 2, 3},
			expectedTotal: 3,
		},
		{
			name:             "Active only with fallback",
			status:           "active",
			expectedIDs:      []int64{1},
			expectedTotal:    3,
			expectedFallback: true,
		},
		{
			name:             "Closed only with fallback",
			status:           "closed",
			expectedIDs:      []int64{2},
			expectedTotal:    3,
			expectedFallback: true,
		},
		{
			name:             "Resolved only with fallback",
			status:           "resolved",
			expectedIDs:      []int64{3},
			expectedTotal:    3,
			expectedFallback: true,
		},
		{
			name:          "All statuses",
			status:        "all",
			expectedIDs:   []int64{1, 2, 3},
			expectedTotal: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := markets.SearchFilters{Status: tt.status, Limit: 10}
			result, err := service.SearchMarkets(context.Background(), "bitcoin", filters)
			if err != nil {
				t.Fatalf("SearchMarkets error: %v", err)
			}

			if result.TotalCount != tt.expectedTotal {
				t.Fatalf("expected total %d, got %d", tt.expectedTotal, result.TotalCount)
			}

			var primaryIDs []int64
			for _, market := range result.PrimaryResults {
				primaryIDs = append(primaryIDs, market.ID)
			}
			sort.Slice(primaryIDs, func(i, j int) bool { return primaryIDs[i] < primaryIDs[j] })

			if len(primaryIDs) != len(tt.expectedIDs) {
				t.Fatalf("expected primary ids %v, got %v", tt.expectedIDs, primaryIDs)
			}

			for i, id := range primaryIDs {
				if id != tt.expectedIDs[i] {
					t.Fatalf("expected primary ids %v, got %v", tt.expectedIDs, primaryIDs)
				}
			}

			if tt.expectedFallback && !result.FallbackUsed {
				t.Fatalf("expected fallback to be used")
			}

			if tt.expectedFallback && result.FallbackCount == 0 {
				t.Fatalf("expected fallback results, got none")
			}
		})
	}
}

func TestServiceSearchMarketsInvalidInput(t *testing.T) {
	service, _ := setupServiceWithDB(t)

	_, err := service.SearchMarkets(context.Background(), "   ", markets.SearchFilters{})
	if err == nil {
		t.Fatal("expected error for empty query")
	}

	if err != markets.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
