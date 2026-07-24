package mcpserver

import (
	"context"
	"errors"
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
	listPage      *dmarkets.MarketDiscoveryPage
	batchCalls    int
	batchMarkets  []*dmarkets.Market
	batchErr      error
	summaryCalls  int
	groupCalls    int
}

func (s *discoveryToolMarketService) ListMarketTags(context.Context, bool) ([]dmarkets.MarketTag, error) {
	return s.tags, nil
}
func (s *discoveryToolMarketService) ListMarketDiscovery(_ context.Context, filters dmarkets.ListFilters) (*dmarkets.MarketDiscoveryPage, error) {
	s.listFilters = filters
	if s.listPage != nil {
		return s.listPage, nil
	}
	return &dmarkets.MarketDiscoveryPage{Rows: []dmarkets.MarketDiscoveryRow{{Market: &dmarkets.Market{ID: 1, QuestionTitle: "One", Status: dmarkets.MarketStatusActive}}}, Total: 1}, nil
}
func (s *discoveryToolMarketService) SearchMarketDiscovery(_ context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.MarketDiscoverySearchResults, error) {
	s.searchQuery, s.searchFilters = query, filters
	return &dmarkets.MarketDiscoverySearchResults{Query: query, PrimaryStatus: filters.Status, PrimaryRows: []dmarkets.MarketDiscoveryRow{{Market: &dmarkets.Market{ID: 2, QuestionTitle: "Two", Status: filters.Status}}}, PrimaryCount: 1, TotalCount: 1}, nil
}
func (s *discoveryToolMarketService) GetMarketSummaryReadModel(_ context.Context, marketID int64) (*dmarkets.MarketSummaryReadModel, error) {
	s.summaryCalls++
	if s.summaries != nil && s.summaries[marketID] != nil {
		return s.summaries[marketID], nil
	}
	return &dmarkets.MarketSummaryReadModel{Market: &dmarkets.Market{ID: marketID, QuestionTitle: "Summary", Status: dmarkets.MarketStatusActive}, Accounting: dmarkets.MarketAccountingSnapshot{MarketID: marketID, LastProbability: .5, VolumeWithDust: 100, MarketDust: 3}}, nil
}
func (s *discoveryToolMarketService) GetMarketDiscoverySummaries(
	_ context.Context,
	markets []*dmarkets.Market,
) (map[int64]*dmarkets.MarketSummaryReadModel, error) {
	s.batchCalls++
	s.batchMarkets = append([]*dmarkets.Market(nil), markets...)
	if s.batchErr != nil {
		return nil, s.batchErr
	}
	out := map[int64]*dmarkets.MarketSummaryReadModel{}
	for _, market := range markets {
		out[market.ID] = s.summaryFor(market)
	}
	return out, nil
}
func (s *discoveryToolMarketService) summaryFor(market *dmarkets.Market) *dmarkets.MarketSummaryReadModel {
	if s.summaries != nil && s.summaries[market.ID] != nil {
		return s.summaries[market.ID]
	}
	return &dmarkets.MarketSummaryReadModel{
		Market: market,
		Accounting: dmarkets.MarketAccountingSnapshot{
			MarketID:        market.ID,
			LastProbability: .5,
			VolumeWithDust:  100,
			MarketDust:      3,
		},
	}
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

func TestListMarketsUsesSingleBatchEnrichmentAndRowGroupLinks(t *testing.T) {
	childA := &dmarkets.Market{ID: 11, QuestionTitle: "A", Status: dmarkets.MarketStatusActive}
	childB := &dmarkets.Market{ID: 12, QuestionTitle: "B", Status: dmarkets.MarketStatusActive}
	group := &dmarkets.MarketGroup{
		ID:              9,
		QuestionTitle:   "Group",
		LifecycleStatus: dmarkets.MarketLifecyclePublished,
		Members: []dmarkets.MarketGroupMember{
			{MarketID: childA.ID, AnswerLabel: "A", DisplayOrder: 1},
			{MarketID: childB.ID, AnswerLabel: "B", DisplayOrder: 2},
		},
	}
	svc := &discoveryToolMarketService{
		listPage: &dmarkets.MarketDiscoveryPage{
			Rows:  []dmarkets.MarketDiscoveryRow{{Market: childA, Group: group, Children: []*dmarkets.Market{childA, childB}}},
			Total: 1,
		},
	}

	_, got, err := NewRuntime(svc, nil).ListMarkets(context.Background(), nil, MarketListInput{})
	if err != nil {
		t.Fatalf("ListMarkets returned error: %v", err)
	}
	if svc.batchCalls != 1 {
		t.Fatalf("batch calls = %d, want 1", svc.batchCalls)
	}
	if svc.summaryCalls != 0 || svc.groupCalls != 0 {
		t.Fatalf("per-market calls: summary=%d group=%d, want 0", svc.summaryCalls, svc.groupCalls)
	}
	if len(svc.batchMarkets) != 2 {
		t.Fatalf("batch markets = %#v, want 2", svc.batchMarkets)
	}
	link := got.Results.Items[0].ChildMarkets[0].Market.MarketGroup
	if link == nil || link.ID != group.ID || link.AnswerLabel != "A" {
		t.Fatalf("group link = %#v", link)
	}
}

func TestListMarketsMapsBatchEnrichmentError(t *testing.T) {
	svc := &discoveryToolMarketService{batchErr: errors.New("batch failed")}
	_, _, err := NewRuntime(svc, nil).ListMarkets(context.Background(), nil, MarketListInput{})
	if err == nil {
		t.Fatalf("ListMarkets returned nil error")
	}
	if got := MapError(err); got.Code != "internal_error" {
		t.Fatalf("error = %#v, want internal_error", got)
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
	s.groupCalls++
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
