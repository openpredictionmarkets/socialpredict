package marketpublicresponse

import (
	"context"
	"errors"
	"testing"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
)

type marketServiceStub struct {
	getPublicMarketFunc func(ctx context.Context, marketID int64) (*dmarkets.PublicMarket, error)
}

func (m marketServiceStub) CreateMarket(context.Context, dmarkets.MarketCreateRequest, string) (*dmarkets.Market, error) {
	panic("not implemented")
}
func (m marketServiceStub) SetCustomLabels(context.Context, int64, string, string) error {
	panic("not implemented")
}
func (m marketServiceStub) GetMarket(context.Context, int64) (*dmarkets.Market, error) {
	panic("not implemented")
}
func (m marketServiceStub) ListMarkets(context.Context, dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	panic("not implemented")
}
func (m marketServiceStub) SearchMarkets(context.Context, string, dmarkets.SearchFilters) (*dmarkets.SearchResults, error) {
	panic("not implemented")
}
func (m marketServiceStub) ResolveMarket(context.Context, int64, string, string) error {
	panic("not implemented")
}
func (m marketServiceStub) ListByStatus(context.Context, string, dmarkets.Page) ([]*dmarkets.Market, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetMarketLeaderboard(context.Context, int64, dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
	panic("not implemented")
}
func (m marketServiceStub) ProjectProbability(context.Context, dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetMarketDetails(context.Context, int64) (*dmarkets.MarketOverview, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetMarketBets(context.Context, int64) ([]*dmarkets.BetDisplayInfo, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetMarketPositions(context.Context, int64) (dmarkets.MarketPositions, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetUserPositionInMarket(context.Context, int64, string) (*dmarkets.UserPosition, error) {
	panic("not implemented")
}
func (m marketServiceStub) CalculateMarketVolume(context.Context, int64) (int64, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetPublicMarket(ctx context.Context, marketID int64) (*dmarkets.PublicMarket, error) {
	if m.getPublicMarketFunc == nil {
		panic("GetPublicMarket called without stub")
	}
	return m.getPublicMarketFunc(ctx, marketID)
}

func TestGetPublicResponseMarketValidation(t *testing.T) {
	_, err := GetPublicResponseMarket(context.Background(), nil, 1)
	if err == nil || err.Error() != "market service is nil" {
		t.Fatalf("expected nil service error, got %v", err)
	}
}

func TestGetPublicResponseMarketErrorPropagates(t *testing.T) {
	wantErr := errors.New("boom")
	svc := marketServiceStub{
		getPublicMarketFunc: func(context.Context, int64) (*dmarkets.PublicMarket, error) {
			return nil, wantErr
		},
	}

	_, err := GetPublicResponseMarket(context.Background(), svc, 5)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}

func TestGetPublicResponseMarketMapping(t *testing.T) {
	now := time.Date(2025, 6, 7, 8, 9, 10, 0, time.UTC)
	final := now.Add(24 * time.Hour)
	svc := marketServiceStub{
		getPublicMarketFunc: func(ctx context.Context, marketID int64) (*dmarkets.PublicMarket, error) {
			if marketID != 42 {
				t.Fatalf("expected marketID 42, got %d", marketID)
			}
			return &dmarkets.PublicMarket{
				ID:                      42,
				QuestionTitle:           "Will it rain?",
				Description:             "Weather forecast",
				OutcomeType:             "BINARY",
				ResolutionDateTime:      now,
				FinalResolutionDateTime: final,
				UTCOffset:               -5,
				IsResolved:              true,
				ResolutionResult:        "YES",
				InitialProbability:      0.6,
				CreatorUsername:         "tester",
				CreatedAt:               now.Add(-time.Hour),
				YesLabel:                "Wet",
				NoLabel:                 "Dry",
			}, nil
		},
	}

	resp, err := GetPublicResponseMarket(context.Background(), svc, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatalf("expected response, got nil")
	}

	if resp.ID != 42 ||
		resp.QuestionTitle != "Will it rain?" ||
		resp.Description != "Weather forecast" ||
		resp.OutcomeType != "BINARY" ||
		!resp.ResolutionDateTime.Equal(now) ||
		!resp.FinalResolutionDateTime.Equal(final) ||
		resp.UTCOffset != -5 ||
		!resp.IsResolved ||
		resp.ResolutionResult != "YES" ||
		resp.InitialProbability != 0.6 ||
		resp.CreatorUsername != "tester" ||
		!resp.CreatedAt.Equal(now.Add(-time.Hour)) ||
		resp.YesLabel != "Wet" ||
		resp.NoLabel != "Dry" {
		t.Fatalf("unexpected mapping result: %+v", resp)
	}
}
