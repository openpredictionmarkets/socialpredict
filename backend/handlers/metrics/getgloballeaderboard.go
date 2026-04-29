package metricshandlers

import (
	"net/http"

	"socialpredict/handlers"
)

// GetGlobalLeaderboardHandler returns an application reporting handler for the global leaderboard.
func GetGlobalLeaderboardHandler(svc GlobalLeaderboardService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		leaderboard, err := svc.ComputeGlobalLeaderboard(r.Context())
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		if err := handlers.WriteResult(w, http.StatusOK, leaderboard); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}
