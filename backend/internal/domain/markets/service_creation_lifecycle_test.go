package markets_test

import (
	"context"
	"errors"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
)

func validCreateRequest(now time.Time) markets.MarketCreateRequest {
	return markets.MarketCreateRequest{
		QuestionTitle:      "Will the launch succeed?",
		Description:        "Moderator lifecycle test market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(24 * time.Hour),
	}
}

func TestCreateMarketOpenModeCreatesPublishedActiveMarket(t *testing.T) {
	now := marketsTestTime()
	var created *markets.Market
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.createFunc = func(_ context.Context, market *markets.Market) error {
			created = market
			market.ID = 101
			return nil
		}
	})
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	market, err := service.CreateMarket(context.Background(), validCreateRequest(now), "alice")
	if err != nil {
		t.Fatalf("CreateMarket returned error: %v", err)
	}
	if market != created || market.Status != markets.MarketStatusActive || market.LifecycleStatus != markets.MarketLifecyclePublished {
		t.Fatalf("unexpected created market: %+v", market)
	}
}

func TestCreateMarketModeratorModeCreatesProposedMarketForActiveModerator(t *testing.T) {
	now := marketsTestTime()
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.createFunc = func(_ context.Context, market *markets.Market) error {
			market.ID = 102
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
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{
		GameMode:               "moderator",
		MarketApprovalRequired: true,
		CreateMarketCost:       10,
	})

	market, err := service.CreateMarket(context.Background(), validCreateRequest(now), "moderator")
	if err != nil {
		t.Fatalf("CreateMarket returned error: %v", err)
	}
	if market.Status != markets.MarketLifecycleProposed || market.LifecycleStatus != markets.MarketLifecycleProposed {
		t.Fatalf("expected proposed lifecycle in moderator mode, got %+v", market)
	}
	if market.ProposalCost != 10 {
		t.Fatalf("proposal cost = %d, want 10", market.ProposalCost)
	}
	if market.IsTradableAt(now) {
		t.Fatalf("proposed market must not be tradable")
	}
}

func TestCreateMarketModeratorModeRejectsRegularAndSuspendedModerators(t *testing.T) {
	now := marketsTestTime()
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.createFunc = func(context.Context, *markets.Market) error {
			t.Fatalf("repository create should not be called")
			return nil
		}
	})

	for _, tt := range []struct {
		name string
		user *dusers.PublicUser
	}{
		{
			name: "regular user",
			user: &dusers.PublicUser{Username: "regular", UserType: string(dusers.UserTypeRegular)},
		},
		{
			name: "suspended moderator",
			user: &dusers.PublicUser{Username: "moderator", UserType: string(dusers.UserTypeModerator), ModeratorStatus: dusers.ModeratorStatusSuspended},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			usersSvc := newNoopUserService(func(service *noopUserService) {
				service.getPublicUserFunc = func(context.Context, string) (*dusers.PublicUser, error) {
					return tt.user, nil
				}
			})
			service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{GameMode: "moderator", MarketApprovalRequired: true})

			_, err := service.CreateMarket(context.Background(), validCreateRequest(now), tt.user.Username)
			if !errors.Is(err, markets.ErrUnauthorized) {
				t.Fatalf("CreateMarket error = %v, want ErrUnauthorized", err)
			}
		})
	}
}

func TestMarketLifecycleTradability(t *testing.T) {
	now := marketsTestTime()
	for _, tt := range []struct {
		name     string
		market   *markets.Market
		tradable bool
	}{
		{"published active", &markets.Market{Status: markets.MarketStatusActive, LifecycleStatus: markets.MarketLifecyclePublished, ResolutionDateTime: now.Add(time.Hour)}, true},
		{"proposed", &markets.Market{Status: markets.MarketLifecycleProposed, LifecycleStatus: markets.MarketLifecycleProposed, ResolutionDateTime: now.Add(time.Hour)}, false},
		{"rejected", &markets.Market{Status: markets.MarketLifecycleRejected, LifecycleStatus: markets.MarketLifecycleRejected, ResolutionDateTime: now.Add(time.Hour)}, false},
		{"cancelled", &markets.Market{Status: markets.MarketLifecycleCancelled, LifecycleStatus: markets.MarketLifecycleCancelled, ResolutionDateTime: now.Add(time.Hour)}, false},
		{"closed", &markets.Market{Status: markets.MarketStatusClosed, LifecycleStatus: markets.MarketLifecyclePublished, ResolutionDateTime: now.Add(-time.Hour)}, false},
		{"resolved", &markets.Market{Status: markets.MarketStatusResolved, LifecycleStatus: markets.MarketLifecycleResolved, ResolutionDateTime: now.Add(-time.Hour)}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.market.IsTradableAt(now); got != tt.tradable {
				t.Fatalf("IsTradableAt = %v, want %v", got, tt.tradable)
			}
		})
	}
}

func TestMarketLifecycleTransitionsUpdateStatus(t *testing.T) {
	now := marketsTestTime()
	market := &markets.Market{ResolutionDateTime: now.Add(24 * time.Hour)}

	if err := market.MarkProposed(now); err != nil {
		t.Fatalf("MarkProposed returned error: %v", err)
	}
	if market.Status != markets.MarketLifecycleProposed || market.LifecycleStatus != markets.MarketLifecycleProposed {
		t.Fatalf("unexpected proposed transition: %+v", market)
	}

	if err := market.Publish(now.Add(time.Minute)); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}
	if market.Status != markets.MarketStatusActive || market.LifecycleStatus != markets.MarketLifecyclePublished {
		t.Fatalf("unexpected publish transition: %+v", market)
	}

	if err := market.Reject(now.Add(2 * time.Minute)); err != nil {
		t.Fatalf("Reject returned error: %v", err)
	}
	if market.Status != markets.MarketLifecycleRejected || market.LifecycleStatus != markets.MarketLifecycleRejected {
		t.Fatalf("unexpected reject transition: %+v", market)
	}

	if err := market.ApplyLifecycleStatus("not-a-state", now); !errors.Is(err, markets.ErrInvalidState) {
		t.Fatalf("ApplyLifecycleStatus error = %v, want ErrInvalidState", err)
	}
}
