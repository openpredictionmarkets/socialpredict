package markets_test

import (
	"context"
	"errors"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	users "socialpredict/internal/domain/users"
)

func TestReassignMarketStewardRequiresActiveModerator(t *testing.T) {
	now := marketsTestTime()
	market := &markets.Market{ID: 10, CreatorUsername: "creator", StewardUsername: "creator", Status: markets.MarketStatusActive, LifecycleStatus: markets.MarketLifecyclePublished}
	reassigned := false
	repo := newProjectionRepo(withProjectionRepoMarket(market), func(repo *projectionRepo) {
		repo.reassignMarketStewardFunc = func(_ context.Context, marketID int64, from string, to string, actor string, reason string, changedAt time.Time) error {
			if marketID != market.ID || from != "creator" || to != "backup" || actor != "admin" || reason != "creator inactive" || !changedAt.Equal(now) {
				t.Fatalf("unexpected reassign args: id=%d from=%q to=%q actor=%q reason=%q at=%s", marketID, from, to, actor, reason, changedAt)
			}
			reassigned = true
			return nil
		}
	})
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*users.PublicUser, error) {
			return &users.PublicUser{Username: username, UserType: string(users.UserTypeModerator), ModeratorStatus: users.ModeratorStatusActive}, nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{GameMode: "moderator"})

	updated, err := service.ReassignMarketSteward(context.Background(), market.ID, "backup", "admin", "creator inactive")
	if err != nil {
		t.Fatalf("ReassignMarketSteward returned error: %v", err)
	}
	if !reassigned || updated.CurrentStewardUsername() != "backup" {
		t.Fatalf("expected steward reassignment, reassigned=%v updated=%+v", reassigned, updated)
	}
}

func TestReassignMarketStewardRejectsSuspendedModerator(t *testing.T) {
	market := &markets.Market{ID: 11, CreatorUsername: "creator", StewardUsername: "creator"}
	repo := newProjectionRepo(withProjectionRepoMarket(market), func(repo *projectionRepo) {
		repo.reassignMarketStewardFunc = func(context.Context, int64, string, string, string, string, time.Time) error {
			t.Fatalf("repository reassign should not be called")
			return nil
		}
	})
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*users.PublicUser, error) {
			return &users.PublicUser{Username: username, UserType: string(users.UserTypeModerator), ModeratorStatus: users.ModeratorStatusSuspended}, nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(marketsTestTime()), markets.Config{GameMode: "moderator"})

	_, err := service.ReassignMarketSteward(context.Background(), market.ID, "suspended", "admin", "coverage gap")
	if !errors.Is(err, markets.ErrUnauthorized) {
		t.Fatalf("ReassignMarketSteward error = %v, want ErrUnauthorized", err)
	}
}

func TestResolveMarketUsesCurrentStewardAndBlocksFormerCreator(t *testing.T) {
	market := &markets.Market{ID: 12, CreatorUsername: "creator", StewardUsername: "backup", Status: markets.MarketStatusActive, LifecycleStatus: markets.MarketLifecyclePublished, ResolutionDateTime: marketsTestTime().Add(time.Hour)}
	repo := newResolveRepo(
		withResolveRepoMarket(market),
		withResolveRepoBets([]*markets.Bet{}),
		withResolveRepoPayouts([]*markets.PayoutPosition{}),
		withResolveRepoResolve(func(_ context.Context, marketID int64, outcome string) error {
			if marketID != market.ID || outcome != "YES" {
				t.Fatalf("unexpected resolve args: id=%d outcome=%q", marketID, outcome)
			}
			return nil
		}),
	)
	usersSvc := newResolveUserService(func(service *resolveUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*users.PublicUser, error) {
			return &users.PublicUser{Username: username, UserType: string(users.UserTypeModerator), ModeratorStatus: users.ModeratorStatusActive}, nil
		}
	})
	service := markets.NewService(repo, usersSvc, newNopClock(marketsTestTime()), markets.Config{GameMode: "moderator"})

	if err := service.ResolveMarket(context.Background(), market.ID, "YES", "creator"); !errors.Is(err, markets.ErrUnauthorized) {
		t.Fatalf("creator resolve error = %v, want ErrUnauthorized", err)
	}
	if err := service.ResolveMarket(context.Background(), market.ID, "YES", "backup"); err != nil {
		t.Fatalf("steward ResolveMarket returned error: %v", err)
	}
}

func TestResolveMarketBlocksSuspendedSteward(t *testing.T) {
	market := &markets.Market{ID: 13, CreatorUsername: "creator", StewardUsername: "backup", Status: markets.MarketStatusActive, LifecycleStatus: markets.MarketLifecyclePublished, ResolutionDateTime: marketsTestTime().Add(time.Hour)}
	repo := newResolveRepo(
		withResolveRepoMarket(market),
		withResolveRepoResolve(func(context.Context, int64, string) error {
			t.Fatalf("repository resolve should not be called")
			return nil
		}),
	)
	usersSvc := newResolveUserService(func(service *resolveUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*users.PublicUser, error) {
			return &users.PublicUser{Username: username, UserType: string(users.UserTypeModerator), ModeratorStatus: users.ModeratorStatusSuspended}, nil
		}
	})
	service := markets.NewService(repo, usersSvc, newNopClock(marketsTestTime()), markets.Config{GameMode: "moderator"})

	if err := service.ResolveMarket(context.Background(), market.ID, "YES", "backup"); !errors.Is(err, markets.ErrUnauthorized) {
		t.Fatalf("suspended steward resolve error = %v, want ErrUnauthorized", err)
	}
}
