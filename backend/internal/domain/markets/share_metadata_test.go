package markets_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
)

func TestGetShareMetadataBuildsPublicMarketReadModel(t *testing.T) {
	now := marketsTestTime()
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.getPublicMarketFunc = func(_ context.Context, marketID int64) (*markets.PublicMarket, error) {
			return &markets.PublicMarket{
				ID:                 marketID,
				QuestionTitle:      "Will open graph previews work?",
				Description:        "A public market description for sharing.",
				ResolutionDateTime: now.Add(24 * time.Hour),
				CreatorUsername:    "creator",
				LifecycleStatus:    markets.MarketLifecyclePublished,
			}, nil
		}
	})
	service := markets.NewService(repo, nil, newFixedClock(now), markets.Config{})

	metadata, err := service.GetShareMetadata(context.Background(), 42, markets.ShareMetadataConfig{
		PublicBaseURL:   "https://kconfs.com",
		DefaultImageURL: "/share-card.png",
		SiteName:        "KConfs",
	})
	if err != nil {
		t.Fatalf("GetShareMetadata returned error: %v", err)
	}

	if metadata.Title != "Will open graph previews work? | KConfs" {
		t.Fatalf("Title = %q", metadata.Title)
	}
	if metadata.CanonicalURL != "https://kconfs.com/markets/42" {
		t.Fatalf("CanonicalURL = %q", metadata.CanonicalURL)
	}
	if metadata.ImageURL != "https://kconfs.com/share-card.png" {
		t.Fatalf("ImageURL = %q", metadata.ImageURL)
	}
	if metadata.PublicStatus != markets.MarketStatusActive || !metadata.Shareable {
		t.Fatalf("unexpected status/shareable: %+v", metadata)
	}
}

func TestGetShareMetadataRejectsNonPublicMarketStates(t *testing.T) {
	now := marketsTestTime()
	for _, lifecycle := range []string{
		markets.MarketLifecycleProposed,
		markets.MarketLifecycleRejected,
		markets.MarketLifecycleCancelled,
	} {
		t.Run(lifecycle, func(t *testing.T) {
			repo := newProjectionRepo(func(repo *projectionRepo) {
				repo.getPublicMarketFunc = func(_ context.Context, marketID int64) (*markets.PublicMarket, error) {
					return &markets.PublicMarket{
						ID:                 marketID,
						QuestionTitle:      "Private workflow market",
						ResolutionDateTime: now.Add(time.Hour),
						LifecycleStatus:    lifecycle,
					}, nil
				}
			})
			service := markets.NewService(repo, nil, newFixedClock(now), markets.Config{})

			_, err := service.GetShareMetadata(context.Background(), 51, markets.ShareMetadataConfig{})
			if !errors.Is(err, markets.ErrMarketNotFound) {
				t.Fatalf("GetShareMetadata error = %v, want ErrMarketNotFound", err)
			}
		})
	}
}

func TestGetShareMetadataBoundsDescription(t *testing.T) {
	now := marketsTestTime()
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.getPublicMarketFunc = func(_ context.Context, marketID int64) (*markets.PublicMarket, error) {
			return &markets.PublicMarket{
				ID:                 marketID,
				QuestionTitle:      "A market",
				Description:        strings.Repeat("word ", 100),
				ResolutionDateTime: now.Add(time.Hour),
				LifecycleStatus:    markets.MarketLifecyclePublished,
			}, nil
		}
	})
	service := markets.NewService(repo, nil, newFixedClock(now), markets.Config{})

	metadata, err := service.GetShareMetadata(context.Background(), 9, markets.ShareMetadataConfig{})
	if err != nil {
		t.Fatalf("GetShareMetadata returned error: %v", err)
	}
	if len([]rune(metadata.Description)) > 220 {
		t.Fatalf("description too long: %d", len([]rune(metadata.Description)))
	}
}

func TestGetShareMetadataUsesConfiguredDefaultDescription(t *testing.T) {
	now := marketsTestTime()
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.getPublicMarketFunc = func(_ context.Context, marketID int64) (*markets.PublicMarket, error) {
			return &markets.PublicMarket{
				ID:                 marketID,
				QuestionTitle:      "A market",
				Description:        "",
				ResolutionDateTime: now.Add(time.Hour),
				LifecycleStatus:    markets.MarketLifecyclePublished,
			}, nil
		}
	})
	service := markets.NewService(repo, nil, newFixedClock(now), markets.Config{})

	metadata, err := service.GetShareMetadata(context.Background(), 10, markets.ShareMetadataConfig{
		DefaultDescription: "CMS-managed description for public share previews.",
	})
	if err != nil {
		t.Fatalf("GetShareMetadata returned error: %v", err)
	}
	if metadata.Description != "CMS-managed description for public share previews." {
		t.Fatalf("Description = %q", metadata.Description)
	}
}

func TestGetShareMetadataOmitsImageWhenDisabled(t *testing.T) {
	now := marketsTestTime()
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.getPublicMarketFunc = func(_ context.Context, marketID int64) (*markets.PublicMarket, error) {
			return &markets.PublicMarket{
				ID:                 marketID,
				QuestionTitle:      "A market",
				Description:        "A public market description.",
				ResolutionDateTime: now.Add(time.Hour),
				LifecycleStatus:    markets.MarketLifecyclePublished,
			}, nil
		}
	})
	service := markets.NewService(repo, nil, newFixedClock(now), markets.Config{})

	metadata, err := service.GetShareMetadata(context.Background(), 11, markets.ShareMetadataConfig{
		PublicBaseURL:   "https://kconfs.com",
		DefaultImageURL: "/share-card.png",
		DisableImage:    true,
	})
	if err != nil {
		t.Fatalf("GetShareMetadata returned error: %v", err)
	}
	if metadata.ImageURL != "" {
		t.Fatalf("ImageURL = %q, want empty", metadata.ImageURL)
	}
}
