package mcpserver

import (
	"context"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	cmsdiscovery "socialpredict/handlers/cms/marketdiscovery"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"
)

type ListMarketTagsOutput struct {
	Tags  []MarketTagOutput `json:"tags"`
	Total int               `json:"total"`
}
type GetMarketTagOutput struct {
	Tag MarketTagOutput `json:"tag"`
}
type ListMarketsOutput struct {
	Status  string                        `json:"status"`
	Results PageItems[DiscoveryRowOutput] `json:"results"`
}
type SearchMarketsOutput struct {
	Query           string                        `json:"query"`
	PrimaryStatus   string                        `json:"primaryStatus"`
	PrimaryResults  PageItems[DiscoveryRowOutput] `json:"primaryResults"`
	FallbackResults []DiscoveryRowOutput          `json:"fallbackResults"`
	PrimaryCount    int                           `json:"primaryCount"`
	FallbackCount   int                           `json:"fallbackCount"`
	TotalCount      int                           `json:"totalCount"`
	FallbackUsed    bool                          `json:"fallbackUsed"`
}
type DiscoveryLayoutOutput struct {
	Slug                       string    `json:"slug"`
	Title                      string    `json:"title"`
	Description                string    `json:"description,omitempty"`
	PageType                   string    `json:"pageType,omitempty"`
	PrimaryTagSlug             string    `json:"primaryTagSlug,omitempty"`
	SearchScope                string    `json:"searchScope,omitempty"`
	FeaturedTopicsEnabled      bool      `json:"featuredTopicsEnabled"`
	FeaturedMarketsEnabled     bool      `json:"featuredMarketsEnabled"`
	DefaultRecommendationLimit int       `json:"defaultRecommendationLimit"`
	CuratedRecommendationLimit int       `json:"curatedRecommendationLimit"`
	RecommendationLimit        int       `json:"recommendationLimit"`
	Version                    uint      `json:"version"`
	UpdatedAt                  time.Time `json:"updatedAt,omitempty"`
}
type MarketDiscoveryOutput struct {
	Layout  DiscoveryLayoutOutput         `json:"layout"`
	Markets PageItems[DiscoveryRowOutput] `json:"markets"`
}

func (rt *Runtime) ListMarketTags(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, ListMarketTagsOutput, error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, ListMarketTagsOutput{}, err
	}
	tags, err := rt.markets.ListMarketTags(ctx, false)
	if err != nil {
		return nil, ListMarketTagsOutput{}, MapError(err)
	}
	out := MarketTagOutputsFromDomain(tags)
	return nil, ListMarketTagsOutput{Tags: out, Total: len(out)}, nil
}

func (rt *Runtime) GetMarketTag(ctx context.Context, _ *mcp.CallToolRequest, in SlugInput) (*mcp.CallToolResult, GetMarketTagOutput, error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, GetMarketTagOutput{}, err
	}
	slug, err := NormalizeTagSlug(in.Slug)
	if err != nil {
		return nil, GetMarketTagOutput{}, err
	}
	if slug == "" {
		return nil, GetMarketTagOutput{}, &ToolError{Code: "validation_error", Message: "slug is required"}
	}
	tags, err := rt.markets.ListMarketTags(ctx, false)
	if err != nil {
		return nil, GetMarketTagOutput{}, MapError(err)
	}
	for _, tag := range tags {
		if tag.IsActive && tag.Slug == slug {
			return nil, GetMarketTagOutput{Tag: MarketTagOutputFromDomain(tag)}, nil
		}
	}
	return nil, GetMarketTagOutput{}, &ToolError{Code: "not_found", Message: "market tag not found"}
}

func (rt *Runtime) ListMarkets(ctx context.Context, _ *mcp.CallToolRequest, in MarketListInput) (*mcp.CallToolResult, ListMarketsOutput, error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, ListMarketsOutput{}, err
	}
	status, err := NormalizeStatus(in.Status)
	if err != nil {
		return nil, ListMarketsOutput{}, err
	}
	tagSlug, err := NormalizeTagSlug(in.TagSlug)
	if err != nil {
		return nil, ListMarketsOutput{}, err
	}
	page := NormalizePage(in.Limit, in.Offset)
	result, err := rt.markets.ListMarketDiscovery(ctx, dmarkets.ListFilters{Status: status.Filter, TagSlug: tagSlug, CreatedBy: strings.TrimSpace(in.CreatedBy), Limit: page.Limit, Offset: page.Offset})
	if err != nil {
		return nil, ListMarketsOutput{}, MapError(err)
	}
	rows, total, err := rt.discoveryRows(ctx, result)
	if err != nil {
		return nil, ListMarketsOutput{}, MapError(err)
	}
	return nil, ListMarketsOutput{Status: status.Canonical, Results: NewPageItems(rows, page.Limit, page.Offset, total)}, nil
}

func (rt *Runtime) SearchMarkets(ctx context.Context, _ *mcp.CallToolRequest, in MarketSearchInput) (*mcp.CallToolResult, SearchMarketsOutput, error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, SearchMarketsOutput{}, err
	}
	query, err := cleanQuery(in.Query)
	if err != nil {
		return nil, SearchMarketsOutput{}, err
	}
	status, err := NormalizeStatus(in.Status)
	if err != nil {
		return nil, SearchMarketsOutput{}, err
	}
	tagSlug, err := NormalizeTagSlug(in.TagSlug)
	if err != nil {
		return nil, SearchMarketsOutput{}, err
	}
	page := NormalizePage(in.Limit, in.Offset)
	result, err := rt.markets.SearchMarketDiscovery(ctx, query, dmarkets.SearchFilters{Status: status.Filter, TagSlug: tagSlug, Limit: page.Limit, Offset: page.Offset})
	if err != nil {
		return nil, SearchMarketsOutput{}, MapError(err)
	}
	primary, err := rt.discoveryRowOutputs(ctx, result.PrimaryRows)
	if err != nil {
		return nil, SearchMarketsOutput{}, MapError(err)
	}
	fallback, err := rt.discoveryRowOutputs(ctx, result.FallbackRows)
	if err != nil {
		return nil, SearchMarketsOutput{}, MapError(err)
	}
	return nil, SearchMarketsOutput{Query: result.Query, PrimaryStatus: status.Canonical, PrimaryResults: NewPageItems(primary, page.Limit, page.Offset, nil), FallbackResults: fallback, PrimaryCount: result.PrimaryCount, FallbackCount: result.FallbackCount, TotalCount: result.TotalCount, FallbackUsed: result.FallbackUsed}, nil
}

func (rt *Runtime) GetMarketDiscovery(ctx context.Context, _ *mcp.CallToolRequest, in MarketDiscoveryInput) (*mcp.CallToolResult, MarketDiscoveryOutput, error) {
	if err := rt.require(ctx, AccessPublicRead); err != nil {
		return nil, MarketDiscoveryOutput{}, err
	}
	pageSlug, err := NormalizeTagSlug(in.Slug)
	if err != nil {
		return nil, MarketDiscoveryOutput{}, err
	}
	if pageSlug == "" {
		return nil, MarketDiscoveryOutput{}, &ToolError{Code: "validation_error", Message: "slug is required; use markets for the top page"}
	}
	layout, err := rt.discoveryLayout(pageSlug)
	if err != nil {
		return nil, MarketDiscoveryOutput{}, MapError(err)
	}
	tagSlug := in.TagSlug
	if pageSlug != cmsdiscovery.PageSlugMarkets && tagSlug == "" {
		tagSlug = pageSlug
	}
	_, listOut, err := rt.ListMarkets(ctx, nil, MarketListInput{Status: in.Status, TagSlug: tagSlug, Limit: in.Limit, Offset: in.Offset})
	if err != nil {
		return nil, MarketDiscoveryOutput{}, err
	}
	return nil, MarketDiscoveryOutput{Layout: layout, Markets: listOut.Results}, nil
}

func (rt *Runtime) discoveryLayout(slug string) (DiscoveryLayoutOutput, error) {
	if rt.discovery == nil {
		return DiscoveryLayoutOutput{Slug: slug}, nil
	}
	composition, err := rt.discovery.GetComposition(slug)
	if err != nil {
		return DiscoveryLayoutOutput{}, err
	}
	return DiscoveryLayoutOutputFromPage(composition.Page), nil
}

func DiscoveryLayoutOutputFromPage(page *models.MarketDiscoveryPage) DiscoveryLayoutOutput {
	if page == nil {
		return DiscoveryLayoutOutput{Slug: cmsdiscovery.PageSlugMarkets}
	}
	limit := page.DefaultRecommendationLimit
	if page.FeaturedTopicsEnabled || page.FeaturedMarketsEnabled {
		limit = page.CuratedRecommendationLimit
	}
	return DiscoveryLayoutOutput{Slug: page.Slug, Title: page.Title, Description: page.Description, PageType: page.PageType, PrimaryTagSlug: page.PrimaryTagSlug, SearchScope: page.SearchScope, FeaturedTopicsEnabled: page.FeaturedTopicsEnabled, FeaturedMarketsEnabled: page.FeaturedMarketsEnabled, DefaultRecommendationLimit: page.DefaultRecommendationLimit, CuratedRecommendationLimit: page.CuratedRecommendationLimit, RecommendationLimit: limit, Version: page.Version, UpdatedAt: page.UpdatedAt}
}

func (rt *Runtime) discoveryRows(ctx context.Context, page *dmarkets.MarketDiscoveryPage) ([]DiscoveryRowOutput, *int, error) {
	if page == nil {
		return []DiscoveryRowOutput{}, nil, nil
	}
	total := page.Total
	rows, err := rt.discoveryRowOutputs(ctx, page.Rows)
	return rows, &total, err
}
func (rt *Runtime) discoveryRowOutputs(ctx context.Context, rows []dmarkets.MarketDiscoveryRow) ([]DiscoveryRowOutput, error) {
	out := make([]DiscoveryRowOutput, 0, len(rows))
	for _, row := range rows {
		enriched, err := rt.discoveryRowOutput(ctx, row)
		if err != nil {
			return nil, err
		}
		out = append(out, enriched)
	}
	return out, nil
}
func (rt *Runtime) discoveryRowOutput(ctx context.Context, row dmarkets.MarketDiscoveryRow) (DiscoveryRowOutput, error) {
	if row.Group == nil || row.Group.ID <= 0 {
		overview, err := rt.marketOverviewOutput(ctx, row.Market)
		if err != nil {
			return DiscoveryRowOutput{}, err
		}
		return DiscoveryRowOutput{Market: &overview, TotalVolume: overview.TotalVolume, MarketDust: overview.MarketDust}, nil
	}
	children := row.Children
	if len(children) == 0 && row.Market != nil {
		children = []*dmarkets.Market{row.Market}
	}
	out := DiscoveryRowOutput{IsMarketGroup: true, Group: MarketGroupOutputFromDomain(row.Group), ChildMarkets: []MarketOverviewOutput{}}
	for _, child := range children {
		overview, err := rt.marketOverviewOutput(ctx, child)
		if err != nil {
			return DiscoveryRowOutput{}, err
		}
		out.ChildMarkets = append(out.ChildMarkets, overview)
		out.TotalVolume += overview.TotalVolume
		out.MarketDust += overview.MarketDust
	}
	return out, nil
}
func (rt *Runtime) marketOverviewOutput(ctx context.Context, market *dmarkets.Market) (MarketOverviewOutput, error) {
	if market == nil || market.ID <= 0 {
		return MarketOverviewOutputFromDomain(&dmarkets.MarketOverview{Market: market}), nil
	}
	summary, err := rt.markets.GetMarketSummaryReadModel(ctx, market.ID)
	if err != nil {
		return MarketOverviewOutput{}, err
	}
	out := MarketOverviewOutputFromDomain(&dmarkets.MarketOverview{Market: summary.Market, Creator: summary.Creator, LastProbability: summary.Accounting.LastProbability, NumUsers: summary.Accounting.UserCount, TotalVolume: summary.Accounting.VolumeWithDust, MarketDust: summary.Accounting.MarketDust})
	if group, err := rt.markets.GetMarketGroupForMarket(ctx, market.ID); err == nil && group != nil {
		out.Market.MarketGroup = MarketGroupLinkOutputFromDomain(group, market.ID)
	}
	return out, nil
}
