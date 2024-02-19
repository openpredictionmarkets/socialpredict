package usershandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/handlers/positions"
	"socialpredict/middleware"
	"socialpredict/util"

	"github.com/gorilla/mux"
)

type UserMarketPositionResponse struct {
	Username    string `json:"username"`
	NetPosition int64  `json:"position"`
}

func UserMarketPositionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketId := vars["marketId"]

	// open up database to utilize connection pooling
	db := util.GetDB()
	user, err := middleware.ValidateTokenAndGetUser(r, db)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	userNetPosition, err := positions.CalculateMarketPositionForUser_WPAM_DBPM(db, marketId, user.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userNetPosition)

}
