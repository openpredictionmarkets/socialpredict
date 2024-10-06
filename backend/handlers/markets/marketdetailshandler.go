package marketshandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/handlers/marketpublicresponse"
	marketmath "socialpredict/handlers/math/market"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/handlers/tradingdata"
	usersHandlers "socialpredict/handlers/users"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"

	"github.com/gorilla/mux"
)

// MarketDetailResponse defines the structure for the market detail response
type MarketDetailHandlerResponse struct {
	Market             marketpublicresponse.PublicResponseMarket `json:"market"`
	Creator            models.PublicUser                         `json:"creator"`
	ProbabilityChanges []wpam.ProbabilityChange                  `json:"probabilityChanges"`
	NumUsers           int                                       `json:"numUsers"`
	TotalVolume        int64                                     `json:"totalVolume"`
}

func MarketDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketId := vars["marketId"]

	// Parsing a String to an Unsigned Integer, base10, 64bits
	marketIDUint64, err := strconv.ParseUint(marketId, 10, 64)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	marketIDUint := uint(marketIDUint64)

	// open up database to utilize connection pooling
	db := util.GetDB()

	// Fetch all bets for the market
	bets := tradingdata.GetBetsForMarket(db, marketIDUint)

	// return the PublicResponse type with information about the market
	publicResponseMarket, err := marketpublicresponse.GetPublicResponseMarketByID(db, marketId)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Calculate probabilities using the fetched bets
	probabilityChanges := wpam.CalculateMarketProbabilitiesWPAM(publicResponseMarket.CreatedAt, bets)

	// find the number of users on the market
	numUsers := models.GetNumMarketUsers(bets)

	// market volume is equivalent to the sum of all bets
	marketVolume := marketmath.GetMarketVolume(bets)
	if err != nil {
		// Handle error
	}

	// get market creator
	// Fetch the Creator's public information using utility function
	publicCreator := usersHandlers.GetPublicUserInfo(db, publicResponseMarket.CreatorUsername)

	// Manually construct the response
	response := MarketDetailHandlerResponse{
		Market:             publicResponseMarket,
		Creator:            publicCreator,
		ProbabilityChanges: probabilityChanges,
		NumUsers:           numUsers,
		TotalVolume:        marketVolume,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
