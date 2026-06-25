package marketshandlers

import (
	"context"
	"net/http"
	"strconv"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	authsvc "socialpredict/internal/service/auth"
)

type lifecycleMarketLister interface {
	ListLifecycleMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error)
	GetMarketGroupForMarket(ctx context.Context, marketID int64) (*dmarkets.MarketGroup, error)
}

type lifecycleMarketDiscoveryLister interface {
	ListLifecycleMarketDiscovery(ctx context.Context, filters dmarkets.ListFilters) (*dmarkets.MarketDiscoveryPage, error)
}

type lifecycleMarketsResponse struct {
	Markets []*dto.MarketResponse `json:"markets"`
	Total   int                   `json:"total"`
}

// ListMyLifecycleMarketsHandler returns proposed/published/rejected markets stewarded by the current user.
func ListMyLifecycleMarketsHandler(svc lifecycleMarketLister, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		if svc == nil || auth == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		user, authErr := auth.CurrentUser(r)
		if authErr != nil {
			_ = authhttp.WriteFailure(w, authErr)
			return
		}

		filters, ok := parseLifecycleMarketFilters(w, r)
		if !ok {
			return
		}
		filters.OwnedBy = user.Username

		if discoverySvc, ok := svc.(lifecycleMarketDiscoveryLister); ok {
			page, err := discoverySvc.ListLifecycleMarketDiscovery(r.Context(), filters)
			if err != nil {
				writeLifecycleListError(w, err)
				return
			}
			_ = handlers.WriteResult(w, http.StatusOK, lifecycleMarketResponseFromDiscoveryRows(r.Context(), svc, page))
			return
		}

		markets, err := svc.ListLifecycleMarkets(r.Context(), filters)
		if err != nil {
			writeLifecycleListError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, lifecycleMarketResponseFromMarkets(r.Context(), svc, markets))
	}
}

func parseLifecycleMarketFilters(w http.ResponseWriter, r *http.Request) (dmarkets.ListFilters, bool) {
	query := r.URL.Query()
	status := dmarkets.NormalizeLifecycleStatus(query.Get("status"))
	if status == "" || status == dmarkets.MarketLifecyclePublished {
		status = dmarkets.MarketLifecyclePublished
	}

	limit := boundedQueryInt(query.Get("limit"), 50, 1, 100)
	offset := boundedQueryInt(query.Get("offset"), 0, 0, 100000)

	filters := dmarkets.ListFilters{
		Status: status,
		Query:  query.Get("query"),
		Limit:  limit,
		Offset: offset,
	}
	if _, ok := allowedLifecycleQueueStatus(status); !ok {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return dmarkets.ListFilters{}, false
	}
	return filters, true
}

func allowedLifecycleQueueStatus(status string) (string, bool) {
	switch status {
	case dmarkets.MarketLifecycleProposed, dmarkets.MarketLifecyclePublished, dmarkets.MarketLifecycleRejected:
		return status, true
	default:
		return "", false
	}
}

func boundedQueryInt(value string, fallback int, min int, max int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	if parsed < min {
		return min
	}
	if parsed > max {
		return max
	}
	return parsed
}

func writeLifecycleListError(w http.ResponseWriter, err error) {
	switch err {
	case dmarkets.ErrInvalidInput:
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
	default:
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}

func lifecycleMarketResponseFromMarkets(ctx context.Context, provider any, markets []*dmarkets.Market) lifecycleMarketsResponse {
	items := make([]*dto.MarketResponse, 0, len(markets))
	for _, market := range markets {
		items = append(items, marketToResponseWithGroup(ctx, provider, market))
	}
	return lifecycleMarketsResponse{Markets: items, Total: len(items)}
}

func lifecycleMarketResponseFromDiscoveryRows(ctx context.Context, provider any, page *dmarkets.MarketDiscoveryPage) lifecycleMarketsResponse {
	if page == nil {
		return lifecycleMarketsResponse{Markets: []*dto.MarketResponse{}, Total: 0}
	}
	items := make([]*dto.MarketResponse, 0, len(page.Rows))
	for _, row := range page.Rows {
		if row.Group == nil || row.Group.ID <= 0 {
			items = append(items, marketToResponseWithGroup(ctx, provider, row.Market))
			continue
		}
		items = append(items, lifecycleMarketGroupResponse(row))
	}
	return lifecycleMarketsResponse{Markets: items, Total: page.Total}
}

func lifecycleMarketGroupResponse(row dmarkets.MarketDiscoveryRow) *dto.MarketResponse {
	representative := row.Market
	if representative == nil && len(row.Children) > 0 {
		representative = row.Children[0]
	}
	response := marketToResponse(representative)
	if response == nil {
		response = &dto.MarketResponse{}
	}
	response.ID = row.Group.ID
	response.QuestionTitle = row.Group.QuestionTitle
	response.Description = row.Group.Description
	response.CreatorUsername = row.Group.CreatorUsername
	response.StewardUsername = row.Group.CurrentStewardUsername()
	response.LifecycleStatus = row.Group.LifecycleStatus
	response.Status = marketGroupLinkFromDomain(row.Group, representativeMarketID(representative, row.Children)).Status
	response.ApprovedBy = row.Group.ApprovedBy
	response.ApprovedAt = row.Group.ApprovedAt
	response.RejectedBy = row.Group.RejectedBy
	response.RejectedAt = row.Group.RejectedAt
	response.RejectionReason = row.Group.RejectionReason
	response.ProposalCost = row.Group.ProposalCost
	response.ResolutionDateTime = row.Group.ResolutionDateTime
	response.CreatedAt = row.Group.CreatedAt
	response.UpdatedAt = row.Group.UpdatedAt
	response.MarketGroup = marketGroupLinkFromDomain(row.Group, representativeMarketID(representative, row.Children))
	response.IsMarketGroup = true
	response.Tags = uniqueMarketTagResponses(tagsFromMarkets(row.Children))
	response.ChildMarkets = make([]*dto.MarketResponse, 0, len(row.Children))
	response.IsResolved = len(row.Children) > 0
	for _, child := range row.Children {
		childResponse := marketToResponse(child)
		childResponse.MarketGroup = marketGroupLinkFromDomain(row.Group, child.ID)
		response.ChildMarkets = append(response.ChildMarkets, childResponse)
		response.IsResolved = response.IsResolved && childResponse.IsResolved
	}
	return response
}

func tagsFromMarkets(markets []*dmarkets.Market) []dmarkets.MarketTag {
	tags := make([]dmarkets.MarketTag, 0)
	for _, market := range markets {
		if market == nil {
			continue
		}
		tags = append(tags, market.Tags...)
	}
	return tags
}
