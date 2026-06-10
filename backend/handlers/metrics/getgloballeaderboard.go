package metricshandlers

import (
	"context"
	"net/http"
	"strconv"

	"socialpredict/handlers"
	analytics "socialpredict/internal/domain/analytics"
	"socialpredict/internal/domain/readmodels"
)

type GlobalLeaderboardResponse struct {
	Entries   []analytics.GlobalUserProfitability `json:"entries"`
	Freshness *FreshnessResponse                  `json:"freshness,omitempty"`
}

// GetGlobalLeaderboardHandler returns an application reporting handler for the global leaderboard.
func GetGlobalLeaderboardHandler(svc GlobalLeaderboardService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := parseLeaderboardPage(r)
		entries, freshness, err := globalLeaderboardReadModel(r.Context(), svc, limit, offset)
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		response := GlobalLeaderboardResponse{Entries: entries}
		if freshness != nil {
			converted := freshnessResponseFromDomain(*freshness)
			response.Freshness = &converted
		}
		if err := handlers.WriteResult(w, http.StatusOK, response); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}

func globalLeaderboardReadModel(ctx context.Context, svc GlobalLeaderboardService, limit int, offset int) ([]analytics.GlobalUserProfitability, *readmodels.Freshness, error) {
	readSvc, ok := svc.(GlobalLeaderboardReadModelService)
	if !ok {
		snapshot, err := svc.ComputeGlobalLeaderboardSnapshot(ctx)
		if err != nil {
			return nil, nil, err
		}
		return snapshot.ResultPage(limit, offset), nil, nil
	}

	readModel, err := readSvc.GetGlobalLeaderboardReadModel(ctx, limit, offset)
	if err == nil && readModel != nil && !freshnessExpired(readModel.Freshness, analytics.GlobalLeaderboardSnapshotTargetFreshness) {
		return readModel.Entries, &readModel.Freshness, nil
	}

	refreshed, refreshErr := readSvc.RefreshGlobalLeaderboardSnapshot(ctx)
	if refreshErr == nil && refreshed != nil {
		paged := (&analytics.GlobalLeaderboardSnapshot{Entries: refreshed.Entries}).ResultPage(limit, offset)
		return paged, &refreshed.Freshness, nil
	}
	if readModel != nil {
		return readModel.Entries, &readModel.Freshness, nil
	}

	snapshot, computeErr := svc.ComputeGlobalLeaderboardSnapshot(ctx)
	if computeErr != nil {
		return nil, nil, computeErr
	}
	return snapshot.ResultPage(limit, offset), nil, nil
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
