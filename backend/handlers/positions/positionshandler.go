package positions

import (
	"encoding/json"
	"net/http"
	"socialpredict/errors"
	"socialpredict/setup"
	"socialpredict/util"

	"github.com/gorilla/mux"
)

func MarketDBPMPositionsHandler(mcl setup.MarketCreationLoader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		marketIdStr := vars["marketId"]

		// open up database to utilize connection pooling
		db := util.GetDB()

		marketDBPMPositions, err := CalculateMarketPositions_WPAM_DBPM(mcl, db, marketIdStr)
		if errors.HandleHTTPError(w, err, http.StatusBadRequest, "Invalid request or data processing error.") {
			return // Stop execution if there was an error.
		}

		// Respond with the bets display information
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(marketDBPMPositions)
	}
}

func MarketDBPMUserPositionsHandler(mcl setup.MarketCreationLoader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		marketIdStr := vars["marketId"]
		userNameStr := vars["username"]

		// open up database to utilize connection pooling
		db := util.GetDB()

		marketDBPMPositions, err := CalculateMarketPositionForUser_WPAM_DBPM(mcl, db, marketIdStr, userNameStr)
		if errors.HandleHTTPError(w, err, http.StatusBadRequest, "Invalid request or data processing error.") {
			return // Stop execution if there was an error.
		}

		// Respond with the bets display information
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(marketDBPMPositions)
	}
}
