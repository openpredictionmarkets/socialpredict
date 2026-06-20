package markets_test

import (
	"context"
	"errors"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	rmarkets "socialpredict/internal/repository/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func activeModeratorUser(username string) *dusers.PublicUser {
	return &dusers.PublicUser{Username: username, UserType: string(dusers.UserTypeModerator), ModeratorStatus: dusers.ModeratorStatusActive}
}

func seedAmendmentMarket(t *testing.T, dbUser string) (*rmarkets.GormRepository, *markets.Market) {
	return seedAmendmentMarketClosingAt(t, dbUser, time.Now().Add(24*time.Hour))
}

func seedAmendmentMarketClosingAt(t *testing.T, dbUser string, closesAt time.Time) (*rmarkets.GormRepository, *markets.Market) {
	t.Helper()
	db := modelstesting.NewFakeDB(t)
	user := modelstesting.GenerateUser(dbUser, 1000)
	user.UserType = string(dusers.UserTypeModerator)
	user.ModeratorStatus = string(dusers.ModeratorStatusActive)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	seed := modelstesting.GenerateMarket(44, dbUser)
	seed.QuestionTitle = "Immutable market title"
	seed.Description = "Original description"
	seed.StewardUsername = dbUser
	seed.LifecycleStatus = markets.MarketLifecyclePublished
	seed.ResolutionDateTime = closesAt
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}
	repo := rmarkets.NewGormRepository(db)
	return repo, &markets.Market{ID: seed.ID, QuestionTitle: seed.QuestionTitle, Description: seed.Description}
}

func TestProposeMarketDescriptionAmendmentAppendsWithoutChangingTitleOrOriginalDescription(t *testing.T) {
	repo, original := seedAmendmentMarket(t, "moderator")
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return activeModeratorUser(username), nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(marketsTestTime()), markets.Config{GameMode: "moderator"})

	amendment, err := service.ProposeMarketDescriptionAmendment(context.Background(), original.ID, "moderator", markets.MarketDescriptionAmendmentRequest{
		Body:         "**Clarification:** settlement source is the public result.",
		SubmitReason: "Clarify settlement source",
	})
	if err != nil {
		t.Fatalf("ProposeMarketDescriptionAmendment returned error: %v", err)
	}
	if amendment.Version != 2 || amendment.Status != markets.DescriptionAmendmentStatusPending {
		t.Fatalf("unexpected amendment: %+v", amendment)
	}

	market, err := repo.GetByID(context.Background(), original.ID)
	if err != nil {
		t.Fatalf("reload market: %v", err)
	}
	if market.QuestionTitle != original.QuestionTitle || market.Description != original.Description {
		t.Fatalf("proposal mutated market title/description: %+v", market)
	}
}

func TestProposeMarketDescriptionAmendmentRequiresCurrentActiveSteward(t *testing.T) {
	repo, original := seedAmendmentMarket(t, "moderator")
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return activeModeratorUser(username), nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(marketsTestTime()), markets.Config{GameMode: "moderator"})

	_, err := service.ProposeMarketDescriptionAmendment(context.Background(), original.ID, "othermod", markets.MarketDescriptionAmendmentRequest{Body: "Clarification"})
	if !errors.Is(err, markets.ErrUnauthorized) {
		t.Fatalf("non-steward error = %v, want ErrUnauthorized", err)
	}
}

func TestProposeMarketDescriptionAmendmentAllowsReassignedStewardNotOriginalCreator(t *testing.T) {
	repo, original := seedAmendmentMarket(t, "creator")
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return activeModeratorUser(username), nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(marketsTestTime()), markets.Config{GameMode: "moderator"})

	reassigned, err := service.ReassignMarketSteward(context.Background(), original.ID, "steward", "admin", "creator unavailable")
	if err != nil {
		t.Fatalf("ReassignMarketSteward returned error: %v", err)
	}
	if reassigned.CurrentStewardUsername() != "steward" {
		t.Fatalf("current steward = %q, want steward", reassigned.CurrentStewardUsername())
	}

	_, err = service.ProposeMarketDescriptionAmendment(context.Background(), original.ID, "creator", markets.MarketDescriptionAmendmentRequest{Body: "Creator should not be allowed after reassignment."})
	if !errors.Is(err, markets.ErrUnauthorized) {
		t.Fatalf("original creator proposal error = %v, want ErrUnauthorized", err)
	}

	amendment, err := service.ProposeMarketDescriptionAmendment(context.Background(), original.ID, "steward", markets.MarketDescriptionAmendmentRequest{Body: "Current steward clarification."})
	if err != nil {
		t.Fatalf("current steward proposal returned error: %v", err)
	}
	if amendment.CreatedBy != "steward" {
		t.Fatalf("amendment creator = %q, want steward", amendment.CreatedBy)
	}
}

func TestProposeMarketDescriptionAmendmentRejectsClosedMarket(t *testing.T) {
	now := marketsTestTime()
	repo, original := seedAmendmentMarketClosingAt(t, "moderator", now.Add(-time.Minute))
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return activeModeratorUser(username), nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{GameMode: "moderator"})

	_, err := service.ProposeMarketDescriptionAmendment(context.Background(), original.ID, "moderator", markets.MarketDescriptionAmendmentRequest{Body: "Too late."})
	if !errors.Is(err, markets.ErrInvalidState) {
		t.Fatalf("closed market proposal error = %v, want ErrInvalidState", err)
	}
}

func TestProposeMarketDescriptionAmendmentAutoApprovesWhenEnabled(t *testing.T) {
	repo, original := seedAmendmentMarket(t, "moderator")
	now := marketsTestTime()
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return activeModeratorUser(username), nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{GameMode: "moderator"})

	defaultSettings, err := service.GetMarketGovernanceSettings(context.Background())
	if err != nil {
		t.Fatalf("GetMarketGovernanceSettings returned error: %v", err)
	}
	if defaultSettings.AutoApproveDescriptionAmendments {
		t.Fatalf("auto approval should be disabled by default")
	}
	if defaultSettings.AutoApproveMarketProposals {
		t.Fatalf("market proposal auto approval should be disabled by default")
	}
	if defaultSettings.MarketGroupAnswerAdditionApprovalPolicy != markets.MarketGroupAnswerAdditionApprovalPolicyModerator {
		t.Fatalf("answer addition policy = %q, want moderator", defaultSettings.MarketGroupAnswerAdditionApprovalPolicy)
	}
	enabled := true
	saved, err := service.UpdateMarketGovernanceSettings(context.Background(), markets.MarketGovernanceSettingsUpdate{
		AutoApproveDescriptionAmendments: &enabled,
		AutoApproveMarketProposals:       &enabled,
		Version:                          defaultSettings.Version,
		UpdatedBy:                        "admin",
	})
	if err != nil {
		t.Fatalf("UpdateMarketGovernanceSettings returned error: %v", err)
	}
	if !saved.AutoApproveDescriptionAmendments {
		t.Fatalf("expected auto approval setting to be enabled")
	}
	if !saved.AutoApproveMarketProposals {
		t.Fatalf("expected market proposal auto approval setting to be enabled")
	}

	amendment, err := service.ProposeMarketDescriptionAmendment(context.Background(), original.ID, "moderator", markets.MarketDescriptionAmendmentRequest{Body: "Auto-approved clarification."})
	if err != nil {
		t.Fatalf("ProposeMarketDescriptionAmendment returned error: %v", err)
	}
	if amendment.Status != markets.DescriptionAmendmentStatusApproved {
		t.Fatalf("status = %q, want approved", amendment.Status)
	}
	if amendment.ApprovedBy != markets.DescriptionAmendmentApprovedByAuto || amendment.ApprovedAt == nil {
		t.Fatalf("unexpected auto approval audit fields: %+v", amendment)
	}
}

func TestProposeMarketDescriptionAmendmentRejectsRawHTML(t *testing.T) {
	repo, original := seedAmendmentMarket(t, "moderator")
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return activeModeratorUser(username), nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(marketsTestTime()), markets.Config{GameMode: "moderator"})

	_, err := service.ProposeMarketDescriptionAmendment(context.Background(), original.ID, "moderator", markets.MarketDescriptionAmendmentRequest{Body: "<script>alert(1)</script>"})
	if !errors.Is(err, markets.ErrInvalidInput) {
		t.Fatalf("raw HTML error = %v, want ErrInvalidInput", err)
	}
}

func TestApprovedDescriptionAmendmentsAppearInMarketDetails(t *testing.T) {
	repo, original := seedAmendmentMarket(t, "moderator")
	now := marketsTestTime()
	created, err := repo.CreateMarketDescriptionAmendment(context.Background(), markets.MarketDescriptionAmendment{
		MarketID:     original.ID,
		Body:         "Approved clarification",
		BodyFormat:   markets.DescriptionAmendmentFormatMarkdownLite,
		Status:       markets.DescriptionAmendmentStatusPending,
		CreatedBy:    "moderator",
		CreatedAt:    now,
		UpdatedAt:    now,
		SubmitReason: "test",
	})
	if err != nil {
		t.Fatalf("create amendment: %v", err)
	}
	if _, err := repo.ReviewMarketDescriptionAmendment(context.Background(), created.ID, markets.DescriptionAmendmentStatusApproved, "admin", "approved", now); err != nil {
		t.Fatalf("approve amendment: %v", err)
	}
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return activeModeratorUser(username), nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{})

	details, err := service.GetMarketDetails(context.Background(), original.ID)
	if err != nil {
		t.Fatalf("GetMarketDetails returned error: %v", err)
	}
	if len(details.DescriptionAmendments) != 1 || details.DescriptionAmendments[0].Body != "Approved clarification" {
		t.Fatalf("unexpected amendments in details: %+v", details.DescriptionAmendments)
	}
}

func TestListMarketDescriptionAmendmentsIncludesAdminReviewContext(t *testing.T) {
	repo, original := seedAmendmentMarket(t, "moderator")
	now := marketsTestTime()
	first, err := repo.CreateMarketDescriptionAmendment(context.Background(), markets.MarketDescriptionAmendment{
		MarketID:   original.ID,
		Body:       "First approved clarification",
		BodyFormat: markets.DescriptionAmendmentFormatMarkdownLite,
		Status:     markets.DescriptionAmendmentStatusPending,
		CreatedBy:  "moderator",
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		t.Fatalf("create first amendment: %v", err)
	}
	if _, err := repo.ReviewMarketDescriptionAmendment(context.Background(), first.ID, markets.DescriptionAmendmentStatusApproved, "admin", "approved", now); err != nil {
		t.Fatalf("approve first amendment: %v", err)
	}
	pending, err := repo.CreateMarketDescriptionAmendment(context.Background(), markets.MarketDescriptionAmendment{
		MarketID:   original.ID,
		Body:       "Pending clarification",
		BodyFormat: markets.DescriptionAmendmentFormatMarkdownLite,
		Status:     markets.DescriptionAmendmentStatusPending,
		CreatedBy:  "moderator",
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		t.Fatalf("create pending amendment: %v", err)
	}
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return activeModeratorUser(username), nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{})

	items, err := service.ListMarketDescriptionAmendments(context.Background(), markets.MarketDescriptionAmendmentFilters{
		Status: markets.DescriptionAmendmentStatusPending,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("ListMarketDescriptionAmendments returned error: %v", err)
	}
	if len(items) != 1 || items[0].ID != pending.ID {
		t.Fatalf("unexpected pending amendments: %+v", items)
	}
	if items[0].MarketTitle != original.QuestionTitle || items[0].MarketDescription != original.Description {
		t.Fatalf("missing market review context: %+v", items[0])
	}
	if len(items[0].PreviousApprovedAmendments) != 1 || items[0].PreviousApprovedAmendments[0].Body != "First approved clarification" {
		t.Fatalf("missing previous approved amendments: %+v", items[0].PreviousApprovedAmendments)
	}
}

func TestReviewMarketDescriptionAmendmentRequiresPendingState(t *testing.T) {
	repo, original := seedAmendmentMarket(t, "moderator")
	now := marketsTestTime()
	created, err := repo.CreateMarketDescriptionAmendment(context.Background(), markets.MarketDescriptionAmendment{
		MarketID:   original.ID,
		Body:       "Clarification",
		BodyFormat: markets.DescriptionAmendmentFormatMarkdownLite,
		Status:     markets.DescriptionAmendmentStatusPending,
		CreatedBy:  "moderator",
	})
	if err != nil {
		t.Fatalf("create amendment: %v", err)
	}
	if _, err := repo.ReviewMarketDescriptionAmendment(context.Background(), created.ID, markets.DescriptionAmendmentStatusRejected, "admin", "bad", now); err != nil {
		t.Fatalf("reject amendment: %v", err)
	}
	_, err = repo.ReviewMarketDescriptionAmendment(context.Background(), created.ID, markets.DescriptionAmendmentStatusApproved, "admin", "second review", now)
	if !errors.Is(err, markets.ErrInvalidState) {
		t.Fatalf("second review error = %v, want ErrInvalidState", err)
	}
}

var _ = models.MarketDescriptionAmendment{}

func TestReviewGroupedMarketDescriptionAmendmentsReviewsChildrenAtomically(t *testing.T) {
	repo, group := seedGroupedDescriptionAmendmentMarket(t, "moderator", []string{"Apples", "Pears"})
	now := marketsTestTime()
	first, err := repo.CreateMarketDescriptionAmendment(context.Background(), markets.MarketDescriptionAmendment{
		MarketID:     group.Members[0].MarketID,
		Body:         "Grouped clarification",
		BodyFormat:   markets.DescriptionAmendmentFormatMarkdownLite,
		Status:       markets.DescriptionAmendmentStatusPending,
		CreatedBy:    "moderator",
		SubmitReason: "same grouped proposal",
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		t.Fatalf("create first amendment: %v", err)
	}
	second, err := repo.CreateMarketDescriptionAmendment(context.Background(), markets.MarketDescriptionAmendment{
		MarketID:     group.Members[1].MarketID,
		Body:         "Grouped clarification",
		BodyFormat:   markets.DescriptionAmendmentFormatMarkdownLite,
		Status:       markets.DescriptionAmendmentStatusPending,
		CreatedBy:    "moderator",
		SubmitReason: "same grouped proposal",
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		t.Fatalf("create second amendment: %v", err)
	}
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	approved, err := service.ReviewGroupedMarketDescriptionAmendments(context.Background(), []int64{first.ID, second.ID}, markets.DescriptionAmendmentStatusApproved, "admin", "approved together")
	if err != nil {
		t.Fatalf("ReviewGroupedMarketDescriptionAmendments returned error: %v", err)
	}
	if len(approved) != 2 {
		t.Fatalf("approved count = %d, want 2", len(approved))
	}
	for _, amendment := range approved {
		if amendment.Status != markets.DescriptionAmendmentStatusApproved || amendment.ApprovedBy != "admin" || amendment.ApprovedAt == nil {
			t.Fatalf("unexpected approved amendment: %+v", amendment)
		}
	}
}

func TestReviewGroupedMarketDescriptionAmendmentsRollsBackWhenChildNotPending(t *testing.T) {
	repo, group := seedGroupedDescriptionAmendmentMarket(t, "moderator", []string{"Apples", "Pears"})
	now := marketsTestTime()
	alreadyReviewed, err := repo.CreateMarketDescriptionAmendment(context.Background(), markets.MarketDescriptionAmendment{
		MarketID:     group.Members[0].MarketID,
		Body:         "Grouped clarification",
		BodyFormat:   markets.DescriptionAmendmentFormatMarkdownLite,
		Status:       markets.DescriptionAmendmentStatusPending,
		CreatedBy:    "moderator",
		SubmitReason: "same grouped proposal",
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		t.Fatalf("create reviewed amendment: %v", err)
	}
	pending, err := repo.CreateMarketDescriptionAmendment(context.Background(), markets.MarketDescriptionAmendment{
		MarketID:     group.Members[1].MarketID,
		Body:         "Grouped clarification",
		BodyFormat:   markets.DescriptionAmendmentFormatMarkdownLite,
		Status:       markets.DescriptionAmendmentStatusPending,
		CreatedBy:    "moderator",
		SubmitReason: "same grouped proposal",
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		t.Fatalf("create pending amendment: %v", err)
	}
	if _, err := repo.ReviewMarketDescriptionAmendment(context.Background(), alreadyReviewed.ID, markets.DescriptionAmendmentStatusRejected, "admin", "reject one first", now); err != nil {
		t.Fatalf("pre-review amendment: %v", err)
	}
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	_, err = service.ReviewGroupedMarketDescriptionAmendments(context.Background(), []int64{alreadyReviewed.ID, pending.ID}, markets.DescriptionAmendmentStatusApproved, "admin", "should fail")
	if !errors.Is(err, markets.ErrInvalidState) {
		t.Fatalf("grouped review error = %v, want ErrInvalidState", err)
	}
	items, err := repo.ListMarketDescriptionAmendments(context.Background(), markets.MarketDescriptionAmendmentFilters{
		MarketID: pending.MarketID,
		Status:   markets.DescriptionAmendmentStatusPending,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("reload pending amendment: %v", err)
	}
	if len(items) != 1 || items[0].ID != pending.ID || items[0].Status != markets.DescriptionAmendmentStatusPending {
		t.Fatalf("pending child should not have changed: %+v", items)
	}
}

func TestReviewGroupedMarketDescriptionAmendmentsRejectsPartialChildSet(t *testing.T) {
	repo, group := seedGroupedDescriptionAmendmentMarket(t, "moderator", []string{"Apples", "Pears"})
	now := marketsTestTime()
	first, err := repo.CreateMarketDescriptionAmendment(context.Background(), markets.MarketDescriptionAmendment{
		MarketID:     group.Members[0].MarketID,
		Body:         "Grouped clarification",
		BodyFormat:   markets.DescriptionAmendmentFormatMarkdownLite,
		Status:       markets.DescriptionAmendmentStatusPending,
		CreatedBy:    "moderator",
		SubmitReason: "same grouped proposal",
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		t.Fatalf("create first amendment: %v", err)
	}
	second, err := repo.CreateMarketDescriptionAmendment(context.Background(), markets.MarketDescriptionAmendment{
		MarketID:     group.Members[1].MarketID,
		Body:         "Grouped clarification",
		BodyFormat:   markets.DescriptionAmendmentFormatMarkdownLite,
		Status:       markets.DescriptionAmendmentStatusPending,
		CreatedBy:    "moderator",
		SubmitReason: "same grouped proposal",
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		t.Fatalf("create second amendment: %v", err)
	}
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	_, err = service.ReviewGroupedMarketDescriptionAmendments(context.Background(), []int64{first.ID}, markets.DescriptionAmendmentStatusApproved, "admin", "partial should fail")
	if !errors.Is(err, markets.ErrInvalidState) {
		t.Fatalf("partial grouped review error = %v, want ErrInvalidState", err)
	}
	for _, item := range []struct {
		name     string
		marketID int64
		wantID   int64
	}{
		{name: "first", marketID: first.MarketID, wantID: first.ID},
		{name: "second", marketID: second.MarketID, wantID: second.ID},
	} {
		items, err := repo.ListMarketDescriptionAmendments(context.Background(), markets.MarketDescriptionAmendmentFilters{
			MarketID: item.marketID,
			Status:   markets.DescriptionAmendmentStatusPending,
			Limit:    10,
		})
		if err != nil {
			t.Fatalf("reload %s amendment: %v", item.name, err)
		}
		if len(items) != 1 || items[0].ID != item.wantID || items[0].Status != markets.DescriptionAmendmentStatusPending {
			t.Fatalf("%s child should remain pending: %+v", item.name, items)
		}
	}
}

func seedGroupedDescriptionAmendmentMarket(t *testing.T, username string, labels []string) (*rmarkets.GormRepository, *markets.MarketGroup) {
	t.Helper()
	db := modelstesting.NewFakeDB(t)
	user := modelstesting.GenerateUser(username, 1000)
	user.UserType = string(dusers.UserTypeModerator)
	user.ModeratorStatus = string(dusers.ModeratorStatusActive)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	repo := rmarkets.NewGormRepository(db)
	members := make([]markets.MarketGroupMember, 0, len(labels))
	for index, label := range labels {
		child := modelstesting.GenerateMarket(0, username)
		child.QuestionTitle = "Grouped amendment market - " + label
		child.Description = "Grouped child description"
		child.StewardUsername = username
		child.LifecycleStatus = markets.MarketLifecyclePublished
		child.ResolutionDateTime = time.Now().UTC().Add(48 * time.Hour)
		if err := db.Create(&child).Error; err != nil {
			t.Fatalf("seed child market %q: %v", label, err)
		}
		members = append(members, markets.MarketGroupMember{
			MarketID:     child.ID,
			AnswerLabel:  label,
			DisplayOrder: index,
		})
	}
	group := &markets.MarketGroup{
		QuestionTitle:      "Grouped amendment market",
		Description:        "Grouped parent description",
		LifecycleStatus:    markets.MarketLifecyclePublished,
		CreatorUsername:    username,
		StewardUsername:    username,
		ResolutionDateTime: time.Now().UTC().Add(48 * time.Hour),
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	if err := repo.CreateMarketGroup(context.Background(), group, members); err != nil {
		t.Fatalf("seed market group: %v", err)
	}
	return repo, group
}
