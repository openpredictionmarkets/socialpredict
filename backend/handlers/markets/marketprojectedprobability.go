package marketshandlers

import (
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"

	"github.com/gorilla/mux"
)

// ProjectNewProbabilityHandler handles the projection of a new probability based on a new bet.
func ProjectNewProbabilityHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		marketIdStr := vars["marketId"]
		amountStr := vars["amount"]
		outcome := vars["outcome"]

		marketId, err := strconv.ParseInt(marketIdStr, 10, 64)
		if err != nil {
			writeInvalidRequest(w)
			return
		}

		amount, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			writeInvalidRequest(w)
			return
		}

		projectionReq := dmarkets.ProbabilityProjectionRequest{
			MarketID: marketId,
			Amount:   amount,
			Outcome:  outcome,
		}

		projection, err := svc.ProjectProbability(r.Context(), projectionReq)
		if err != nil {
			writeProjectionError(w, err)
			return
		}

		response := dto.ProbabilityProjectionResponse{
			MarketID:             marketId,
			CurrentProbability:   projection.CurrentProbability,
			ProjectedProbability: projection.ProjectedProbability,
			Amount:               amount,
			Outcome:              outcome,
		}

		_ = writeJSON(w, http.StatusOK, response)
	}
}
