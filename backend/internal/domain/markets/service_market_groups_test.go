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
	if group.AutoApproveAnswerAdditions {
		t.Fatalf("auto approve answer additions should default false")
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

func TestCreateMarketGroupPersistsAnswerAdditionPolicy(t *testing.T) {
	now := marketsTestTime()
	var createdGroup *markets.MarketGroup
	nextMarketID := int64(100)

	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.listMarketTagsFunc = func(context.Context, bool) ([]markets.MarketTag, error) {
			return []markets.MarketTag{}, nil
		}
		repo.createFunc = func(_ context.Context, market *markets.Market) error {
			nextMarketID++
			market.ID = nextMarketID
			return nil
		}
		repo.createMarketGroupFunc = func(_ context.Context, group *markets.MarketGroup, members []markets.MarketGroupMember) error {
			group.ID = 8
			group.Members = members
			createdGroup = group
			return nil
		}
	})
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.validateUserBalanceFunc = func(context.Context, string, int64, int64) error { return nil }
		service.deductBalanceFunc = func(context.Context, string, int64) error { return nil }
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

	req := validMarketGroupCreateRequest(now)
	req.TagSlugs = nil
	req.AutoApproveAnswerAdditions = true
	group, err := service.CreateMarketGroup(context.Background(), req, "moderator")
	if err != nil {
		t.Fatalf("CreateMarketGroup returned error: %v", err)
	}
	if group != createdGroup || !group.AutoApproveAnswerAdditions {
		t.Fatalf("expected group auto-approval policy to persist: %+v", group)
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

func TestMarketGroupActivityAggregatesChildRows(t *testing.T) {
	now := marketsTestTime()
	group := &markets.MarketGroup{
		ID:              50,
		QuestionTitle:   "Spain vs Canada",
		LifecycleStatus: markets.MarketLifecyclePublished,
		Members: []markets.MarketGroupMember{
			{ID: 1, GroupID: 50, MarketID: 101, AnswerLabel: "Spain", DisplayOrder: 0},
			{ID: 2, GroupID: 50, MarketID: 102, AnswerLabel: "Canada", DisplayOrder: 1},
		},
	}
	marketsByID := map[int64]*markets.Market{
		101: {ID: 101, Status: markets.MarketStatusActive, CreatedAt: now, ResolutionDateTime: now.Add(24 * time.Hour)},
		102: {ID: 102, Status: markets.MarketStatusActive, CreatedAt: now, ResolutionDateTime: now.Add(24 * time.Hour)},
	}
	betsByMarket := map[int64][]*markets.Bet{
		101: {
			{Username: "alice", MarketID: 101, Amount: 10, Outcome: "YES", PlacedAt: now.Add(1 * time.Minute), CreatedAt: now.Add(1 * time.Minute)},
		},
		102: {
			{Username: "bob", MarketID: 102, Amount: 7, Outcome: "NO", PlacedAt: now.Add(2 * time.Minute), CreatedAt: now.Add(2 * time.Minute)},
		},
	}
	positionsByMarket := map[int64]markets.MarketPositions{
		101: {
			{Username: "alice", MarketID: 101, YesSharesOwned: 3, Value: 3, TotalSpent: 10, TotalSpentInPlay: 10},
		},
		102: {
			{Username: "alice", MarketID: 102, NoSharesOwned: 4, Value: 4, TotalSpent: 7, TotalSpentInPlay: 7},
		},
	}
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.getMarketGroupFunc = func(_ context.Context, groupID int64) (*markets.MarketGroup, error) {
			if groupID != group.ID {
				return nil, markets.ErrMarketGroupNotFound
			}
			return group, nil
		}
		repo.getByIDFunc = func(_ context.Context, marketID int64) (*markets.Market, error) {
			return marketsByID[marketID], nil
		}
		repo.listBetsForMarketFunc = func(_ context.Context, marketID int64) ([]*markets.Bet, error) {
			return betsByMarket[marketID], nil
		}
		repo.listMarketPositionsFunc = func(_ context.Context, marketID int64) (markets.MarketPositions, error) {
			return positionsByMarket[marketID], nil
		}
	})
	service := markets.NewService(repo, nil, newProjectionClock(now), markets.Config{})

	bets, err := service.GetMarketGroupBetsPage(context.Background(), group.ID, markets.Page{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("GetMarketGroupBetsPage returned error: %v", err)
	}
	if bets.Total != 2 || len(bets.Bets) != 2 {
		t.Fatalf("expected two grouped bets, got total=%d len=%d", bets.Total, len(bets.Bets))
	}
	if bets.Bets[0].AnswerLabel != "Canada" {
		t.Fatalf("expected newest bet first from Canada child, got %+v", bets.Bets[0])
	}

	positions, err := service.GetMarketGroupPositionsPage(context.Background(), group.ID, markets.Page{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("GetMarketGroupPositionsPage returned error: %v", err)
	}
	if positions.Total != 1 || len(positions.Positions) != 1 {
		t.Fatalf("expected one grouped position row, got total=%d len=%d", positions.Total, len(positions.Positions))
	}
	alice := positions.Positions[0]
	if alice.Username != "alice" || alice.YesSharesOwned != 3 || alice.NoSharesOwned != 4 || len(alice.Answers) != 2 {
		t.Fatalf("unexpected grouped position: %+v", alice)
	}
}

func TestResolveMarketGroupExclusiveYesResolvesChildrenAndMarksParent(t *testing.T) {
	now := marketsTestTime()
	group := &markets.MarketGroup{
		ID:              9,
		QuestionTitle:   "Match winner",
		LifecycleStatus: markets.MarketLifecyclePublished,
		CreatorUsername: "moderator",
		StewardUsername: "moderator",
		Members: []markets.MarketGroupMember{
			{ID: 1, GroupID: 9, MarketID: 101, AnswerLabel: "Home", DisplayOrder: 0},
			{ID: 2, GroupID: 9, MarketID: 102, AnswerLabel: "Away", DisplayOrder: 1},
			{ID: 3, GroupID: 9, MarketID: 103, AnswerLabel: "Draw", DisplayOrder: 2},
		},
	}
	marketsByID := map[int64]*markets.Market{}
	for _, member := range group.Members {
		marketsByID[member.MarketID] = &markets.Market{
			ID:              member.MarketID,
			QuestionTitle:   member.AnswerLabel,
			Status:          markets.MarketStatusActive,
			LifecycleStatus: markets.MarketLifecyclePublished,
			CreatorUsername: "moderator",
			StewardUsername: "moderator",
		}
	}
	var resolved []markets.MarketGroupChildResolution
	var markedGroupID int64
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.getMarketGroupFunc = func(_ context.Context, groupID int64) (*markets.MarketGroup, error) {
			if groupID != group.ID {
				t.Fatalf("groupID = %d, want %d", groupID, group.ID)
			}
			return group, nil
		}
		repo.getByIDFunc = func(_ context.Context, marketID int64) (*markets.Market, error) {
			return marketsByID[marketID], nil
		}
		repo.resolveMarketFunc = func(_ context.Context, marketID int64, resolution string) error {
			resolved = append(resolved, markets.MarketGroupChildResolution{MarketID: marketID, Resolution: resolution})
			return nil
		}
		repo.calculatePayoutPositionsFunc = func(context.Context, int64) ([]*markets.PayoutPosition, error) {
			return []*markets.PayoutPosition{}, nil
		}
		repo.listBetsForMarketFunc = func(context.Context, int64) ([]*markets.Bet, error) {
			return []*markets.Bet{}, nil
		}
		repo.markMarketGroupResolvedFunc = func(_ context.Context, groupID int64, resolvedAt time.Time) error {
			markedGroupID = groupID
			if !resolvedAt.Equal(now) {
				t.Fatalf("resolvedAt = %s, want %s", resolvedAt, now)
			}
			return nil
		}
	})
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return &dusers.PublicUser{
				Username:        username,
				UserType:        string(dusers.UserTypeModerator),
				ModeratorStatus: dusers.ModeratorStatusActive,
			}, nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{GameMode: "moderator"})

	resolvedGroup, err := service.ResolveMarketGroup(context.Background(), group.ID, markets.MarketGroupResolveRequest{
		Mode:            markets.MarketGroupResolveModeExclusiveYes,
		WinningMarketID: 102,
	}, "moderator")
	if err != nil {
		t.Fatalf("ResolveMarketGroup returned error: %v", err)
	}

	expected := []markets.MarketGroupChildResolution{
		{MarketID: 101, Resolution: "NO"},
		{MarketID: 102, Resolution: "YES"},
		{MarketID: 103, Resolution: "NO"},
	}
	if !reflect.DeepEqual(resolved, expected) {
		t.Fatalf("resolved = %+v, want %+v", resolved, expected)
	}
	if markedGroupID != group.ID {
		t.Fatalf("markedGroupID = %d, want %d", markedGroupID, group.ID)
	}
	if resolvedGroup.LifecycleStatus != markets.MarketLifecycleResolved {
		t.Fatalf("resolved group lifecycle = %q, want resolved", resolvedGroup.LifecycleStatus)
	}
}

func TestResolveMarketGroupRejectsUnpublishedChildWithDetailsBeforeMutating(t *testing.T) {
	now := marketsTestTime()
	group := &markets.MarketGroup{
		ID:              11,
		QuestionTitle:   "Match winner",
		LifecycleStatus: markets.MarketLifecyclePublished,
		CreatorUsername: "moderator",
		StewardUsername: "moderator",
		Members: []markets.MarketGroupMember{
			{ID: 1, GroupID: 11, MarketID: 111, AnswerLabel: "Home", DisplayOrder: 0},
			{ID: 2, GroupID: 11, MarketID: 112, AnswerLabel: "Away", DisplayOrder: 1},
		},
	}
	marketsByID := map[int64]*markets.Market{
		111: &markets.Market{
			ID:              111,
			QuestionTitle:   "Home",
			Status:          markets.MarketStatusActive,
			LifecycleStatus: markets.MarketLifecyclePublished,
			CreatorUsername: "moderator",
			StewardUsername: "moderator",
		},
		112: &markets.Market{
			ID:              112,
			QuestionTitle:   "Away",
			Status:          markets.MarketStatusActive,
			LifecycleStatus: markets.MarketLifecycleProposed,
			CreatorUsername: "moderator",
			StewardUsername: "moderator",
		},
	}
	var resolved []int64
	var markedGroupID int64
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.getMarketGroupFunc = func(context.Context, int64) (*markets.MarketGroup, error) {
			return group, nil
		}
		repo.getByIDFunc = func(_ context.Context, marketID int64) (*markets.Market, error) {
			return marketsByID[marketID], nil
		}
		repo.resolveMarketFunc = func(_ context.Context, marketID int64, resolution string) error {
			resolved = append(resolved, marketID)
			return nil
		}
		repo.markMarketGroupResolvedFunc = func(_ context.Context, groupID int64, resolvedAt time.Time) error {
			markedGroupID = groupID
			return nil
		}
	})
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return &dusers.PublicUser{
				Username:        username,
				UserType:        string(dusers.UserTypeModerator),
				ModeratorStatus: dusers.ModeratorStatusActive,
			}, nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{GameMode: "moderator"})

	_, err := service.ResolveMarketGroup(context.Background(), group.ID, markets.MarketGroupResolveRequest{
		Mode:            markets.MarketGroupResolveModeExclusiveYes,
		WinningMarketID: 111,
	}, "moderator")
	if err == nil {
		t.Fatalf("ResolveMarketGroup error = nil, want unpublished child error")
	}
	if !errors.Is(err, markets.ErrInvalidState) {
		t.Fatalf("ResolveMarketGroup error = %v, want ErrInvalidState-compatible error", err)
	}
	var childErr *markets.MarketGroupChildNotPublishedError
	if !errors.As(err, &childErr) {
		t.Fatalf("ResolveMarketGroup error = %T %v, want MarketGroupChildNotPublishedError", err, err)
	}
	if childErr.MarketID != 112 || childErr.AnswerLabel != "Away" || childErr.LifecycleStatus != markets.MarketLifecycleProposed {
		t.Fatalf("unexpected unpublished child details: %+v", childErr)
	}
	if len(resolved) != 0 {
		t.Fatalf("expected no child markets to resolve before validation passes, got %+v", resolved)
	}
	if markedGroupID != 0 {
		t.Fatalf("expected parent group not to be marked resolved, got %d", markedGroupID)
	}
}

func TestResolveMarketGroupPaysSingleGroupWorkProfit(t *testing.T) {
	now := marketsTestTime()
	group := &markets.MarketGroup{
		ID:              19,
		QuestionTitle:   "Who wins?",
		LifecycleStatus: markets.MarketLifecyclePublished,
		CreatorUsername: "creator",
		StewardUsername: "steward",
		ProposalCost:    2,
		Members: []markets.MarketGroupMember{
			{ID: 1, GroupID: 19, MarketID: 201, AnswerLabel: "Home", DisplayOrder: 0},
			{ID: 2, GroupID: 19, MarketID: 202, AnswerLabel: "Away", DisplayOrder: 1},
		},
	}
	marketsByID := map[int64]*markets.Market{}
	for _, member := range group.Members {
		marketsByID[member.MarketID] = &markets.Market{
			ID:              member.MarketID,
			Status:          markets.MarketStatusActive,
			LifecycleStatus: markets.MarketLifecyclePublished,
			CreatorUsername: "creator",
			StewardUsername: "steward",
		}
	}
	betsByMarket := map[int64][]*markets.Bet{
		201: {
			{Username: "alice", Amount: 10, Outcome: "YES"},
			{Username: "alice", Amount: -2, Outcome: "YES"},
			{Username: "bob", Amount: 5, Outcome: "NO"},
		},
		202: {
			{Username: "alice", Amount: 4, Outcome: "YES"},
			{Username: "carol", Amount: 6, Outcome: "YES"},
		},
	}
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.getMarketGroupFunc = func(context.Context, int64) (*markets.MarketGroup, error) {
			return group, nil
		}
		repo.getByIDFunc = func(_ context.Context, marketID int64) (*markets.Market, error) {
			return marketsByID[marketID], nil
		}
		repo.resolveMarketFunc = func(context.Context, int64, string) error {
			return nil
		}
		repo.calculatePayoutPositionsFunc = func(context.Context, int64) ([]*markets.PayoutPosition, error) {
			return []*markets.PayoutPosition{}, nil
		}
		repo.listBetsForMarketFunc = func(_ context.Context, marketID int64) ([]*markets.Bet, error) {
			return betsByMarket[marketID], nil
		}
		repo.markMarketGroupResolvedFunc = func(context.Context, int64, time.Time) error {
			return nil
		}
	})
	var applied []struct {
		username string
		amount   int64
		txType   string
	}
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return &dusers.PublicUser{
				Username:        username,
				UserType:        string(dusers.UserTypeModerator),
				ModeratorStatus: dusers.ModeratorStatusActive,
			}, nil
		}
		service.applyTransactionFunc = func(_ context.Context, username string, amount int64, txType string) error {
			applied = append(applied, struct {
				username string
				amount   int64
				txType   string
			}{username: username, amount: amount, txType: txType})
			return nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{
		GameMode:               "moderator",
		InitialBetFee:          2,
		CreateMarketCost:       10,
		MarketApprovalRequired: true,
	})

	_, err := service.ResolveMarketGroup(context.Background(), group.ID, markets.MarketGroupResolveRequest{
		Mode:            markets.MarketGroupResolveModeExclusiveYes,
		WinningMarketID: 201,
	}, "steward")
	if err != nil {
		t.Fatalf("ResolveMarketGroup returned error: %v", err)
	}

	if len(applied) != 1 {
		t.Fatalf("expected one work-profit transaction, got %+v", applied)
	}
	if applied[0].username != "steward" || applied[0].amount != 6 || applied[0].txType != dusers.TransactionWorkProfit {
		t.Fatalf("unexpected work-profit transaction: %+v", applied[0])
	}
}
