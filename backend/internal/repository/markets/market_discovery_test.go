package markets

import (
	"context"
	"testing"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

func TestGormRepositoryListMarketDiscoveryGroupsBeforePagination(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	creator := seedDiscoveryUser(t, db, "discovery_creator")

	children := seedDiscoveryGroup(t, db, discoveryGroupSeed{
		GroupID:     100,
		Title:       "Match winner",
		Description: "Grouped match winner",
		Creator:     creator,
		CreatedAt:   now.Add(10 * time.Minute),
		Answers: []discoveryAnswerSeed{
			{MarketID: 101, Label: "Spain", Title: "Answer child one"},
			{MarketID: 102, Label: "Canada", Title: "Answer child two"},
			{MarketID: 103, Label: "Draw", Title: "Answer child three"},
		},
	})
	standalone := seedDiscoveryMarket(t, db, 200, creator, "Standalone market", now.Add(-10*time.Minute))

	firstPage, err := repo.ListMarketDiscovery(ctx, dmarkets.ListFilters{
		Status: dmarkets.MarketStatusActive,
		Limit:  1,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListMarketDiscovery first page: %v", err)
	}
	if firstPage.Total != 2 {
		t.Fatalf("expected grouped total 2, got %d", firstPage.Total)
	}
	if len(firstPage.Rows) != 1 {
		t.Fatalf("expected one page row, got %d", len(firstPage.Rows))
	}
	if firstPage.Rows[0].Group == nil || firstPage.Rows[0].Group.ID != 100 {
		t.Fatalf("expected group row first, got %+v", firstPage.Rows[0])
	}
	if got := childIDsFromDiscoveryRows(firstPage.Rows[0].Children); len(got) != len(children) || got[0] != 101 || got[1] != 102 || got[2] != 103 {
		t.Fatalf("expected all ordered children in first grouped row, got %v", got)
	}

	secondPage, err := repo.ListMarketDiscovery(ctx, dmarkets.ListFilters{
		Status: dmarkets.MarketStatusActive,
		Limit:  1,
		Offset: 1,
	})
	if err != nil {
		t.Fatalf("ListMarketDiscovery second page: %v", err)
	}
	if secondPage.Total != 2 || len(secondPage.Rows) != 1 {
		t.Fatalf("expected one standalone row on second page with total 2, got total=%d rows=%d", secondPage.Total, len(secondPage.Rows))
	}
	if secondPage.Rows[0].Group != nil || secondPage.Rows[0].Market == nil || secondPage.Rows[0].Market.ID != standalone.ID {
		t.Fatalf("expected standalone market on second page, got %+v", secondPage.Rows[0])
	}
}

func TestGormRepositorySearchMarketDiscoveryMatchesAnswerLabelAndReturnsParentRow(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	creator := seedDiscoveryUser(t, db, "answer_label_creator")

	seedDiscoveryGroup(t, db, discoveryGroupSeed{
		GroupID:     300,
		Title:       "Championship outcome",
		Description: "Parent group without the searched answer word",
		Creator:     creator,
		CreatedAt:   now,
		Answers: []discoveryAnswerSeed{
			{MarketID: 301, Label: "France", Title: "Outcome child alpha"},
			{MarketID: 302, Label: "Germany", Title: "Outcome child beta"},
			{MarketID: 303, Label: "Brazil", Title: "Outcome child gamma"},
		},
	})

	page, err := repo.SearchMarketDiscovery(ctx, "germany", dmarkets.SearchFilters{
		Status: dmarkets.MarketStatusActive,
		Limit:  20,
	})
	if err != nil {
		t.Fatalf("SearchMarketDiscovery: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("expected one grouped search row, got total=%d rows=%d", page.Total, len(page.Rows))
	}
	row := page.Rows[0]
	if row.Group == nil || row.Group.ID != 300 {
		t.Fatalf("expected parent group row, got %+v", row)
	}
	if row.Group.QuestionTitle != "Championship outcome" {
		t.Fatalf("expected parent title, got %q", row.Group.QuestionTitle)
	}
	if got := childIDsFromDiscoveryRows(row.Children); len(got) != 3 || got[0] != 301 || got[1] != 302 || got[2] != 303 {
		t.Fatalf("expected all group children after single answer-label match, got %v", got)
	}
}

type discoveryGroupSeed struct {
	GroupID     int64
	Title       string
	Description string
	Creator     string
	CreatedAt   time.Time
	Answers     []discoveryAnswerSeed
}

type discoveryAnswerSeed struct {
	MarketID int64
	Label    string
	Title    string
}

func seedDiscoveryUser(t *testing.T, db *gorm.DB, username string) string {
	t.Helper()
	user := modelstesting.GenerateUser(username, 1000)
	user.UserType = "MODERATOR"
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user %s: %v", username, err)
	}
	return user.Username
}

func seedDiscoveryMarket(t *testing.T, db *gorm.DB, id int64, creator string, title string, createdAt time.Time) models.Market {
	t.Helper()
	market := modelstesting.GenerateMarket(id, creator)
	market.QuestionTitle = title
	market.Description = "Discovery market"
	market.LifecycleStatus = dmarkets.MarketLifecyclePublished
	market.IsResolved = false
	market.ResolutionDateTime = createdAt.Add(30 * 24 * time.Hour)
	market.CreatedAt = createdAt
	market.UpdatedAt = createdAt
	market.StewardUsername = creator
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market %d: %v", id, err)
	}
	return market
}

func seedDiscoveryGroup(t *testing.T, db *gorm.DB, seed discoveryGroupSeed) []models.Market {
	t.Helper()
	children := make([]models.Market, 0, len(seed.Answers))
	for index, answer := range seed.Answers {
		market := seedDiscoveryMarket(t, db, answer.MarketID, seed.Creator, answer.Title, seed.CreatedAt.Add(time.Duration(index)*time.Second))
		children = append(children, market)
	}
	group := models.MarketGroup{
		ID:                 seed.GroupID,
		QuestionTitle:      seed.Title,
		Description:        seed.Description,
		GroupType:          dmarkets.MarketGroupTypeMultipleChoiceBinary,
		ProbabilityPolicy:  dmarkets.MarketGroupProbabilityPolicyIndependentBinary,
		ResolutionPolicy:   dmarkets.MarketGroupResolutionPolicyIndependentChildren,
		LifecycleStatus:    dmarkets.MarketLifecyclePublished,
		ProposalCost:       10,
		CreatorUsername:    seed.Creator,
		StewardUsername:    seed.Creator,
		ResolutionDateTime: seed.CreatedAt.Add(30 * 24 * time.Hour),
	}
	group.CreatedAt = seed.CreatedAt
	group.UpdatedAt = seed.CreatedAt
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed group %d: %v", seed.GroupID, err)
	}
	members := make([]models.MarketGroupMember, 0, len(seed.Answers))
	for index, answer := range seed.Answers {
		members = append(members, models.MarketGroupMember{
			GroupID:      seed.GroupID,
			MarketID:     answer.MarketID,
			AnswerLabel:  answer.Label,
			DisplayOrder: index,
		})
	}
	if err := db.Create(&members).Error; err != nil {
		t.Fatalf("seed group members: %v", err)
	}
	return children
}

func childIDsFromDiscoveryRows(children []*dmarkets.Market) []int64 {
	ids := make([]int64, 0, len(children))
	for _, child := range children {
		if child != nil {
			ids = append(ids, child.ID)
		}
	}
	return ids
}
