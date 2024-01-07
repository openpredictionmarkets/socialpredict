package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	betsHandlers "socialpredict/handlers/bets"
	marketMathHandlers "socialpredict/handlers/math/market"
	usersHandlers "socialpredict/handlers/users"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type PublicResponseMarket struct {
	ID                      uint      `json:"id"`
	QuestionTitle           string    `json:"questionTitle"`
	Description             string    `json:"description"`
	OutcomeType             string    `json:"outcomeType"`
	ResolutionDateTime      time.Time `json:"resolutionDateTime"`
	FinalResolutionDateTime time.Time `json:"finalResolutionDateTime"`
	UTCOffset               int       `json:"utcOffset"`
	IsResolved              bool      `json:"isResolved"`
	ResolutionResult        string    `json:"resolutionResult"`
	InitialProbability      float64   `json:"initialProbability"`
	CreatorUsername         string    `json:"creatorUsername"`
}

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
	probabilityChanges := marketMathHandlers.CalculateMarketProbabilities(market, bets)

	numUsers := usersHandlers.GetMarketUsers(bets)
	if err != nil {
		http.Error(w, "Error retrieving number of users.", http.StatusInternalServerError)
		return
	}

	// Inside your handler
	marketVolume := marketMathHandlers.GetMarketVolume(bets)
	if err != nil {
		// Handle error
	}

	// get market creator
	// Fetch the Creator's public information using utility function
	publicCreator := usersHandlers.GetPublicUserInfo(db, market.CreatorUsername)

	// Manually construct the response
	response := struct {
		Market             PublicResponseMarket                   `json:"market"`
		Creator            usersHandlers.PublicUserType           `json:"creator"`
		ProbabilityChanges []marketMathHandlers.ProbabilityChange `json:"probabilityChanges"`
		NumUsers           int                                    `json:"numUsers"`
		TotalVolume        float64                                `json:"totalVolume"`
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
