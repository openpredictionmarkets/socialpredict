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

func TestGormRepositoryCreateAndGetMarketGroup(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("moderator", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}

	childA := modelstesting.GenerateMarket(1101, creator.Username)
	childA.QuestionTitle = "Will Team A win the tournament?"
	childB := modelstesting.GenerateMarket(1102, creator.Username)
	childB.QuestionTitle = "Will Team B win the tournament?"
	childC := modelstesting.GenerateMarket(1103, creator.Username)
	childC.QuestionTitle = "Will Team C win the tournament?"
	for _, market := range []*models.Market{&childA, &childB, &childC} {
		if err := db.Create(market).Error; err != nil {
			t.Fatalf("seed child market %d: %v", market.ID, err)
		}
	}

	resolutionTime := time.Now().UTC().Add(30 * 24 * time.Hour).Truncate(time.Second)
	group := &dmarkets.MarketGroup{
		QuestionTitle:      "Who will win the tournament?",
		Description:        "One parent question with several binary child markets.",
		ProposalCost:       10,
		CreatorUsername:    creator.Username,
		ResolutionDateTime: resolutionTime,
	}
	members := []dmarkets.MarketGroupMember{
		{MarketID: childC.ID, AnswerLabel: "Team C", DisplayOrder: 2},
		{MarketID: childA.ID, AnswerLabel: "Team A", DisplayOrder: 0},
		{MarketID: childB.ID, AnswerLabel: "Team B", DisplayOrder: 1},
	}

	if err := repo.CreateMarketGroup(ctx, group, members); err != nil {
		t.Fatalf("CreateMarketGroup returned error: %v", err)
	}
	if group.ID == 0 {
		t.Fatalf("expected group ID to be set")
	}
	if group.GroupType != dmarkets.MarketGroupTypeMultipleChoiceBinary {
		t.Fatalf("expected default group type, got %q", group.GroupType)
	}
	if group.ProbabilityPolicy != dmarkets.MarketGroupProbabilityPolicyIndependentBinary {
		t.Fatalf("expected independent probability policy, got %q", group.ProbabilityPolicy)
	}
	if group.StewardUsername != creator.Username {
		t.Fatalf("expected steward to default to creator, got %q", group.StewardUsername)
	}

	got, err := repo.GetMarketGroup(ctx, group.ID)
	if err != nil {
		t.Fatalf("GetMarketGroup returned error: %v", err)
	}
	if got.QuestionTitle != group.QuestionTitle || got.ProposalCost != 10 || got.ResolutionDateTime.IsZero() {
		t.Fatalf("unexpected group: %+v", got)
	}
	if len(got.Members) != 3 {
		t.Fatalf("expected 3 members, got %d", len(got.Members))
	}
	if got.Members[0].AnswerLabel != "Team A" || got.Members[1].AnswerLabel != "Team B" || got.Members[2].AnswerLabel != "Team C" {
		t.Fatalf("members not returned in display order: %+v", got.Members)
	}
	if got.Members[0].MarketID != childA.ID {
		t.Fatalf("expected first member market ID %d, got %d", childA.ID, got.Members[0].MarketID)
	}

	lookedUp, err := repo.GetMarketGroupForMarket(ctx, childB.ID)
	if err != nil {
		t.Fatalf("GetMarketGroupForMarket returned error: %v", err)
	}
	if lookedUp.ID != group.ID || len(lookedUp.Members) != 3 {
		t.Fatalf("unexpected looked up group: %+v", lookedUp)
	}
	if lookedUp.Members[1].AnswerLabel != "Team B" {
		t.Fatalf("expected looked up group to include ordered members, got %+v", lookedUp.Members)
	}
}

func TestGormRepositoryCreateMarketGroupRejectsInvalidMembers(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	group := &dmarkets.MarketGroup{
		QuestionTitle:      "Invalid group",
		Description:        "Invalid group",
		CreatorUsername:    "moderator",
		ResolutionDateTime: time.Now().UTC().Add(24 * time.Hour),
	}

	err := repo.CreateMarketGroup(ctx, group, []dmarkets.MarketGroupMember{
		{MarketID: 1, AnswerLabel: "Team A", DisplayOrder: 0},
	})
	if !errors.Is(err, dmarkets.ErrInvalidInput) {
		t.Fatalf("CreateMarketGroup error = %v, want ErrInvalidInput", err)
	}

	var count int64
	if err := db.Model(&models.MarketGroup{}).Count(&count).Error; err != nil {
		t.Fatalf("count groups: %v", err)
	}
	if count != 0 {
		t.Fatalf("invalid group should not be persisted, got %d rows", count)
	}
}

func TestGormRepositoryGetMarketGroupMissing(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)

	if _, err := repo.GetMarketGroup(context.Background(), 9999); !errors.Is(err, dmarkets.ErrMarketGroupNotFound) {
		t.Fatalf("GetMarketGroup error = %v, want ErrMarketGroupNotFound", err)
	}
	if _, err := repo.GetMarketGroupForMarket(context.Background(), 9999); !errors.Is(err, dmarkets.ErrMarketGroupNotFound) {
		t.Fatalf("GetMarketGroupForMarket error = %v, want ErrMarketGroupNotFound", err)
	}
}
