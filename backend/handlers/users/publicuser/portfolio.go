package publicuser

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
)

// GetPortfolioHandler returns an HTTP handler that responds with a user's portfolio by delegating to the users service.
func GetPortfolioHandler(svc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		username := mux.Vars(r)["username"]
		if username == "" {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		portfolio, err := svc.GetUserPortfolio(r.Context(), username)
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		items := make([]dto.PortfolioItemResponse, 0, len(portfolio.Items))
		for _, item := range portfolio.Items {
			items = append(items, dto.PortfolioItemResponse{
				MarketID:       item.MarketID,
				QuestionTitle:  item.QuestionTitle,
				YesSharesOwned: item.YesSharesOwned,
				NoSharesOwned:  item.NoSharesOwned,
				LastBetPlaced:  item.LastBetPlaced,
			})
		}

		response := dto.PortfolioResponse{
			PortfolioItems:   items,
			TotalSharesOwned: portfolio.TotalSharesOwned,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}
