package usershandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/handlers/positions"
	"socialpredict/middleware"
	"socialpredict/util"

	"github.com/gorilla/mux"
)

func UserMarketPositionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketId := vars["marketId"]

	// Open up database to utilize connection pooling
	db := util.GetDB()
	user, err := middleware.ValidateTokenAndGetUser(r, db)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	userPosition, err := positions.CalculateMarketPositionForUser_WPAM_DBPM(db, marketId, user.Username)
	if err != nil {
		http.Error(w, "Error calculating user market position: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userPosition)
}
