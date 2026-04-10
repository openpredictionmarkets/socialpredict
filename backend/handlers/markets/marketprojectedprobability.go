package marketshandlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"

	"github.com/gorilla/mux"
)

// ProjectNewProbabilityHandler handles the projection of a new probability based on a new bet.
func ProjectNewProbabilityHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Parse HTTP parameters
		vars := mux.Vars(r)
		marketIdStr := vars["marketId"]
		amountStr := vars["amount"]
		outcome := vars["outcome"]

		// Parse marketId
		marketId, err := strconv.ParseInt(marketIdStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid market ID", http.StatusBadRequest)
			return
		}

		// Parse amount
		amount, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid amount value", http.StatusBadRequest)
			return
		}

		// 2. Build domain request
		projectionReq := dmarkets.ProbabilityProjectionRequest{
			MarketID: marketId,
			Amount:   amount,
			Outcome:  outcome,
		}

		// 3. Call domain service
		projection, err := svc.ProjectProbability(r.Context(), projectionReq)
		if err != nil {
			// 4. Map domain errors to HTTP status codes
			switch err {
			case dmarkets.ErrMarketNotFound:
				http.Error(w, "Market not found", http.StatusNotFound)
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid input parameters", http.StatusBadRequest)
			default:
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// 5. Return response DTO
		response := dto.ProbabilityProjectionResponse{
			MarketID:             marketId,
			CurrentProbability:   projection.CurrentProbability,
			ProjectedProbability: projection.ProjectedProbability,
			Amount:               amount,
			Outcome:              outcome,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
