package positions

import (
	"encoding/json"
	"net/http"

	"socialpredict/errors"
	"socialpredict/util"

	"github.com/gorilla/mux"
)

func MarketDBPMCombinedPositionValuationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketIdStr := vars["marketId"]

	// open up database to utilize connection pooling
	db := util.GetDB()

	response, err := CalculateUserPositionWithValuationResponse(db, marketIdStr)
	if errors.HandleHTTPError(w, err, http.StatusInternalServerError, "Failed to calculate position valuations.") {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
