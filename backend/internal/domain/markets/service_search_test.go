package markets_test

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

func seedSearchMarkets(t *testing.T, db *gorm.DB, username string) {
	t.Helper()

	now := time.Now()
	marketsToSeed := []any{
		mustSearchMarket(1, username, now, "Will Bitcoin reach $100k by end of year?", "Bitcoin price prediction", now.Add(48*time.Hour), time.Time{}, false, ""),
		mustSearchMarket(2, username, now, "Bitcoin market prediction", "Closed bitcoin market", now.Add(-1*time.Hour), time.Time{}, false, ""),
		mustSearchMarket(3, username, now, "Will Bitcoin overtake gold market cap?", "Resolved bitcoin market", now.Add(-24*time.Hour), now.Add(-12*time.Hour), true, "YES"),
		mustSearchMarket(4, username, now, "Stock market crash prediction", "Market about stocks", now.Add(24*time.Hour), time.Time{}, false, ""),
	}
	for _, market := range marketsToSeed {
		if err := db.Create(market).Error; err != nil {
			t.Fatalf("seed market: %v", err)
		}
	}
}

func mustSearchMarket(id int64, username string, _ time.Time, title, description string, resolution, finalResolution time.Time, resolved bool, result string) any {
	market := modelstesting.GenerateMarket(id, username)
	market.QuestionTitle = title
	market.Description = description
	market.ResolutionDateTime = resolution
	market.FinalResolutionDateTime = finalResolution
	market.IsResolved = resolved
	market.ResolutionResult = result
	return &market
}

func TestServiceSearchMarketsFiltersByStatus(t *testing.T) {
	service, db, _ := setupServiceWithDB(t)

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

			assertSortedIDs(t, primaryIDs, tt.expectedIDs)

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
	service, _, _ := setupServiceWithDB(t)

	_, err := service.SearchMarkets(context.Background(), "   ", markets.SearchFilters{})
	if !errors.Is(err, markets.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	fallback := markets.NewService(nil, nil, nil, markets.Config{}, markets.WithSearchPolicy(nil))
	if fallback == nil {
		t.Fatalf("expected fallback service")
	}
}
