package marketshandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/errors"
	positionsmath "socialpredict/handlers/math/positions"
	"socialpredict/util"

	"github.com/gorilla/mux"
)

// MarketLeaderboardHandler handles requests for market profitability leaderboards
func MarketLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketIdStr := vars["marketId"]

	// Set content type header early to ensure it's always set
	w.Header().Set("Content-Type", "application/json")

	// Open up database to utilize connection pooling
	db := util.GetDB()

	leaderboard, err := positionsmath.CalculateMarketLeaderboard(db, marketIdStr)
	if errors.HandleHTTPError(w, err, http.StatusBadRequest, "Invalid request or data processing error.") {
		return // Stop execution if there was an error.
	}

	// Respond with the leaderboard information
	json.NewEncoder(w).Encode(leaderboard)
}
