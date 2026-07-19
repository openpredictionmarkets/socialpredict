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

type GetMarketUserPositionOutput struct {
	Position UserPositionOutput `json:"position"`
}

func (rt *Runtime) ListMarketBets(ctx context.Context, _ *mcp.CallToolRequest, in MarketActivityInput) (*mcp.CallToolResult, PageItems[BetOutput], error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, PageItems[BetOutput]{}, err
	}
	marketID, err := NormalizeID(in.MarketID, "marketId")
	if err != nil {
		return nil, PageItems[BetOutput]{}, err
	}
	page := NormalizePage(in.Limit, in.Offset)
	rows, err := rt.markets.GetMarketBetsPage(ctx, marketID, page)
	if err != nil {
		return nil, PageItems[BetOutput]{}, MapError(err)
	}
	out := make([]BetOutput, 0, len(rows))
	for _, row := range rows {
		out = append(out, BetOutputFromDomain(row))
	}
	return nil, NewPageItems(out, page.Limit, page.Offset, nil), nil
}

func (rt *Runtime) ListMarketPositions(ctx context.Context, _ *mcp.CallToolRequest, in MarketActivityInput) (*mcp.CallToolResult, PageItems[UserPositionOutput], error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, PageItems[UserPositionOutput]{}, err
	}
	marketID, err := NormalizeID(in.MarketID, "marketId")
	if err != nil {
		return nil, PageItems[UserPositionOutput]{}, err
	}
	page := NormalizePage(in.Limit, in.Offset)
	rows, err := rt.markets.GetMarketPositionsPage(ctx, marketID, page)
	if err != nil {
		return nil, PageItems[UserPositionOutput]{}, MapError(err)
	}
	out := make([]UserPositionOutput, 0, len(rows))
	for _, row := range rows {
		out = append(out, UserPositionOutputFromDomain(row))
	}
	return nil, NewPageItems(out, page.Limit, page.Offset, nil), nil
}

func (rt *Runtime) GetMarketUserPosition(ctx context.Context, _ *mcp.CallToolRequest, in MarketUserPositionInput) (*mcp.CallToolResult, GetMarketUserPositionOutput, error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, GetMarketUserPositionOutput{}, err
	}
	marketID, err := NormalizeID(in.MarketID, "marketId")
	if err != nil {
		return nil, GetMarketUserPositionOutput{}, err
	}
	username := strings.TrimSpace(in.Username)
	if username == "" {
		return nil, GetMarketUserPositionOutput{}, &ToolError{Code: "validation_error", Message: "username is required"}
	}
	position, err := rt.markets.GetUserPositionInMarket(ctx, marketID, username)
	if err != nil {
		return nil, GetMarketUserPositionOutput{}, MapError(err)
	}
	return nil, GetMarketUserPositionOutput{Position: UserPositionOutputFromDomain(position)}, nil
}

func (rt *Runtime) GetMarketLeaderboard(ctx context.Context, _ *mcp.CallToolRequest, in MarketActivityInput) (*mcp.CallToolResult, PageItems[LeaderboardRowOutput], error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, PageItems[LeaderboardRowOutput]{}, err
	}
	marketID, err := NormalizeID(in.MarketID, "marketId")
	if err != nil {
		return nil, PageItems[LeaderboardRowOutput]{}, err
	}
	page := NormalizePage(in.Limit, in.Offset)
	rows, err := rt.markets.GetMarketLeaderboard(ctx, marketID, page)
	if err != nil {
		return nil, PageItems[LeaderboardRowOutput]{}, MapError(err)
	}
	out := make([]LeaderboardRowOutput, 0, len(rows))
	for _, row := range rows {
		out = append(out, LeaderboardRowOutputFromDomain(row))
	}
	return nil, NewPageItems(out, page.Limit, page.Offset, nil), nil
}

func (rt *Runtime) ListMarketGroupBets(ctx context.Context, _ *mcp.CallToolRequest, in MarketGroupActivityInput) (*mcp.CallToolResult, PageItems[MarketGroupBetOutput], error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, PageItems[MarketGroupBetOutput]{}, err
	}
	groupID, err := NormalizeID(in.GroupID, "groupId")
	if err != nil {
		return nil, PageItems[MarketGroupBetOutput]{}, err
	}
	page := NormalizePage(in.Limit, in.Offset)
	result, err := rt.markets.GetMarketGroupBetsPage(ctx, groupID, page)
	if err != nil {
		return nil, PageItems[MarketGroupBetOutput]{}, MapError(err)
	}
	rows := []MarketGroupBetOutput{}
	total := 0
	if result != nil {
		total = result.Total
		for _, row := range result.Bets {
			rows = append(rows, MarketGroupBetOutputFromDomain(row))
		}
	}
	return nil, NewPageItems(rows, page.Limit, page.Offset, &total), nil
}

func (rt *Runtime) ListMarketGroupPositions(ctx context.Context, _ *mcp.CallToolRequest, in MarketGroupActivityInput) (*mcp.CallToolResult, PageItems[MarketGroupPositionOutput], error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, PageItems[MarketGroupPositionOutput]{}, err
	}
	groupID, err := NormalizeID(in.GroupID, "groupId")
	if err != nil {
		return nil, PageItems[MarketGroupPositionOutput]{}, err
	}
	page := NormalizePage(in.Limit, in.Offset)
	result, err := rt.markets.GetMarketGroupPositionsPage(ctx, groupID, page)
	if err != nil {
		return nil, PageItems[MarketGroupPositionOutput]{}, MapError(err)
	}
	rows := []MarketGroupPositionOutput{}
	total := 0
	if result != nil {
		total = result.Total
		for _, row := range result.Positions {
			rows = append(rows, MarketGroupPositionOutputFromDomain(row))
		}
	}
	return nil, NewPageItems(rows, page.Limit, page.Offset, &total), nil
}

func (rt *Runtime) GetMarketGroupLeaderboard(ctx context.Context, _ *mcp.CallToolRequest, in MarketGroupActivityInput) (*mcp.CallToolResult, PageItems[MarketGroupLeaderboardRowOutput], error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, PageItems[MarketGroupLeaderboardRowOutput]{}, err
	}
	groupID, err := NormalizeID(in.GroupID, "groupId")
	if err != nil {
		return nil, PageItems[MarketGroupLeaderboardRowOutput]{}, err
	}
	page := NormalizePage(in.Limit, in.Offset)
	result, err := rt.markets.GetMarketGroupLeaderboardPage(ctx, groupID, page)
	if err != nil {
		return nil, PageItems[MarketGroupLeaderboardRowOutput]{}, MapError(err)
	}
	rows := []MarketGroupLeaderboardRowOutput{}
	total := 0
	if result != nil {
		total = result.Total
		for _, row := range result.Leaderboard {
			rows = append(rows, MarketGroupLeaderboardRowOutputFromDomain(row))
		}
	}
	return nil, NewPageItems(rows, page.Limit, page.Offset, &total), nil
}
