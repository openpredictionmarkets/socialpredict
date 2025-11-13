package setup

import (
	"encoding/json"
	"net/http"
	"socialpredict/setup"
)

func GetSetupHandler(loadEconomicsConfig func() (*setup.EconomicConfig, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		appConfig, err := loadEconomicsConfig()
		if err != nil {
			http.Error(w, "Failed to load economic config", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(appConfig.Economics)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

type frontendChartsResponse struct {
	SigFigs int `json:"sigFigs"`
}

type frontendConfigResponse struct {
	Charts frontendChartsResponse `json:"charts"`
}

func GetFrontendSetupHandler(loadEconomicsConfig func() (*setup.EconomicConfig, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := loadEconomicsConfig()
		if err != nil {
			http.Error(w, "Failed to load frontend config", http.StatusInternalServerError)
			return
		}

		response := frontendConfigResponse{
			Charts: frontendChartsResponse{
				SigFigs: setup.ChartSigFigs(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode frontend setup response", http.StatusInternalServerError)
			return
		}
	}
}
