package marketshandlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"

	"github.com/gorilla/mux"
)

// MarketDetailsHandler handles requests for detailed market information
func MarketDetailsHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Parse HTTP parameters
		vars := mux.Vars(r)
		marketIdStr := vars["marketId"]

		marketId, err := strconv.ParseInt(marketIdStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid market ID", http.StatusBadRequest)
			return
		}

		// 2. Call domain service to get market details
		details, err := svc.GetMarketDetails(r.Context(), marketId)
		if err != nil {
			// 3. Map domain errors to HTTP status codes
			switch err {
			case dmarkets.ErrMarketNotFound:
				http.Error(w, "Market not found", http.StatusNotFound)
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid market ID", http.StatusBadRequest)
			default:
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// 4. Convert domain model to response DTO
		// The domain service should provide all necessary data including creator info
		response := dto.MarketDetailsResponse{
			MarketID:           marketId,
			Market:             details.Market,
			Creator:            details.Creator, // Creator info should come from domain service
			ProbabilityChanges: details.ProbabilityChanges,
			NumUsers:           details.NumUsers,
			TotalVolume:        details.TotalVolume,
			MarketDust:         details.MarketDust,
		}

		// 5. Return response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
