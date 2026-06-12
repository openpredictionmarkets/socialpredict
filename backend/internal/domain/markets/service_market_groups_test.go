package markets_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
)

func validMarketGroupCreateRequest(now time.Time) markets.MarketGroupCreateRequest {
	return markets.MarketGroupCreateRequest{
		QuestionTitle:      "Who will win the tournament?",
		Description:        "Pick the answer by trading its child binary market.",
		ResolutionDateTime: now.Add(48 * time.Hour),
		AnswerLabels:       []string{"Alice", "Bob", "Carol"},
		TagSlugs:           []string{"sports"},
	}
}

func TestCreateMarketGroupCreatesZeroCostChildMarketsAndChargesParentOnce(t *testing.T) {
	now := marketsTestTime()
	var createdMarkets []*markets.Market
	var createdGroup *markets.MarketGroup
	var createdMembers []markets.MarketGroupMember
	var balanceChecks []int64
	var deductions []int64
	var tagAssignments []int64
	nextMarketID := int64(100)

	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.listMarketTagsFunc = func(context.Context, bool) ([]markets.MarketTag, error) {
			return []markets.MarketTag{{ID: 1, Slug: "sports", DisplayName: "Sports", IsActive: true}}, nil
		}
		repo.setMarketTagsFunc = func(_ context.Context, marketID int64, tagSlugs []string, assignedBy string, source string, assignedAt time.Time) ([]markets.MarketTag, error) {
			if !reflect.DeepEqual(tagSlugs, []string{"sports"}) {
				t.Fatalf("tagSlugs = %v, want [sports]", tagSlugs)
			}
			if assignedBy != "moderator" || source != markets.MarketTagAssignmentSourceCreate || !assignedAt.Equal(now) {
				t.Fatalf("tag assignment metadata mismatch: by=%q source=%q at=%s", assignedBy, source, assignedAt)
			}
			tagAssignments = append(tagAssignments, marketID)
			return []markets.MarketTag{{ID: 1, Slug: "sports", DisplayName: "Sports", IsActive: true}}, nil
		}
		repo.createFunc = func(_ context.Context, market *markets.Market) error {
			nextMarketID++
			market.ID = nextMarketID
			createdMarkets = append(createdMarkets, market)
			return nil
		}
		repo.createMarketGroupFunc = func(_ context.Context, group *markets.MarketGroup, members []markets.MarketGroupMember) error {
			group.ID = 7
			for index := range members {
				members[index].ID = int64(index + 1)
				members[index].GroupID = group.ID
			}
			group.Members = append([]markets.MarketGroupMember(nil), members...)
			createdGroup = group
			createdMembers = append([]markets.MarketGroupMember(nil), members...)
			return nil
		}
	})
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.validateUserBalanceFunc = func(_ context.Context, username string, amount int64, maxDebt int64) error {
			if username != "moderator" {
				t.Fatalf("balance username = %q, want moderator", username)
			}
			balanceChecks = append(balanceChecks, amount)
			return nil
		}
		service.deductBalanceFunc = func(_ context.Context, username string, amount int64) error {
			if username != "moderator" {
				t.Fatalf("deduct username = %q, want moderator", username)
			}
			deductions = append(deductions, amount)
			return nil
		}
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return &dusers.PublicUser{
				Username:        username,
				UserType:        string(dusers.UserTypeModerator),
				ModeratorStatus: dusers.ModeratorStatusActive,
			}, nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{
		GameMode:               "moderator",
		MarketApprovalRequired: true,
		CreateMarketCost:       10,
		MaximumDebtAllowed:     500,
	})

	group, err := service.CreateMarketGroup(context.Background(), validMarketGroupCreateRequest(now), "moderator")
	if err != nil {
		t.Fatalf("CreateMarketGroup returned error: %v", err)
	}

	if group != createdGroup || group.ID != 7 || group.ProposalCost != 10 {
		t.Fatalf("unexpected group: %+v created=%+v", group, createdGroup)
	}
	if group.LifecycleStatus != markets.MarketLifecycleProposed {
		t.Fatalf("group lifecycle = %q, want proposed", group.LifecycleStatus)
	}
	if len(balanceChecks) != 1 || balanceChecks[0] != 10 || len(deductions) != 1 || deductions[0] != 10 {
		t.Fatalf("cost should be checked/deducted once: checks=%v deductions=%v", balanceChecks, deductions)
	}
	if len(createdMarkets) != 3 || len(createdMembers) != 3 {
		t.Fatalf("created markets=%d members=%d, want 3 each", len(createdMarkets), len(createdMembers))
	}
	for index, market := range createdMarkets {
		if market.ProposalCost != 0 {
			t.Fatalf("child %d proposal cost = %d, want 0", index, market.ProposalCost)
		}
		if market.Status != markets.MarketLifecycleProposed || market.LifecycleStatus != markets.MarketLifecycleProposed {
			t.Fatalf("child %d lifecycle mismatch: %+v", index, market)
		}
		if market.YesLabel != "YES" || market.NoLabel != "NO" || market.OutcomeType != "BINARY" {
			t.Fatalf("child %d should be normal binary YES/NO market: %+v", index, market)
		}
		if createdMembers[index].MarketID != market.ID {
			t.Fatalf("member %d marketID = %d, want %d", index, createdMembers[index].MarketID, market.ID)
		}
	}
	if !reflect.DeepEqual(tagAssignments, []int64{101, 102, 103}) {
		t.Fatalf("tag assignments = %v, want child market IDs [101 102 103]", tagAssignments)
	}
}

func TestCreateMarketGroupRejectsInvalidAnswersBeforeCreatingChildren(t *testing.T) {
	now := marketsTestTime()
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.createFunc = func(context.Context, *markets.Market) error {
			t.Fatal("Create should not be called for invalid answers")
			return nil
		}
	})
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	tooManyLabels := make([]string, markets.MaxMarketGroupAnswers+1)
	for index := range tooManyLabels {
		tooManyLabels[index] = "Answer " + string(rune('A'+index))
	}

	cases := []struct {
		name   string
		labels []string
	}{
		{name: "duplicate labels", labels: []string{"Alice", " alice "}},
		{name: "too few labels", labels: []string{"Only one"}},
		{name: "too many labels", labels: tooManyLabels},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := validMarketGroupCreateRequest(now)
			req.AnswerLabels = tc.labels

			_, err := service.CreateMarketGroup(context.Background(), req, "moderator")
			if !errors.Is(err, markets.ErrInvalidInput) {
				t.Fatalf("CreateMarketGroup error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestGetMarketGroupOverviewReturnsChildMarketOverviews(t *testing.T) {
	now := marketsTestTime()
	group := &markets.MarketGroup{
		ID:                 3,
		QuestionTitle:      "Who wins?",
		GroupType:          markets.MarketGroupTypeMultipleChoiceBinary,
		ProbabilityPolicy:  markets.MarketGroupProbabilityPolicyIndependentBinary,
		ResolutionPolicy:   markets.MarketGroupResolutionPolicyIndependentChildren,
		LifecycleStatus:    markets.MarketLifecyclePublished,
		CreatorUsername:    "moderator",
		StewardUsername:    "moderator",
		ResolutionDateTime: now.Add(48 * time.Hour),
		Members: []markets.MarketGroupMember{
			{ID: 10, GroupID: 3, MarketID: 51, AnswerLabel: "Bob", DisplayOrder: 1},
			{ID: 9, GroupID: 3, MarketID: 50, AnswerLabel: "Alice", DisplayOrder: 0},
		},
	}
	marketsByID := map[int64]*markets.Market{
		50: {ID: 50, QuestionTitle: "Who wins? - Alice", CreatorUsername: "moderator", Status: markets.MarketStatusActive, LifecycleStatus: markets.MarketLifecyclePublished, ResolutionDateTime: now.Add(48 * time.Hour), CreatedAt: now, YesLabel: "YES", NoLabel: "NO"},
		51: {ID: 51, QuestionTitle: "Who wins? - Bob", CreatorUsername: "moderator", Status: markets.MarketStatusActive, LifecycleStatus: markets.MarketLifecyclePublished, ResolutionDateTime: now.Add(48 * time.Hour), CreatedAt: now, YesLabel: "YES", NoLabel: "NO"},
	}

	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.getMarketGroupFunc = func(context.Context, int64) (*markets.MarketGroup, error) {
			return group, nil
		}
		repo.getByIDFunc = func(_ context.Context, marketID int64) (*markets.Market, error) {
			return marketsByID[marketID], nil
		}
		repo.listBetsForMarketFunc = func(context.Context, int64) ([]*markets.Bet, error) {
			return []*markets.Bet{}, nil
		}
	})
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	overview, err := service.GetMarketGroupOverview(context.Background(), group.ID)
	if err != nil {
		t.Fatalf("GetMarketGroupOverview returned error: %v", err)
	}
	if overview.Group.ID != group.ID || len(overview.Answers) != 2 {
		t.Fatalf("unexpected overview: %+v", overview)
	}
	if overview.Answers[0].Member.AnswerLabel != "Alice" || overview.Answers[0].Overview.Market.ID != 50 {
		t.Fatalf("answers not sorted/hydrated: %+v", overview.Answers)
	}
	if overview.Answers[1].Member.AnswerLabel != "Bob" || overview.Answers[1].Overview.Market.ID != 51 {
		t.Fatalf("answers not sorted/hydrated: %+v", overview.Answers)
	}
}
