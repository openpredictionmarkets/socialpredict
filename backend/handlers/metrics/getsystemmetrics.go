package metricshandlers

import (
	"net/http"

	"socialpredict/handlers"
	analytics "socialpredict/internal/domain/analytics"
)

// GetSystemMetricsHandler returns an HTTP handler that emits system metrics via the analytics service.
func GetSystemMetricsHandler(svc *analytics.Service) http.HandlerFunc {
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
