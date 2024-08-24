package setuphandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/setup"
)

func GetSetupHandler(w http.ResponseWriter, r *http.Request) {
	appConfig, err := setup.LoadEconomicsConfig()
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
