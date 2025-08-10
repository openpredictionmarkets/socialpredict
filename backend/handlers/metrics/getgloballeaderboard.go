package metricshandlers

import (
	"encoding/json"
	"net/http"
	positionsmath "socialpredict/handlers/math/positions"
	"socialpredict/util"
)

func GetGlobalLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()

	leaderboard, err := positionsmath.CalculateGlobalLeaderboard(db)
	if err != nil {
		http.Error(w, "failed to compute global leaderboard: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(leaderboard); err != nil {
		http.Error(w, "Failed to encode leaderboard response: "+err.Error(), http.StatusInternalServerError)
	}
}
