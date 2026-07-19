package mcpserver

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	dmarkets "socialpredict/internal/domain/markets"
)

func (rt *Runtime) GetMarket(ctx context.Context, _ *mcp.CallToolRequest, in MarketIDInput) (*mcp.CallToolResult, MarketDetailsOutput, error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, MarketDetailsOutput{}, err
	}
	marketID, err := NormalizeID(in.MarketID, "marketId")
	if err != nil {
		return nil, MarketDetailsOutput{}, err
	}
	overview, err := rt.markets.GetMarketDetails(ctx, marketID)
	if err != nil {
		return nil, MarketDetailsOutput{}, MapError(err)
	}
	return nil, MarketDetailsOutputFromDomain(overview), nil
}

func (rt *Runtime) GetMarketSummary(ctx context.Context, _ *mcp.CallToolRequest, in MarketIDInput) (*mcp.CallToolResult, MarketSummaryOutput, error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, MarketSummaryOutput{}, err
	}
	marketID, err := NormalizeID(in.MarketID, "marketId")
	if err != nil {
		return nil, MarketSummaryOutput{}, err
	}
	summary, err := rt.markets.GetMarketSummaryReadModel(ctx, marketID)
	if err != nil {
		return nil, MarketSummaryOutput{}, MapError(err)
	}
	return nil, MarketSummaryOutputFromDomain(summary), nil
}

func (rt *Runtime) QuoteMarketProbability(ctx context.Context, _ *mcp.CallToolRequest, in ProbabilityQuoteInput) (*mcp.CallToolResult, ProbabilityQuoteOutput, error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, ProbabilityQuoteOutput{}, err
	}
	marketID, err := NormalizeID(in.MarketID, "marketId")
	if err != nil {
		return nil, ProbabilityQuoteOutput{}, err
	}
	if in.Amount <= 0 {
		return nil, ProbabilityQuoteOutput{}, &ToolError{Code: "validation_error", Message: "amount must be positive"}
	}
	outcome, err := NormalizeOutcome(in.Outcome)
	if err != nil {
		return nil, ProbabilityQuoteOutput{}, err
	}
	projection, err := rt.markets.ProjectProbability(ctx, dmarkets.ProbabilityProjectionRequest{MarketID: marketID, Amount: in.Amount, Outcome: outcome})
	if err != nil {
		return nil, ProbabilityQuoteOutput{}, MapError(err)
	}
	return nil, ProbabilityQuoteOutput{
		MarketID:             marketID,
		CurrentProbability:   projection.CurrentProbability,
		ProjectedProbability: projection.ProjectedProbability,
		Amount:               in.Amount,
		Outcome:              strings.ToUpper(outcome),
	}, nil
}
