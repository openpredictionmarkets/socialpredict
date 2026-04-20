package setup

import (
	"encoding/json"
	"net/http"

	configsvc "socialpredict/internal/service/config"
)

func GetSetupHandler(configService configsvc.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if configService == nil {
			http.Error(w, "Failed to load economic config", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(configService.Economics())
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

func GetFrontendSetupHandler(configService configsvc.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if configService == nil {
			http.Error(w, "Failed to load frontend config", http.StatusInternalServerError)
			return
		}

		response := frontendConfigResponse{
			Charts: frontendChartsResponse{
				SigFigs: configService.ChartSigFigs(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode frontend setup response", http.StatusInternalServerError)
			return
		}
	}
}
