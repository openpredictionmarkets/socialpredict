package setup

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"socialpredict/setup"
)

// isPlatformPrivate reads the PLATFORM_PRIVATE env var (default false).
func isPlatformPrivate() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("PLATFORM_PRIVATE")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

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
	Charts          frontendChartsResponse `json:"charts"`
	PlatformPrivate bool                   `json:"platformPrivate"`
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
			PlatformPrivate: isPlatformPrivate(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode frontend setup response", http.StatusInternalServerError)
			return
		}
	}
}
