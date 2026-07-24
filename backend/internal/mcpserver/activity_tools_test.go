package mcpserver

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	dmarkets "socialpredict/internal/domain/markets"
)

type activityToolService struct {
	marketToolService
	lastMarketPage dmarkets.Page
	lastGroupPage  dmarkets.Page
	nilGroupResult bool
}

func (s *activityToolService) GetMarketBetsPage(_ context.Context, marketID int64, page dmarkets.Page) ([]*dmarkets.BetDisplayInfo, error) {
	s.lastMarketPage = page
	return []*dmarkets.BetDisplayInfo{{Username: "alice", Outcome: "YES", Amount: 10, Probability: 0.51, PlacedAt: time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)}}, nil
}

func (s *activityToolService) GetMarketPositionsPage(_ context.Context, marketID int64, page dmarkets.Page) (dmarkets.MarketPositions, error) {
	s.lastMarketPage = page
	return dmarkets.MarketPositions{{Username: "alice", MarketID: marketID, YesSharesOwned: 3, Value: 30}}, nil
}

func (s *activityToolService) GetUserPositionInMarket(_ context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	return &dmarkets.UserPosition{Username: username, MarketID: marketID, NoSharesOwned: 4, Value: 40}, nil
}

func (s *activityToolService) GetMarketLeaderboard(_ context.Context, marketID int64, page dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
	s.lastMarketPage = page
	return []*dmarkets.LeaderboardRow{{Username: "alice", Profit: 8, Rank: 1}}, nil
}

func (s *activityToolService) GetMarketGroupBetsPage(_ context.Context, groupID int64, page dmarkets.Page) (*dmarkets.MarketGroupBetsPage, error) {
	s.lastGroupPage = page
	if s.nilGroupResult {
		return nil, nil
	}
	return &dmarkets.MarketGroupBetsPage{GroupID: groupID, Bets: []*dmarkets.MarketGroupBetDisplayInfo{{AnswerMarketID: 9, AnswerLabel: "A", Username: "alice", Outcome: "YES"}}, Total: 33}, nil
}

func (s *activityToolService) GetMarketGroupPositionsPage(_ context.Context, groupID int64, page dmarkets.Page) (*dmarkets.MarketGroupPositionsPage, error) {
	s.lastGroupPage = page
	if s.nilGroupResult {
		return nil, nil
	}
	return &dmarkets.MarketGroupPositionsPage{GroupID: groupID, Positions: []*dmarkets.MarketGroupPositionRow{{Username: "alice", YesSharesOwned: 1}}, Total: 44}, nil
}

func (s *activityToolService) GetMarketGroupLeaderboardPage(_ context.Context, groupID int64, page dmarkets.Page) (*dmarkets.MarketGroupLeaderboardPage, error) {
	s.lastGroupPage = page
	if s.nilGroupResult {
		return nil, nil
	}
	return &dmarkets.MarketGroupLeaderboardPage{GroupID: groupID, Leaderboard: []*dmarkets.MarketGroupLeaderboardRow{{Username: "alice", Profit: 10, Rank: 1}}, Total: 55}, nil
}

func TestListMarketBetsUsesNormalizedPaginationWithoutTotal(t *testing.T) {
	svc := &activityToolService{}
	rt := NewRuntime(svc, nil)
	_, got, err := rt.ListMarketBets(context.Background(), &mcp.CallToolRequest{}, MarketActivityInput{MarketID: 4, Limit: 500, Offset: -9})
	if err != nil {
		t.Fatalf("ListMarketBets returned error: %v", err)
	}
	if svc.lastMarketPage.Limit != 100 || svc.lastMarketPage.Offset != 0 {
		t.Fatalf("page = %#v", svc.lastMarketPage)
	}
	if got.Page.Total != nil || got.Items[0].Username != "alice" {
		t.Fatalf("bets output = %#v", got)
	}
}

func TestMarketGroupActivityUsesServiceTotal(t *testing.T) {
	svc := &activityToolService{}
	rt := NewRuntime(svc, nil)
	_, got, err := rt.ListMarketGroupBets(context.Background(), &mcp.CallToolRequest{}, MarketGroupActivityInput{GroupID: 7, Limit: 20, Offset: 20})
	if err != nil {
		t.Fatalf("ListMarketGroupBets returned error: %v", err)
	}
	if got.Page.Total == nil || *got.Page.Total != 33 {
		t.Fatalf("group bets page = %#v", got.Page)
	}
}

func TestGetMarketUserPositionReturnsPublicPosition(t *testing.T) {
	svc := &activityToolService{}
	rt := NewRuntime(svc, nil)
	_, got, err := rt.GetMarketUserPosition(context.Background(), &mcp.CallToolRequest{}, MarketUserPositionInput{MarketID: 8, Username: " alice "})
	if err != nil {
		t.Fatalf("GetMarketUserPosition returned error: %v", err)
	}
	if got.Position.Username != "alice" || got.Position.MarketID != 8 {
		t.Fatalf("position output = %#v", got)
	}
}

func TestListMarketGroupBetsOmitsTotalForNilServiceResult(t *testing.T) {
	rt := NewRuntime(&activityToolService{nilGroupResult: true}, nil)
	_, got, err := rt.ListMarketGroupBets(context.Background(), &mcp.CallToolRequest{}, MarketGroupActivityInput{GroupID: 7, Limit: 20, Offset: 20})
	if err != nil {
		t.Fatalf("ListMarketGroupBets returned error: %v", err)
	}
	assertNilGroupPage(t, got.Items, got.Page)
}

func TestListMarketGroupPositionsOmitsTotalForNilServiceResult(t *testing.T) {
	rt := NewRuntime(&activityToolService{nilGroupResult: true}, nil)
	_, got, err := rt.ListMarketGroupPositions(context.Background(), &mcp.CallToolRequest{}, MarketGroupActivityInput{GroupID: 7, Limit: 20, Offset: 20})
	if err != nil {
		t.Fatalf("ListMarketGroupPositions returned error: %v", err)
	}
	assertNilGroupPage(t, got.Items, got.Page)
}

func TestGetMarketGroupLeaderboardOmitsTotalForNilServiceResult(t *testing.T) {
	rt := NewRuntime(&activityToolService{nilGroupResult: true}, nil)
	_, got, err := rt.GetMarketGroupLeaderboard(context.Background(), &mcp.CallToolRequest{}, MarketGroupActivityInput{GroupID: 7, Limit: 20, Offset: 20})
	if err != nil {
		t.Fatalf("GetMarketGroupLeaderboard returned error: %v", err)
	}
	assertNilGroupPage(t, got.Items, got.Page)
}

func assertNilGroupPage[T any](t *testing.T, items []T, page PageOutput) {
	t.Helper()
	if items == nil || len(items) != 0 {
		t.Fatalf("items = %#v, want non-nil empty slice", items)
	}
	if page.Total != nil || page.Limit != 20 || page.Offset != 20 || page.Count != 0 || page.HasMore {
		t.Fatalf("page = %#v", page)
	}
}
