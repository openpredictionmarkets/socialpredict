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
}

type lifecycleMarketsResponse struct {
	Markets []*dto.MarketResponse `json:"markets"`
	Total   int                   `json:"total"`
}

// ListMyLifecycleMarketsHandler returns proposed/published/rejected markets for the current user.
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
		filters.CreatedBy = user.Username

		markets, err := svc.ListLifecycleMarkets(r.Context(), filters)
		if err != nil {
			writeLifecycleListError(w, err)
			return
		}

		_ = handlers.WriteResult(w, http.StatusOK, lifecycleMarketResponseFromMarkets(markets))
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

func lifecycleMarketResponseFromMarkets(markets []*dmarkets.Market) lifecycleMarketsResponse {
	items := make([]*dto.MarketResponse, 0, len(markets))
	for _, market := range markets {
		items = append(items, marketToResponse(market))
	}
	return lifecycleMarketsResponse{Markets: items, Total: len(items)}
}
