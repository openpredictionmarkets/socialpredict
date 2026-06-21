package metricshandlers

import (
	"context"
	"net/http"

	"socialpredict/handlers"
	analytics "socialpredict/internal/domain/analytics"
	"socialpredict/internal/domain/readmodels"
)

type SystemMetricsResponse struct {
	analytics.SystemMetrics
	Freshness *FreshnessResponse `json:"freshness,omitempty"`
}

// GetSystemMetricsHandler returns an HTTP handler for application metrics.
// Runtime health and readiness probes are intentionally owned outside this handler.
func GetSystemMetricsHandler(svc SystemMetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, freshness, err := systemMetricsReadModel(r.Context(), svc)
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		response := SystemMetricsResponse{SystemMetrics: *metrics}
		if freshness != nil {
			converted := freshnessResponseFromDomain(*freshness)
			response.Freshness = &converted
		}

		if err := handlers.WriteResult(w, http.StatusOK, response); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}

func systemMetricsReadModel(ctx context.Context, svc SystemMetricsService) (*analytics.SystemMetrics, *readmodels.Freshness, error) {
	readSvc, ok := svc.(SystemMetricsReadModelService)
	if !ok {
		metrics, err := svc.ComputeSystemMetrics(ctx)
		return metrics, nil, err
	}

	readModel, err := readSvc.GetSystemMetricsReadModel(ctx)
	if err == nil && readModel != nil {
		return &readModel.Metrics, &readModel.Freshness, nil
	}

	refreshed, refreshErr := readSvc.RefreshSystemMetricsSnapshot(ctx)
	if refreshErr == nil && refreshed != nil {
		return &refreshed.Metrics, &refreshed.Freshness, nil
	}
	if readModel != nil {
		return &readModel.Metrics, &readModel.Freshness, nil
	}

	metrics, computeErr := svc.ComputeSystemMetrics(ctx)
	if computeErr != nil {
		return nil, nil, computeErr
	}
	return metrics, nil, nil
}
