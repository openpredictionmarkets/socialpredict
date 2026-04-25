package metricshandlers

import (
	"net/http"

	"socialpredict/handlers"
	analytics "socialpredict/internal/domain/analytics"
)

// GetGlobalLeaderboardHandler returns an HTTP handler that responds with the global leaderboard.
func GetGlobalLeaderboardHandler(svc *analytics.Service) http.HandlerFunc {
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
