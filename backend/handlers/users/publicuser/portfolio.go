package publicuser

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
)

// GetPortfolioHandler returns an HTTP handler that responds with a user's portfolio by delegating to the users service.
func GetPortfolioHandler(svc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		username := mux.Vars(r)["username"]
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}

		portfolio, err := svc.GetUserPortfolio(r.Context(), username)
		if err != nil {
			http.Error(w, "failed to fetch user portfolio", http.StatusInternalServerError)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

