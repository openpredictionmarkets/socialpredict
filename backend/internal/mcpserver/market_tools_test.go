package mcpserver

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	dmarkets "socialpredict/internal/domain/markets"
)

type marketToolService struct {
	discoveryToolMarketService
	detailsID int64
	summaryID int64
	quoteReq  dmarkets.ProbabilityProjectionRequest
}

func (s *marketToolService) GetMarketDetails(_ context.Context, marketID int64) (*dmarkets.MarketOverview, error) {
	s.detailsID = marketID
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	return &dmarkets.MarketOverview{
		Market:          &dmarkets.Market{ID: marketID, QuestionTitle: "Detail", Status: dmarkets.MarketStatusActive, CreatedAt: now, UpdatedAt: now},
		Creator:         &dmarkets.CreatorSummary{Username: "alice"},
		LastProbability: 0.5,
		TotalVolume:     900,
		MarketDust:      9,
	}, nil
}

func (s *marketToolService) GetMarketSummaryReadModel(_ context.Context, marketID int64) (*dmarkets.MarketSummaryReadModel, error) {
	s.summaryID = marketID
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	return &dmarkets.MarketSummaryReadModel{
		Market:     &dmarkets.Market{ID: marketID, QuestionTitle: "Summary", Status: dmarkets.MarketStatusActive, CreatedAt: now, UpdatedAt: now},
		Creator:    &dmarkets.CreatorSummary{Username: "alice"},
		Accounting: dmarkets.MarketAccountingSnapshot{MarketID: marketID, LastProbability: 0.55, VolumeWithDust: 1000, MarketDust: 4, UserCount: 3},
	}, nil
}

func (s *marketToolService) ProjectProbability(_ context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	s.quoteReq = req
	return &dmarkets.ProbabilityProjection{CurrentProbability: 0.5, ProjectedProbability: 0.62}, nil
}

func TestGetMarketUsesMarketDetails(t *testing.T) {
	svc := &marketToolService{}
	rt := NewRuntime(svc, nil)
	_, got, err := rt.GetMarket(context.Background(), &mcp.CallToolRequest{}, MarketIDInput{MarketID: 9})
	if err != nil {
		t.Fatalf("GetMarket returned error: %v", err)
	}
	if svc.detailsID != 9 || got.Market.ID != 9 || got.TotalVolume != 900 {
		t.Fatalf("details output = %#v detailsID=%d", got, svc.detailsID)
	}
}

func TestGetMarketSummaryUsesReadModel(t *testing.T) {
	svc := &marketToolService{}
	rt := NewRuntime(svc, nil)
	_, got, err := rt.GetMarketSummary(context.Background(), &mcp.CallToolRequest{}, MarketIDInput{MarketID: 10})
	if err != nil {
		t.Fatalf("GetMarketSummary returned error: %v", err)
	}
	if svc.summaryID != 10 || got.Market.ID != 10 || got.TotalVolume != 1000 || got.MarketDust != 4 {
		t.Fatalf("summary output = %#v summaryID=%d", got, svc.summaryID)
	}
}

func TestQuoteMarketProbabilityNormalizesOutcome(t *testing.T) {
	svc := &marketToolService{}
	rt := NewRuntime(svc, nil)
	_, got, err := rt.QuoteMarketProbability(context.Background(), &mcp.CallToolRequest{}, ProbabilityQuoteInput{MarketID: 11, Amount: 25, Outcome: " yes "})
	if err != nil {
		t.Fatalf("QuoteMarketProbability returned error: %v", err)
	}
	if svc.quoteReq.MarketID != 11 || svc.quoteReq.Amount != 25 || svc.quoteReq.Outcome != "YES" {
		t.Fatalf("quote request = %#v", svc.quoteReq)
	}
	if got.ProjectedProbability != 0.62 || got.Outcome != "YES" {
		t.Fatalf("quote output = %#v", got)
	}
}
