package metricshandlers

import (
	"encoding/json"
	"net/http"

	analytics "socialpredict/internal/domain/analytics"
)

// GetSystemMetricsHandler returns an HTTP handler that emits system metrics via the analytics service.
func GetSystemMetricsHandler(svc *analytics.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := svc.ComputeSystemMetrics(r.Context())
		if err != nil {
			http.Error(w, "failed to compute metrics: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			http.Error(w, "Failed to encode metrics response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
