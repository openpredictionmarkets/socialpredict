package marketshandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	"socialpredict/handlers/cms/marketdiscovery"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/internal/domain/readmodels"
	readmodelrepo "socialpredict/internal/repository/readmodels"
	"socialpredict/logger"
	"socialpredict/models"
)

const (
	marketDiscoverySnapshotKind            = "market_discovery"
	marketDiscoverySnapshotTargetFreshness = 10 * time.Minute
)

type marketDiscoveryService interface {
	GetComposition(slug string) (*marketdiscovery.PageComposition, error)
}

type marketDiscoverySnapshotStore interface {
	Get(ctx context.Context, key string) (*readmodelrepo.Snapshot, error)
	Upsert(ctx context.Context, snapshot readmodelrepo.Snapshot) error
}

type MarketDiscoveryReadModelHandler struct {
	markets   Service
	discovery marketDiscoveryService
	snapshots marketDiscoverySnapshotStore
	clock     func() time.Time
}

func NewMarketDiscoveryReadModelHandler(markets Service, discovery marketDiscoveryService, snapshots marketDiscoverySnapshotStore) *MarketDiscoveryReadModelHandler {
	return &MarketDiscoveryReadModelHandler{
		markets:   markets,
		discovery: discovery,
		snapshots: snapshots,
		clock:     func() time.Time { return time.Now().UTC() },
	}
}

type marketDiscoveryReadModelResponse struct {
	Layout        discoveryPageResponse         `json:"layout"`
	TopicNav      *discoveryPageResponse        `json:"topicNav,omitempty"`
	Markets       []*dto.MarketOverviewResponse `json:"markets"`
	PinnedMarkets []pinnedMarketResponse        `json:"pinnedMarkets"`
	Total         int                           `json:"total"`
	Freshness     dto.Freshness                 `json:"freshness"`
}

type discoveryPageResponse struct {
	Slug                       string                 `json:"slug"`
	Title                      string                 `json:"title"`
	Description                string                 `json:"description"`
	PageType                   string                 `json:"pageType"`
	PrimaryTagSlug             string                 `json:"primaryTagSlug"`
	SearchScope                string                 `json:"searchScope"`
	FeaturedTopicsEnabled      bool                   `json:"featuredTopicsEnabled"`
	FeaturedMarketsEnabled     bool                   `json:"featuredMarketsEnabled"`
	DefaultRecommendationLimit int                    `json:"defaultRecommendationLimit"`
	CuratedRecommendationLimit int                    `json:"curatedRecommendationLimit"`
	RecommendationLimit        int                    `json:"recommendationLimit"`
	Version                    uint                   `json:"version"`
	UpdatedAt                  string                 `json:"updatedAt,omitempty"`
	Pins                       []discoveryPinResponse `json:"pins"`
}

type discoveryPinResponse struct {
	ID             uint   `json:"id,omitempty"`
	PinType        string `json:"pinType"`
	MarketID       int64  `json:"marketId,omitempty"`
	TargetPageSlug string `json:"targetPageSlug,omitempty"`
	Label          string `json:"label,omitempty"`
	SortOrder      int    `json:"sortOrder"`
}

type pinnedMarketResponse struct {
	Pin     discoveryPinResponse      `json:"pin"`
	Details dto.MarketDetailsResponse `json:"details"`
}

func (h *MarketDiscoveryReadModelHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if h == nil || h.markets == nil || h.discovery == nil || h.snapshots == nil {
		writeInternalError(w)
		return
	}

	params, err := parseListMarketsParams(r)
	if err != nil {
		writeInvalidRequest(w)
		return
	}
	slug := strings.TrimSpace(mux.Vars(r)["slug"])
	if slug == "" {
		slug = marketdiscovery.PageSlugMarkets
	}

	snapshotKey := marketDiscoverySnapshotKey(slug, params.status, params.filters.TagSlug, params.limit, params.offset)
	if snapshot, err := h.snapshots.Get(r.Context(), snapshotKey); err == nil && h.snapshotUsable(snapshot) {
		var response marketDiscoveryReadModelResponse
		if unmarshalErr := json.Unmarshal([]byte(snapshot.PayloadJSON), &response); unmarshalErr == nil {
			response.Freshness = readModelFreshnessToResponse(freshnessFromSnapshot(snapshot))
			_ = handlers.WriteResult(w, http.StatusOK, response)
			return
		}
	} else if err != nil {
		logger.LogError("MarketDiscoveryReadModel", "GetSnapshot", err)
	}

	response, err := h.buildResponse(r.Context(), slug, params)
	if err != nil {
		if isRequestCanceled(err) {
			return
		}
		logger.LogError("MarketDiscoveryReadModel", "BuildResponse", err)
		writeListError(w, err)
		return
	}

	payload, err := json.Marshal(response)
	if err == nil {
		now := response.Freshness.GeneratedAt
		if now.IsZero() {
			now = h.now()
		}
		if upsertErr := h.snapshots.Upsert(r.Context(), readmodelrepo.Snapshot{
			Key:         snapshotKey,
			Kind:        marketDiscoverySnapshotKind,
			PayloadJSON: string(payload),
			GeneratedAt: now,
			Source:      "read_model",
		}); upsertErr != nil {
			logger.LogError("MarketDiscoveryReadModel", "UpsertSnapshot", upsertErr)
		}
	}

	_ = handlers.WriteResult(w, http.StatusOK, response)
}

func (h *MarketDiscoveryReadModelHandler) buildResponse(ctx context.Context, slug string, params listMarketsParams) (marketDiscoveryReadModelResponse, error) {
	composition, err := h.discovery.GetComposition(slug)
	if err != nil {
		return marketDiscoveryReadModelResponse{}, err
	}
	layout := discoveryPageResponseFromComposition(composition)

	var topicNav *discoveryPageResponse
	if slug != marketdiscovery.PageSlugMarkets {
		top, topErr := h.discovery.GetComposition(marketdiscovery.PageSlugMarkets)
		if topErr == nil {
			topResponse := discoveryPageResponseFromComposition(top)
			topicNav = &topResponse
		}
	}

	filters := params.filters
	if slug != marketdiscovery.PageSlugMarkets {
		filters.TagSlug = slug
	} else if layout.SearchScope == marketdiscovery.SearchScopeTag && layout.PrimaryTagSlug != "" {
		filters.TagSlug = layout.PrimaryTagSlug
	}

	markets, err := h.fetchMarkets(ctx, params.status, filters, params.limit, params.offset)
	if err != nil {
		return marketDiscoveryReadModelResponse{}, err
	}
	overviews, err := buildMarketOverviewResponses(ctx, h.markets, markets)
	if err != nil {
		return marketDiscoveryReadModelResponse{}, err
	}

	pinned, err := h.pinnedMarketResponses(ctx, layout.Pins)
	if err != nil {
		return marketDiscoveryReadModelResponse{}, err
	}

	return marketDiscoveryReadModelResponse{
		Layout:        layout,
		TopicNav:      topicNav,
		Markets:       overviews,
		PinnedMarkets: pinned,
		Total:         len(overviews),
		Freshness:     readModelFreshnessToResponse(readmodels.NewFreshness(h.now(), "read_model", marketDiscoverySnapshotTargetFreshness, false)),
	}, nil
}

func (h *MarketDiscoveryReadModelHandler) fetchMarkets(ctx context.Context, status string, filters dmarkets.ListFilters, limit int, offset int) ([]*dmarkets.Market, error) {
	if status != "" && filters.CreatedBy == "" && filters.TagSlug == "" {
		return h.markets.ListByStatus(ctx, status, dmarkets.Page{Limit: limit, Offset: offset})
	}
	filters.Limit = limit
	filters.Offset = offset
	return h.markets.ListMarkets(ctx, filters)
}

func (h *MarketDiscoveryReadModelHandler) pinnedMarketResponses(ctx context.Context, pins []discoveryPinResponse) ([]pinnedMarketResponse, error) {
	responses := make([]pinnedMarketResponse, 0)
	for _, pin := range pins {
		if pin.PinType != marketdiscovery.PinTypeMarket || pin.MarketID <= 0 {
			continue
		}
		if summaryProvider, ok := h.markets.(marketSummaryProvider); ok {
			summary, err := summaryProvider.GetMarketSummaryReadModel(ctx, pin.MarketID)
			if err != nil {
				return nil, fmt.Errorf("pinned market %d: %w", pin.MarketID, err)
			}
			responses = append(responses, pinnedMarketResponse{
				Pin:     pin,
				Details: marketSummaryToDetailsResponse(ctx, h.markets, summary),
			})
			continue
		}
		details, err := h.markets.GetMarketDetails(ctx, pin.MarketID)
		if err != nil {
			return nil, fmt.Errorf("pinned market %d: %w", pin.MarketID, err)
		}
		responses = append(responses, pinnedMarketResponse{
			Pin:     pin,
			Details: marketDetailsToResponse(ctx, h.markets, details),
		})
	}
	return responses, nil
}

func (h *MarketDiscoveryReadModelHandler) snapshotUsable(snapshot *readmodelrepo.Snapshot) bool {
	if snapshot == nil || snapshot.PayloadJSON == "" {
		return false
	}
	return h.now().Sub(snapshot.GeneratedAt) <= marketDiscoverySnapshotTargetFreshness
}

func (h *MarketDiscoveryReadModelHandler) now() time.Time {
	if h.clock == nil {
		return time.Now().UTC()
	}
	return h.clock().UTC()
}

func marketDiscoverySnapshotKey(slug string, status string, tagSlug string, limit int, offset int) string {
	if status == "" {
		status = "all"
	}
	if tagSlug == "" {
		tagSlug = "none"
	}
	return fmt.Sprintf("market_discovery:%s:status=%s:tag=%s:limit=%d:offset=%d", slug, status, tagSlug, limit, offset)
}

func freshnessFromSnapshot(snapshot *readmodelrepo.Snapshot) readmodels.Freshness {
	if snapshot == nil {
		return readmodels.NewFreshness(time.Now().UTC(), "read_model", marketDiscoverySnapshotTargetFreshness, false)
	}
	if snapshot.IsStale {
		return readmodels.NewStaleFreshness(snapshot.GeneratedAt, snapshot.Source, marketDiscoverySnapshotTargetFreshness, false, snapshot.StaleReason, snapshot.MarkedStaleAt)
	}
	return readmodels.NewFreshness(snapshot.GeneratedAt, snapshot.Source, marketDiscoverySnapshotTargetFreshness, false)
}

func discoveryPageResponseFromComposition(composition *marketdiscovery.PageComposition) discoveryPageResponse {
	if composition == nil || composition.Page == nil {
		return discoveryPageResponse{Pins: []discoveryPinResponse{}}
	}
	return discoveryPageResponseFromPage(composition.Page, composition.Pins)
}

func discoveryPageResponseFromPage(page *models.MarketDiscoveryPage, pins []models.MarketDiscoveryPin) discoveryPageResponse {
	response := discoveryPageResponse{
		Slug:                       page.Slug,
		Title:                      page.Title,
		Description:                page.Description,
		PageType:                   page.PageType,
		PrimaryTagSlug:             page.PrimaryTagSlug,
		SearchScope:                page.SearchScope,
		FeaturedTopicsEnabled:      page.FeaturedTopicsEnabled,
		FeaturedMarketsEnabled:     page.FeaturedMarketsEnabled,
		DefaultRecommendationLimit: page.DefaultRecommendationLimit,
		CuratedRecommendationLimit: page.CuratedRecommendationLimit,
		Version:                    page.Version,
		Pins:                       discoveryPinResponses(pins),
	}
	if !page.UpdatedAt.IsZero() {
		response.UpdatedAt = page.UpdatedAt.Format(time.RFC3339)
	}
	response.RecommendationLimit = response.DefaultRecommendationLimit
	if page.FeaturedTopicsEnabled || page.FeaturedMarketsEnabled {
		response.RecommendationLimit = response.CuratedRecommendationLimit
	}
	return response
}

func discoveryPinResponses(pins []models.MarketDiscoveryPin) []discoveryPinResponse {
	responses := make([]discoveryPinResponse, 0, len(pins))
	for _, pin := range pins {
		response := discoveryPinResponse{
			ID:             pin.ID,
			PinType:        pin.PinType,
			TargetPageSlug: pin.TargetPageSlug,
			Label:          pin.Label,
			SortOrder:      pin.SortOrder,
		}
		if pin.MarketID != nil {
			response.MarketID = *pin.MarketID
		}
		responses = append(responses, response)
	}
	return responses
}
