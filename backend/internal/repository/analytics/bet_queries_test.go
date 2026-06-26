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

func TestGormRepositoryListMarketGroupFeeRecordsHydratesMembersInBulkOrder(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	groups := []models.MarketGroup{
		{
			ID:                 2001,
			QuestionTitle:      "Group One",
			Description:        "First group",
			GroupType:          "MULTIPLE_CHOICE_BINARY",
			ProbabilityPolicy:  "INDEPENDENT_BINARY",
			ResolutionPolicy:   "ONLY_ONE_YES",
			LifecycleStatus:    "published",
			ProposalCost:       10,
			CreatorUsername:    "creator",
			ResolutionDateTime: time.Now().Add(24 * time.Hour),
		},
		{
			ID:                 2002,
			QuestionTitle:      "Group Two",
			Description:        "Second group",
			GroupType:          "MULTIPLE_CHOICE_BINARY",
			ProbabilityPolicy:  "INDEPENDENT_BINARY",
			ResolutionPolicy:   "ONLY_ONE_YES",
			LifecycleStatus:    "resolved",
			ProposalCost:       10,
			CreatorUsername:    "creator",
			StewardUsername:    "steward",
			ResolutionDateTime: time.Now().Add(24 * time.Hour),
		},
	}
	for i := range groups {
		if err := db.Create(&groups[i]).Error; err != nil {
			t.Fatalf("create group: %v", err)
		}
	}

	members := []models.MarketGroupMember{
		{GroupID: 2001, MarketID: 3002, AnswerLabel: "B", DisplayOrder: 2},
		{GroupID: 2001, MarketID: 3001, AnswerLabel: "A", DisplayOrder: 1},
		{GroupID: 2002, MarketID: 4001, AnswerLabel: "Only", DisplayOrder: 1},
	}
	for i := range members {
		if err := db.Create(&members[i]).Error; err != nil {
			t.Fatalf("create member: %v", err)
		}
	}

	records, err := repo.ListMarketGroupFeeRecords(ctx)
	if err != nil {
		t.Fatalf("ListMarketGroupFeeRecords returned error: %v", err)
	}

	byID := make(map[uint]WorkProfitMarketGroupRecord, len(records))
	for _, record := range records {
		byID[record.ID] = record
	}
	first := byID[2001]
	second := byID[2002]
	if got := first.MemberMarketIDs; len(got) != 2 || got[0] != 3001 || got[1] != 3002 {
		t.Fatalf("expected first group members ordered by display order, got %+v", got)
	}
	if got := second.MemberMarketIDs; len(got) != 1 || got[0] != 4001 {
		t.Fatalf("expected second group member, got %+v", got)
	}
	if second.StewardUsername != "steward" {
		t.Fatalf("expected explicit steward to be preserved, got %q", second.StewardUsername)
	}
}
