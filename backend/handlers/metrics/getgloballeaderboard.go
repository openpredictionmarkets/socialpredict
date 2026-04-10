package metricshandlers

import (
	"encoding/json"
	"net/http"

	analytics "socialpredict/internal/domain/analytics"
)

// GetGlobalLeaderboardHandler returns an HTTP handler that responds with the global leaderboard.
func GetGlobalLeaderboardHandler(svc *analytics.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		leaderboard, err := svc.ComputeGlobalLeaderboard(r.Context())
		if err != nil {
			http.Error(w, "failed to compute global leaderboard: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(leaderboard); err != nil {
			http.Error(w, "failed to encode leaderboard response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
