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
	seed.ResolutionDateTime = time.Now().Add(24 * time.Hour)
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
