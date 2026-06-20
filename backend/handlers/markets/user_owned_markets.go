package marketshandlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	authsvc "socialpredict/internal/service/auth"
)

type userOwnedMarketLister interface {
	ListLifecycleMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error)
	GetMarketGroupForMarket(ctx context.Context, marketID int64) (*dmarkets.MarketGroup, error)
}

type userOwnedMarketsResponse struct {
	Markets []*dto.MarketResponse `json:"markets"`
	Total   int                   `json:"total"`
}

// ListUserOwnedMarketsHandler returns markets created by or currently stewarded by a user.
func ListUserOwnedMarketsHandler(svc userOwnedMarketLister, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		if svc == nil || auth == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		if _, authErr := auth.CurrentUser(r); authErr != nil {
			_ = authhttp.WriteFailure(w, authErr)
			return
		}

		username := strings.TrimSpace(mux.Vars(r)["username"])
		if username == "" {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		filters := dmarkets.ListFilters{
			Status:  dmarkets.MarketStatusAll,
			OwnedBy: username,
			Limit:   boundedQueryInt(r.URL.Query().Get("limit"), 50, 1, 100),
			Offset:  boundedQueryInt(r.URL.Query().Get("offset"), 0, 0, 100000),
		}
		markets, err := svc.ListLifecycleMarkets(r.Context(), filters)
		if err != nil {
			writeLifecycleListError(w, err)
			return
		}

		response := lifecycleMarketResponseFromMarkets(r.Context(), svc, markets)
		_ = handlers.WriteResult(w, http.StatusOK, userOwnedMarketsResponse{
			Markets: response.Markets,
			Total:   response.Total,
		})
	}
}
