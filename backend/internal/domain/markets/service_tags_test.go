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

func TestCreateMarketAssignsValidatedActiveTags(t *testing.T) {
	now := marketsTestTime()
	var assignedMarketID int64
	var assignedSlugs []string
	var assignedBy string
	var assignmentSource string
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.listMarketTagsFunc = func(_ context.Context, includeInactive bool) ([]markets.MarketTag, error) {
			if includeInactive {
				t.Fatalf("create validation should request active tags only")
			}
			return []markets.MarketTag{
				{ID: 1, Slug: "politics", DisplayName: "Politics", IsActive: true},
				{ID: 2, Slug: "sports", DisplayName: "Sports", IsActive: true},
			}, nil
		}
		repo.createFunc = func(_ context.Context, market *markets.Market) error {
			market.ID = 123
			return nil
		}
		repo.setMarketTagsFunc = func(_ context.Context, marketID int64, tagSlugs []string, by string, source string, assignedAt time.Time) ([]markets.MarketTag, error) {
			assignedMarketID = marketID
			assignedSlugs = append([]string(nil), tagSlugs...)
			assignedBy = by
			assignmentSource = source
			if !assignedAt.Equal(now) {
				t.Fatalf("assignedAt = %s, want %s", assignedAt, now)
			}
			return []markets.MarketTag{{ID: 1, Slug: "politics", DisplayName: "Politics", IsActive: true}}, nil
		}
	})
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return &dusers.PublicUser{Username: username, UserType: string(dusers.UserTypeModerator), ModeratorStatus: dusers.ModeratorStatusActive}, nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{GameMode: "moderator", MarketApprovalRequired: true})

	req := validCreateRequest(now)
	req.TagSlugs = []string{"Politics", "sports", "politics"}
	market, err := service.CreateMarket(context.Background(), req, "moderator")
	if err != nil {
		t.Fatalf("CreateMarket returned error: %v", err)
	}
	if assignedMarketID != 123 || assignedBy != "moderator" || assignmentSource != markets.MarketTagAssignmentSourceCreate {
		t.Fatalf("unexpected assignment metadata id=%d by=%q source=%q", assignedMarketID, assignedBy, assignmentSource)
	}
	if !reflect.DeepEqual(assignedSlugs, []string{"politics", "sports"}) {
		t.Fatalf("assigned slugs = %+v", assignedSlugs)
	}
	if len(market.Tags) != 1 || market.Tags[0].Slug != "politics" {
		t.Fatalf("expected returned market tags, got %+v", market.Tags)
	}
}

func TestCreateMarketRejectsUnknownOrInactiveTagBeforeCreate(t *testing.T) {
	now := marketsTestTime()
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.listMarketTagsFunc = func(context.Context, bool) ([]markets.MarketTag, error) {
			return []markets.MarketTag{{ID: 1, Slug: "active", DisplayName: "Active", IsActive: true}}, nil
		}
		repo.createFunc = func(context.Context, *markets.Market) error {
			t.Fatalf("market should not be created when tags are invalid")
			return nil
		}
	})
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return &dusers.PublicUser{Username: username, UserType: string(dusers.UserTypeModerator), ModeratorStatus: dusers.ModeratorStatusActive}, nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{GameMode: "moderator", MarketApprovalRequired: true})

	req := validCreateRequest(now)
	req.TagSlugs = []string{"missing"}
	_, err := service.CreateMarket(context.Background(), req, "moderator")
	if !errors.Is(err, markets.ErrInvalidInput) {
		t.Fatalf("CreateMarket error = %v, want ErrInvalidInput", err)
	}
}

func TestCreateAndUpdateMarketTagNormalizeAndValidate(t *testing.T) {
	now := marketsTestTime()
	var createdBy string
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.createMarketTagFunc = func(_ context.Context, tag markets.MarketTag) (*markets.MarketTag, error) {
			createdBy = tag.CreatedBy
			if tag.Slug != "public-health" || tag.DisplayName != "Public Health" || !tag.IsActive {
				t.Fatalf("unexpected create tag: %+v", tag)
			}
			tag.ID = 7
			return &tag, nil
		}
		repo.updateMarketTagFunc = func(_ context.Context, slug string, req markets.MarketTagRequest) (*markets.MarketTag, error) {
			if slug != "public-health" || req.DisplayName != "Health" || req.IsActive == nil || *req.IsActive {
				t.Fatalf("unexpected update args slug=%q req=%+v", slug, req)
			}
			return &markets.MarketTag{ID: 7, Slug: slug, DisplayName: req.DisplayName, IsActive: *req.IsActive}, nil
		}
	})
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	created, err := service.CreateMarketTag(context.Background(), markets.MarketTagRequest{DisplayName: "Public Health"}, "admin")
	if err != nil {
		t.Fatalf("CreateMarketTag returned error: %v", err)
	}
	if created.ID != 7 || createdBy != "admin" {
		t.Fatalf("unexpected created tag: %+v by=%q", created, createdBy)
	}

	active := false
	updated, err := service.UpdateMarketTag(context.Background(), "Public-Health", markets.MarketTagRequest{DisplayName: "Health", IsActive: &active})
	if err != nil {
		t.Fatalf("UpdateMarketTag returned error: %v", err)
	}
	if updated.DisplayName != "Health" || updated.IsActive {
		t.Fatalf("unexpected updated tag: %+v", updated)
	}
}

func TestUpdateMarketTagsRewritesProposedOrPublishedAssignments(t *testing.T) {
	now := marketsTestTime()
	var assignedMarketID int64
	var assignedSlugs []string
	var assignedBy string
	var assignmentSource string
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.getByIDFunc = func(_ context.Context, marketID int64) (*markets.Market, error) {
			return &markets.Market{ID: marketID, LifecycleStatus: markets.MarketLifecyclePublished, Status: markets.MarketStatusActive}, nil
		}
		repo.listMarketTagsFunc = func(_ context.Context, includeInactive bool) ([]markets.MarketTag, error) {
			if includeInactive {
				t.Fatalf("admin market tag assignment validation should request active tags only")
			}
			return []markets.MarketTag{
				{ID: 1, Slug: "policy", DisplayName: "Policy", IsActive: true},
				{ID: 2, Slug: "sports", DisplayName: "Sports", IsActive: true},
			}, nil
		}
		repo.setMarketTagsFunc = func(_ context.Context, marketID int64, tagSlugs []string, by string, source string, assignedAt time.Time) ([]markets.MarketTag, error) {
			assignedMarketID = marketID
			assignedSlugs = append([]string(nil), tagSlugs...)
			assignedBy = by
			assignmentSource = source
			if !assignedAt.Equal(now) {
				t.Fatalf("assignedAt = %s, want %s", assignedAt, now)
			}
			return []markets.MarketTag{{ID: 1, Slug: "policy", DisplayName: "Policy", IsActive: true}}, nil
		}
	})
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	market, err := service.UpdateMarketTags(context.Background(), 99, []string{"sports", "policy"}, "admin")
	if err != nil {
		t.Fatalf("UpdateMarketTags returned error: %v", err)
	}
	if assignedMarketID != 99 || assignedBy != "admin" || assignmentSource != markets.MarketTagAssignmentSourceAdmin {
		t.Fatalf("unexpected assignment metadata id=%d by=%q source=%q", assignedMarketID, assignedBy, assignmentSource)
	}
	if !reflect.DeepEqual(assignedSlugs, []string{"policy", "sports"}) {
		t.Fatalf("assigned slugs = %+v", assignedSlugs)
	}
	if len(market.Tags) != 1 || market.Tags[0].Slug != "policy" {
		t.Fatalf("expected returned market tags, got %+v", market.Tags)
	}
}

func TestUpdateMarketTagsRejectsRejectedMarkets(t *testing.T) {
	now := marketsTestTime()
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.getByIDFunc = func(_ context.Context, marketID int64) (*markets.Market, error) {
			return &markets.Market{ID: marketID, LifecycleStatus: markets.MarketLifecycleRejected, Status: markets.MarketLifecycleRejected}, nil
		}
		repo.listMarketTagsFunc = func(context.Context, bool) ([]markets.MarketTag, error) {
			t.Fatalf("rejected market should fail before tag validation")
			return nil, nil
		}
	})
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	_, err := service.UpdateMarketTags(context.Background(), 99, []string{"policy"}, "admin")
	if !errors.Is(err, markets.ErrInvalidState) {
		t.Fatalf("UpdateMarketTags error = %v, want ErrInvalidState", err)
	}
}

func TestNormalizeMarketTagSlugsRejectsTooManyOrInvalidSlugs(t *testing.T) {
	if _, err := markets.NormalizeMarketTagSlugs([]string{"valid", "not valid"}); !errors.Is(err, markets.ErrInvalidInput) {
		t.Fatalf("invalid slug error = %v, want ErrInvalidInput", err)
	}
	if _, err := markets.NormalizeMarketTagSlugs([]string{"a", "b", "c", "d", "e", "f"}); !errors.Is(err, markets.ErrInvalidInput) {
		t.Fatalf("too many tags error = %v, want ErrInvalidInput", err)
	}
}
