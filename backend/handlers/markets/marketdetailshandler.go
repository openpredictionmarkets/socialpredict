package marketshandlers

import (
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
			writeInvalidRequest(w)
			return
		}

		// 2. Call domain service to get market details
		details, err := svc.GetMarketDetails(r.Context(), marketId)
		if err != nil {
			writeDetailsError(w, err)
			return
		}

		// 4. Convert domain model to response DTO
		// The domain service should provide all necessary data including creator info
		response := dto.MarketDetailsResponse{
			Market:             publicMarketResponseFromDomain(details.Market),
			Creator:            creatorResponseFromSummary(details.Creator),
			ProbabilityChanges: probabilityChangesToResponse(details.ProbabilityChanges),
			NumUsers:           details.NumUsers,
			TotalVolume:        details.TotalVolume,
			MarketDust:         details.MarketDust,
		}

		// 5. Return response
		_ = writeJSON(w, http.StatusOK, response)
	}
}
