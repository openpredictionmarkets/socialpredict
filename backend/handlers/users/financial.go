package usershandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/handlers/math/financials"
	"socialpredict/handlers/users/publicuser"
	"socialpredict/setup"
	"socialpredict/util"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// GetUserFinancialHandlerWithDB returns a handler function with injected database connection
// This follows the higher-order function pattern used elsewhere in the codebase
func GetUserFinancialHandlerWithDB(db *gorm.DB, econConfigLoader func() (*setup.EconomicConfig, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		// Extract username from URL parameter
		vars := mux.Vars(r)
		username := vars["username"]

		if username == "" {
			http.Error(w, "Username parameter is required", http.StatusBadRequest)
			return
		}

		// Get user's public information to extract account balance
		userPublicInfo := publicuser.GetPublicUserInfo(db, username)

		// Load economic configuration
		econ, err := econConfigLoader()
		if err != nil {
			log.Printf("Error loading economic config: %v", err)
			http.Error(w, "Unable to load configuration", http.StatusInternalServerError)
			return
		}

		// Compute comprehensive financial snapshot
		snapshot, err := financials.ComputeUserFinancials(db, username, userPublicInfo.AccountBalance, econ)
		if err != nil {
			log.Printf("Error generating user financial snapshot: %v", err)
			http.Error(w, "Unable to generate financial snapshot", http.StatusInternalServerError)
			return
		}

		// Return financial data as JSON
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"financial": snapshot,
		}

		json.NewEncoder(w).Encode(response)
	}
}

// GetUserFinancialHandler returns comprehensive financial metrics for a user
// Endpoint: GET /v0/users/{username}/financial
// This is the production version that uses the actual database and config loader
func GetUserFinancialHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()
	handler := GetUserFinancialHandlerWithDB(db, setup.LoadEconomicsConfig)
	handler(w, r)
}
