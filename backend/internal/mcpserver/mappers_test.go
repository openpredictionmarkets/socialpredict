package mcpserver

import (
	"testing"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
)

func TestMarketTagOutputFromDomain(t *testing.T) {
	got := MarketTagOutputFromDomain(dmarkets.MarketTag{
		ID: 2, Slug: "macro", DisplayName: "Macro", Description: "Macro markets", ColorKey: "blue", SortOrder: 7, IsActive: true,
	})
	if got.ID != 2 || got.Slug != "macro" || got.DisplayName != "Macro" || !got.IsActive {
		t.Fatalf("tag output = %#v", got)
	}
}

func TestMarketOverviewOutputIncludesVolumeAndTags(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	overview := &dmarkets.MarketOverview{
		Market: &dmarkets.Market{
			ID: 11, QuestionTitle: "Will it rain?", Description: "Rain in Austin", OutcomeType: "BINARY",
			ResolutionDateTime: now, CreatorUsername: "alice", StewardUsername: "bob",
			YesLabel: "YES", NoLabel: "NO", Status: dmarkets.MarketStatusActive,
			LifecycleStatus: dmarkets.MarketLifecyclePublished, CreatedAt: now, UpdatedAt: now,
			InitialProbability: 0.5, UTCOffset: -5,
			Tags: []dmarkets.MarketTag{{ID: 3, Slug: "weather", DisplayName: "Weather", IsActive: true}},
		},
		Creator:         &dmarkets.CreatorSummary{Username: "alice", DisplayName: "Alice"},
		LastProbability: 0.62,
		NumUsers:        9,
		TotalVolume:     1234,
		MarketDust:      12,
	}

	got := MarketOverviewOutputFromDomain(overview)
	if got.Market.ID != 11 || got.LastProbability != 0.62 || got.TotalVolume != 1234 || got.MarketDust != 12 {
		t.Fatalf("overview output = %#v", got)
	}
	if len(got.Market.Tags) != 1 || got.Market.Tags[0].Slug != "weather" {
		t.Fatalf("tags = %#v", got.Market.Tags)
	}
}

func TestDiscoveryRowOutputFromGroupedRowPreservesGroupAndChildren(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	row := dmarkets.MarketDiscoveryRow{
		Group: &dmarkets.MarketGroup{ID: 5, QuestionTitle: "Which team wins?", LifecycleStatus: dmarkets.MarketLifecyclePublished, CreatorUsername: "alice", Members: []dmarkets.MarketGroupMember{
			{MarketID: 10, AnswerLabel: "A", DisplayOrder: 1},
			{MarketID: 20, AnswerLabel: "B", DisplayOrder: 2},
		}},
		Children: []*dmarkets.Market{
			{ID: 10, QuestionTitle: "A wins", Status: dmarkets.MarketStatusActive, LifecycleStatus: dmarkets.MarketLifecyclePublished, CreatedAt: now, UpdatedAt: now},
			{ID: 20, QuestionTitle: "B wins", Status: dmarkets.MarketStatusActive, LifecycleStatus: dmarkets.MarketLifecyclePublished, CreatedAt: now, UpdatedAt: now},
		},
	}
	got := DiscoveryRowOutputFromDomain(row)
	if !got.IsMarketGroup || got.Group.ID != 5 || len(got.ChildMarkets) != 2 {
		t.Fatalf("grouped row output = %#v", got)
	}
}
