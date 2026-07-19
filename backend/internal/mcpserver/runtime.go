package mcpserver

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	cmsdiscovery "socialpredict/handlers/cms/marketdiscovery"
	dmarkets "socialpredict/internal/domain/markets"
)

type MarketService interface {
	ListMarketTags(context.Context, bool) ([]dmarkets.MarketTag, error)
	ListMarketDiscovery(context.Context, dmarkets.ListFilters) (*dmarkets.MarketDiscoveryPage, error)
	SearchMarketDiscovery(context.Context, string, dmarkets.SearchFilters) (*dmarkets.MarketDiscoverySearchResults, error)
	GetMarketDetails(context.Context, int64) (*dmarkets.MarketOverview, error)
	GetMarketSummaryReadModel(context.Context, int64) (*dmarkets.MarketSummaryReadModel, error)
	GetMarketGroupOverview(context.Context, int64) (*dmarkets.MarketGroupOverview, error)
	GetMarketGroupForMarket(context.Context, int64) (*dmarkets.MarketGroup, error)
	GetMarketBetsPage(context.Context, int64, dmarkets.Page) ([]*dmarkets.BetDisplayInfo, error)
	GetMarketPositionsPage(context.Context, int64, dmarkets.Page) (dmarkets.MarketPositions, error)
	GetUserPositionInMarket(context.Context, int64, string) (*dmarkets.UserPosition, error)
	GetMarketLeaderboard(context.Context, int64, dmarkets.Page) ([]*dmarkets.LeaderboardRow, error)
	GetMarketGroupBetsPage(context.Context, int64, dmarkets.Page) (*dmarkets.MarketGroupBetsPage, error)
	GetMarketGroupPositionsPage(context.Context, int64, dmarkets.Page) (*dmarkets.MarketGroupPositionsPage, error)
	GetMarketGroupLeaderboardPage(context.Context, int64, dmarkets.Page) (*dmarkets.MarketGroupLeaderboardPage, error)
	ProjectProbability(context.Context, dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error)
}

type DiscoveryService interface {
	GetComposition(string) (*cmsdiscovery.PageComposition, error)
}

type Runtime struct {
	markets   MarketService
	discovery DiscoveryService
	resolver  PrincipalResolver
}

func NewRuntime(markets MarketService, discovery DiscoveryService) *Runtime {
	return &Runtime{markets: markets, discovery: discovery, resolver: AnonymousResolver{}}
}

func (rt *Runtime) require(ctx context.Context, level AccessLevel) error {
	if rt == nil || rt.resolver == nil {
		return &ToolError{Code: "internal_error", Message: "mcp runtime is not initialized"}
	}
	_, err := rt.resolver.Resolve(ctx, level)
	return err
}

func (rt *Runtime) MCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "socialpredict-public-markets", Version: "v0.1.0"}, nil)
	rt.RegisterTools(server)
	return server
}

func (rt *Runtime) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "list_market_tags", Description: "List active SocialPredict market tags."}, rt.ListMarketTags)
	mcp.AddTool(server, &mcp.Tool{Name: "get_market_tag", Description: "Get one active SocialPredict market tag by slug."}, rt.GetMarketTag)
	mcp.AddTool(server, &mcp.Tool{Name: "list_markets", Description: "List public SocialPredict markets using status and tag filters."}, rt.ListMarkets)
	mcp.AddTool(server, &mcp.Tool{Name: "search_markets", Description: "Search public SocialPredict markets with fallback metadata."}, rt.SearchMarkets)
	mcp.AddTool(server, &mcp.Tool{Name: "get_market_discovery", Description: "Get the public market discovery page and rows."}, rt.GetMarketDiscovery)
	mcp.AddTool(server, &mcp.Tool{Name: "get_market", Description: "Get details for one public SocialPredict market."}, rt.GetMarket)
	mcp.AddTool(server, &mcp.Tool{Name: "get_market_summary", Description: "Get the accounting summary for one public SocialPredict market."}, rt.GetMarketSummary)
	mcp.AddTool(server, &mcp.Tool{Name: "quote_market_probability", Description: "Quote the projected probability for a hypothetical market bet."}, rt.QuoteMarketProbability)
}

func cleanQuery(query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", &ToolError{Code: "validation_error", Message: "query is required"}
	}
	return query, nil
}
