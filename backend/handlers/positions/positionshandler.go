package positions

import (
	"encoding/json"
	"net/http"
	"socialpredict/errors"
	"socialpredict/util"

	"github.com/gorilla/mux"
)

func MarketDBPMPositionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketIdStr := vars["marketId"]

	// open up database to utilize connection pooling
	db := util.GetDB()

	marketDBPMPositions, err := CalculateMarketPositions_WPAM_DBPM(db, marketIdStr)
	if errors.HandleHTTPError(w, err, http.StatusBadRequest, "Invalid request or data processing error.") {
		return // Stop execution if there was an error.
	}

	// Respond with the bets display information
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(marketDBPMPositions)
}
