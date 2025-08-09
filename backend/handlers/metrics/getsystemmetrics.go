package metricshandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/handlers/math/financials"
	"socialpredict/setup"
	"socialpredict/util"
)

func GetSystemMetricsHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()
	load := setup.EconomicsConfig // matches EconConfigLoader (func() *EconomicConfig)

	res, err := financials.ComputeSystemMetrics(db, load)
	if err != nil {
		http.Error(w, "failed to compute metrics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode metrics response: "+err.Error(), http.StatusInternalServerError)
	}
}
