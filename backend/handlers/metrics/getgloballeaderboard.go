package metricshandlers

import (
	"net/http"
	"strconv"

	"socialpredict/handlers"
)

// GetGlobalLeaderboardHandler returns an application reporting handler for the global leaderboard.
func GetGlobalLeaderboardHandler(svc GlobalLeaderboardService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot, err := svc.ComputeGlobalLeaderboardSnapshot(r.Context())
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		limit, offset := parseLeaderboardPage(r)
		if err := handlers.WriteResult(w, http.StatusOK, snapshot.ResultPage(limit, offset)); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}

func parseLeaderboardPage(r *http.Request) (int, int) {
	query := r.URL.Query()
	limit := 20
	if raw := query.Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	offset := 0
	if raw := query.Get("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			offset = parsed
		}
	}
	return limit, offset
}
