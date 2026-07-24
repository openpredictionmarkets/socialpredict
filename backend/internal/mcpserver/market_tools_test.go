package mcpserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	dmarkets "socialpredict/internal/domain/markets"
)

type marketToolService struct {
	discoveryToolMarketService
	detailsID    int64
	summaryID    int64
	quoteReq     dmarkets.ProbabilityProjectionRequest
	detailsErr   error
	summaryErr   error
	quoteErr     error
	detailsCalls int
	summaryCalls int
	quoteCalls   int
}

func (s *marketToolService) GetMarketDetails(_ context.Context, marketID int64) (*dmarkets.MarketOverview, error) {
	s.detailsCalls++
	s.detailsID = marketID
	if s.detailsErr != nil {
		return nil, s.detailsErr
	}
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
	s.summaryCalls++
	s.summaryID = marketID
	if s.summaryErr != nil {
		return nil, s.summaryErr
	}
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	return &dmarkets.MarketSummaryReadModel{
		Market:     &dmarkets.Market{ID: marketID, QuestionTitle: "Summary", Status: dmarkets.MarketStatusActive, CreatedAt: now, UpdatedAt: now},
		Creator:    &dmarkets.CreatorSummary{Username: "alice"},
		Accounting: dmarkets.MarketAccountingSnapshot{MarketID: marketID, LastProbability: 0.55, VolumeWithDust: 1000, MarketDust: 4, UserCount: 3},
	}, nil
}

func (s *marketToolService) ProjectProbability(_ context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	s.quoteCalls++
	s.quoteReq = req
	if s.quoteErr != nil {
		return nil, s.quoteErr
	}
	return &dmarkets.ProbabilityProjection{CurrentProbability: 0.5, ProjectedProbability: 0.62}, nil
}

type denyingResolver struct{}

func (denyingResolver) Resolve(context.Context, AccessLevel) (Principal, error) {
	return Principal{}, &ToolError{Code: "unauthorized", Message: "denied"}
}

func TestMarketToolsRejectInvalidInputBeforeServiceCall(t *testing.T) {
	tests := []struct {
		name string
		call func(*Runtime) error
	}{
		{name: "get market nonpositive id", call: func(rt *Runtime) error {
			_, _, err := rt.GetMarket(context.Background(), nil, MarketIDInput{MarketID: 0})
			return err
		}},
		{name: "get summary nonpositive id", call: func(rt *Runtime) error {
			_, _, err := rt.GetMarketSummary(context.Background(), nil, MarketIDInput{MarketID: -1})
			return err
		}},
		{name: "quote nonpositive id", call: func(rt *Runtime) error {
			_, _, err := rt.QuoteMarketProbability(context.Background(), nil, ProbabilityQuoteInput{MarketID: 0, Amount: 1, Outcome: "YES"})
			return err
		}},
		{name: "quote nonpositive amount", call: func(rt *Runtime) error {
			_, _, err := rt.QuoteMarketProbability(context.Background(), nil, ProbabilityQuoteInput{MarketID: 1, Amount: 0, Outcome: "YES"})
			return err
		}},
		{name: "quote negative amount", call: func(rt *Runtime) error {
			_, _, err := rt.QuoteMarketProbability(context.Background(), nil, ProbabilityQuoteInput{MarketID: 1, Amount: -1, Outcome: "YES"})
			return err
		}},
		{name: "quote invalid outcome", call: func(rt *Runtime) error {
			_, _, err := rt.QuoteMarketProbability(context.Background(), nil, ProbabilityQuoteInput{MarketID: 1, Amount: 1, Outcome: "MAYBE"})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &marketToolService{}
			err := tt.call(NewRuntime(svc, nil))
			assertToolErrorCode(t, err, "validation_error")
			if svc.detailsCalls != 0 || svc.summaryCalls != 0 || svc.quoteCalls != 0 {
				t.Fatalf("service called after rejected input: details=%d summary=%d quote=%d", svc.detailsCalls, svc.summaryCalls, svc.quoteCalls)
			}
		})
	}
}

func TestMarketToolsRejectDeniedAccessBeforeServiceCall(t *testing.T) {
	tests := []struct {
		name string
		call func(*Runtime) error
	}{
		{name: "get market", call: func(rt *Runtime) error {
			_, _, err := rt.GetMarket(context.Background(), nil, MarketIDInput{MarketID: 1})
			return err
		}},
		{name: "get summary", call: func(rt *Runtime) error {
			_, _, err := rt.GetMarketSummary(context.Background(), nil, MarketIDInput{MarketID: 1})
			return err
		}},
		{name: "quote", call: func(rt *Runtime) error {
			_, _, err := rt.QuoteMarketProbability(context.Background(), nil, ProbabilityQuoteInput{MarketID: 1, Amount: 1, Outcome: "YES"})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &marketToolService{}
			rt := NewRuntime(svc, nil)
			rt.resolver = denyingResolver{}
			assertToolErrorCode(t, tt.call(rt), "unauthorized")
			if svc.detailsCalls != 0 || svc.summaryCalls != 0 || svc.quoteCalls != 0 {
				t.Fatalf("service called after denied access: details=%d summary=%d quote=%d", svc.detailsCalls, svc.summaryCalls, svc.quoteCalls)
			}
		})
	}
}

func TestMarketToolsMapServiceErrors(t *testing.T) {
	tests := []struct {
		name string
		code string
		call func(*Runtime) error
	}{
		{name: "validation", code: "validation_error", call: func(rt *Runtime) error {
			_, _, err := rt.GetMarket(context.Background(), nil, MarketIDInput{MarketID: 1})
			return err
		}},
		{name: "not found", code: "not_found", call: func(rt *Runtime) error {
			_, _, err := rt.GetMarketSummary(context.Background(), nil, MarketIDInput{MarketID: 1})
			return err
		}},
		{name: "conflict", code: "conflict", call: func(rt *Runtime) error {
			_, _, err := rt.QuoteMarketProbability(context.Background(), nil, ProbabilityQuoteInput{MarketID: 1, Amount: 1, Outcome: "YES"})
			return err
		}},
		{name: "internal", code: "internal_error", call: func(rt *Runtime) error {
			_, _, err := rt.QuoteMarketProbability(context.Background(), nil, ProbabilityQuoteInput{MarketID: 1, Amount: 1, Outcome: "YES"})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &marketToolService{}
			switch tt.name {
			case "validation":
				svc.detailsErr = dmarkets.ErrInvalidInput
			case "not found":
				svc.summaryErr = dmarkets.ErrMarketNotFound
			case "conflict":
				svc.quoteErr = dmarkets.ErrInvalidState
			case "internal":
				svc.quoteErr = errors.New("database unavailable")
			}
			assertToolErrorCode(t, tt.call(NewRuntime(svc, nil)), tt.code)
		})
	}
}

func assertToolErrorCode(t *testing.T, err error, want string) {
	t.Helper()
	var toolErr *ToolError
	if !errors.As(err, &toolErr) {
		t.Fatalf("error = %v, want *ToolError", err)
	}
	if toolErr.Code != want {
		t.Fatalf("error code = %q, want %q", toolErr.Code, want)
	}
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
