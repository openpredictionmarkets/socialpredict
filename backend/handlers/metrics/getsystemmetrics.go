package metricshandlers

import (
	"net/http"

	"socialpredict/handlers"
)

// GetSystemMetricsHandler returns an HTTP handler for application metrics.
// Runtime health and readiness probes are intentionally owned outside this handler.
func GetSystemMetricsHandler(svc SystemMetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := svc.ComputeSystemMetrics(r.Context())
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		if err := handlers.WriteResult(w, http.StatusOK, metrics); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}
