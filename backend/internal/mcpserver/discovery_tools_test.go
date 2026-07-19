package mcpserver

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	cmsdiscovery "socialpredict/handlers/cms/marketdiscovery"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"
)

type discoveryToolMarketService struct {
	tags          []dmarkets.MarketTag
	listFilters   dmarkets.ListFilters
	searchQuery   string
	searchFilters dmarkets.SearchFilters
	summaries     map[int64]*dmarkets.MarketSummaryReadModel
}

func (s *discoveryToolMarketService) ListMarketTags(context.Context, bool) ([]dmarkets.MarketTag, error) {
	return s.tags, nil
}
func (s *discoveryToolMarketService) ListMarketDiscovery(_ context.Context, filters dmarkets.ListFilters) (*dmarkets.MarketDiscoveryPage, error) {
	s.listFilters = filters
	return &dmarkets.MarketDiscoveryPage{Rows: []dmarkets.MarketDiscoveryRow{{Market: &dmarkets.Market{ID: 1, QuestionTitle: "One", Status: dmarkets.MarketStatusActive}}}, Total: 1}, nil
}
func (s *discoveryToolMarketService) SearchMarketDiscovery(_ context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.MarketDiscoverySearchResults, error) {
	s.searchQuery, s.searchFilters = query, filters
	return &dmarkets.MarketDiscoverySearchResults{Query: query, PrimaryStatus: filters.Status, PrimaryRows: []dmarkets.MarketDiscoveryRow{{Market: &dmarkets.Market{ID: 2, QuestionTitle: "Two", Status: filters.Status}}}, PrimaryCount: 1, TotalCount: 1}, nil
}
func (s *discoveryToolMarketService) GetMarketSummaryReadModel(_ context.Context, marketID int64) (*dmarkets.MarketSummaryReadModel, error) {
	if s.summaries != nil && s.summaries[marketID] != nil {
		return s.summaries[marketID], nil
	}
	return &dmarkets.MarketSummaryReadModel{Market: &dmarkets.Market{ID: marketID, QuestionTitle: "Summary", Status: dmarkets.MarketStatusActive}, Accounting: dmarkets.MarketAccountingSnapshot{MarketID: marketID, LastProbability: .5, VolumeWithDust: 100, MarketDust: 3}}, nil
}

func TestListMarketTagsReturnsActiveTags(t *testing.T) {
	svc := &discoveryToolMarketService{tags: []dmarkets.MarketTag{{ID: 1, Slug: "macro", DisplayName: "Macro", IsActive: true}}}
	_, got, err := NewRuntime(svc, nil).ListMarketTags(context.Background(), &mcp.CallToolRequest{}, EmptyInput{})
	if err != nil {
		t.Fatalf("ListMarketTags returned error: %v", err)
	}
	if got.Total != 1 || got.Tags[0].Slug != "macro" {
		t.Fatalf("tags output = %#v", got)
	}
}

func TestGetMarketTagNormalizesSlugAndRejectsMissing(t *testing.T) {
	svc := &discoveryToolMarketService{tags: []dmarkets.MarketTag{{ID: 1, Slug: "macro-news", DisplayName: "Macro News", IsActive: true}}}
	rt := NewRuntime(svc, nil)
	_, got, err := rt.GetMarketTag(context.Background(), &mcp.CallToolRequest{}, SlugInput{Slug: " --Macro-News-- "})
	if err != nil {
		t.Fatalf("GetMarketTag returned error: %v", err)
	}
	if got.Tag.Slug != "macro-news" {
		t.Fatalf("tag slug = %q", got.Tag.Slug)
	}
	_, _, err = rt.GetMarketTag(context.Background(), &mcp.CallToolRequest{}, SlugInput{Slug: "missing"})
	if mapped := MapError(err); mapped.Code != "not_found" {
		t.Fatalf("missing tag error = %#v", mapped)
	}
}

func TestListMarketsUsesDiscoveryWithOpenAliasAndPagination(t *testing.T) {
	svc := &discoveryToolMarketService{}
	_, got, err := NewRuntime(svc, nil).ListMarkets(context.Background(), &mcp.CallToolRequest{}, MarketListInput{Status: "open", TagSlug: " macro ", Limit: 200, Offset: -1})
	if err != nil {
		t.Fatalf("ListMarkets returned error: %v", err)
	}
	if svc.listFilters.Status != "active" || svc.listFilters.TagSlug != "macro" || svc.listFilters.Limit != 100 || svc.listFilters.Offset != 0 {
		t.Fatalf("filters = %#v", svc.listFilters)
	}
	if got.Status != "active" || got.Results.Page.Total == nil || *got.Results.Page.Total != 1 {
		t.Fatalf("output = %#v", got)
	}
}

func TestSearchMarketsPreservesFallbackMetadata(t *testing.T) {
	svc := &discoveryToolMarketService{}
	_, got, err := NewRuntime(svc, nil).SearchMarkets(context.Background(), &mcp.CallToolRequest{}, MarketSearchInput{Query: "rain", Status: "resolved"})
	if err != nil {
		t.Fatalf("SearchMarkets returned error: %v", err)
	}
	if svc.searchQuery != "rain" || svc.searchFilters.Status != "resolved" {
		t.Fatalf("search args = %q %#v", svc.searchQuery, svc.searchFilters)
	}
	if got.Query != "rain" || got.PrimaryStatus != "resolved" || got.TotalCount != 1 {
		t.Fatalf("search output = %#v", got)
	}
}

type discoveryServiceStub struct{}

func (discoveryServiceStub) GetComposition(slug string) (*cmsdiscovery.PageComposition, error) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	return &cmsdiscovery.PageComposition{Page: &models.MarketDiscoveryPage{Slug: slug, Title: "Markets", PageType: "markets", UpdatedAt: now}}, nil
}
func TestGetMarketDiscoveryReturnsLayoutAndRows(t *testing.T) {
	_, got, err := NewRuntime(&discoveryToolMarketService{}, discoveryServiceStub{}).GetMarketDiscovery(context.Background(), &mcp.CallToolRequest{}, MarketDiscoveryInput{Slug: "markets", TagSlug: "macro"})
	if err != nil {
		t.Fatalf("GetMarketDiscovery returned error: %v", err)
	}
	if got.Layout.Slug != "markets" || got.Markets.Page.Total == nil || *got.Markets.Page.Total != 1 {
		t.Fatalf("discovery output = %#v", got)
	}
}

func (s *discoveryToolMarketService) GetMarketDetails(context.Context, int64) (*dmarkets.MarketOverview, error) {
	return nil, dmarkets.ErrInvalidInput
}
func (s *discoveryToolMarketService) GetMarketGroupOverview(context.Context, int64) (*dmarkets.MarketGroupOverview, error) {
	return nil, dmarkets.ErrInvalidInput
}
func (s *discoveryToolMarketService) GetMarketGroupForMarket(context.Context, int64) (*dmarkets.MarketGroup, error) {
	return nil, nil
}
func (s *discoveryToolMarketService) GetMarketBetsPage(context.Context, int64, dmarkets.Page) ([]*dmarkets.BetDisplayInfo, error) {
	return nil, dmarkets.ErrInvalidInput
}
func (s *discoveryToolMarketService) GetMarketPositionsPage(context.Context, int64, dmarkets.Page) (dmarkets.MarketPositions, error) {
	return nil, dmarkets.ErrInvalidInput
}
func (s *discoveryToolMarketService) GetUserPositionInMarket(context.Context, int64, string) (*dmarkets.UserPosition, error) {
	return nil, dmarkets.ErrInvalidInput
}
func (s *discoveryToolMarketService) GetMarketLeaderboard(context.Context, int64, dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
	return nil, dmarkets.ErrInvalidInput
}
func (s *discoveryToolMarketService) GetMarketGroupBetsPage(context.Context, int64, dmarkets.Page) (*dmarkets.MarketGroupBetsPage, error) {
	return nil, dmarkets.ErrInvalidInput
}
func (s *discoveryToolMarketService) GetMarketGroupPositionsPage(context.Context, int64, dmarkets.Page) (*dmarkets.MarketGroupPositionsPage, error) {
	return nil, dmarkets.ErrInvalidInput
}
func (s *discoveryToolMarketService) GetMarketGroupLeaderboardPage(context.Context, int64, dmarkets.Page) (*dmarkets.MarketGroupLeaderboardPage, error) {
	return nil, dmarkets.ErrInvalidInput
}
func (s *discoveryToolMarketService) ProjectProbability(context.Context, dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	return nil, dmarkets.ErrInvalidInput
}
