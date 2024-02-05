package marketshandlers

import (
	"encoding/json"
	"errors"
	"net/http"
	betsHandlers "socialpredict/handlers/bets"
	marketmath "socialpredict/handlers/math/market"
	"socialpredict/handlers/math/probabilities/wpam"
	usersHandlers "socialpredict/handlers/users"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func MarketDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketId := vars["marketId"]

	// Use database connection
	db := util.GetDB()

	var market models.Market
	// Fetch the Market without preloading the Creator
	result := db.Where("ID = ?", marketId).First(&market)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Market not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching market", http.StatusInternalServerError)
		}
		return
	}

	// Construct ResponseMarket from the market model
	responseMarket := PublicResponseMarket{
		ID:                      market.ID,
		QuestionTitle:           market.QuestionTitle,
		Description:             market.Description,
		OutcomeType:             market.OutcomeType,
		ResolutionDateTime:      market.ResolutionDateTime,
		FinalResolutionDateTime: market.FinalResolutionDateTime,
		UTCOffset:               market.UTCOffset,
		IsResolved:              market.IsResolved,
		ResolutionResult:        market.ResolutionResult,
		InitialProbability:      market.InitialProbability,
		CreatorUsername:         market.CreatorUsername,
	}

	// Parsing a String to an Unsigned Integer, base10, 32bits
	marketIDUint, err := strconv.ParseUint(marketId, 10, 32)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Fetch all bets for the market once
	bets := betsHandlers.GetBetsForMarket(marketIDUint)

	// Calculate probabilities using the fetched bets
	probabilityChanges := wpam.CalculateMarketProbabilitiesWPAM(market, bets)

	numUsers := usersHandlers.GetNumMarketUsers(bets)
	if err != nil {
		http.Error(w, "Error retrieving number of users.", http.StatusInternalServerError)
		return
	}

	// Inside your handler
	marketVolume := marketmath.GetMarketVolume(bets)
	if err != nil {
		// Handle error
	}

	// get market creator
	// Fetch the Creator's public information using utility function
	publicCreator := usersHandlers.GetPublicUserInfo(db, market.CreatorUsername)

	// Manually construct the response
	response := struct {
		Market             PublicResponseMarket         `json:"market"`
		Creator            usersHandlers.PublicUserType `json:"creator"`
		ProbabilityChanges []wpam.ProbabilityChange     `json:"probabilityChanges"`
		NumUsers           int                          `json:"numUsers"`
		TotalVolume        int64                        `json:"totalVolume"`
	}{
		Market:             responseMarket,
		Creator:            publicCreator,
		ProbabilityChanges: probabilityChanges,
		NumUsers:           numUsers,
		TotalVolume:        marketVolume,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
