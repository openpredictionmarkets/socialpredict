package marketshandlers

import (
	"encoding/json"
	"log"
	"net/http"
	betshandlers "socialpredict/handlers/bets"
	marketmath "socialpredict/handlers/math/market"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"

	"github.com/gorilla/mux"
)

type MarketPosition struct {
	Username         string
	SharesOwned      uint
	PurchaseProb     float64
	CalculatedPayout float64
}

func MarketPositionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketIdStr := vars["marketId"]
	// Convert marketId to uint
	marketIDUint, err := strconv.ParseUint(marketIdStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	log.Println("marketIdStr: ", marketIdStr)

	// Database connection
	db := util.GetDB()

	var allBetsOnMarket []models.Bet

	// Fetch bets for the market
	allBetsOnMarket = betshandlers.GetBetsForMarket(marketIDUint)

	// get a timeline of probability changes for the market
	// input the market the safe way
	probabilityChanges := marketmath.CalculateMarketProbabilities(market, allBetsOnMarket)

	// calculate number of shares that exist in the entire market, based upon cpmm
	totalShares := marketmath.CalculateTotalShares(allBetsOnMarket, probabilityChanges)

	// sum up the current liquidity pool in the market across all bets, e.g. total volume
	marketVolume := marketmath.GetMarketVolume(allBetsOnMarket)

	// for each individual better in the market, find shares owned.
	// we want shares owned to be a function of the probability bought at
	// such that betters who bought at a probability p when they bought will get a proportionally larger
	// payout when p was further away from the ultimate end resolution.
	// So for example, those who boughtat 0.1 should get a proportionally larger share than those who bet at 0.9
	// for a market that eventually resolves YES.
	var positions []MarketPosition

	bettorShares := make(map[string]uint) // Map to hold the shares owned by each bettor

	// this below way of distributing better shares does not make sense
	// because it's basically saying shares : amount/probability, so lower probabilities automatically get more shares
	for i, bet := range allBetsOnMarket {
		probability := probabilityChanges[i+1].Probability // Using i+1 because the first entry is the initial condition
		if probability > 0 {
			shares := bet.Amount / probability
			bettorShares[bet.Username] += uint(shares)
		}
	}

	finalResolutionProb := marketmath.GetFinalResolutionProbability(marketIDUint)

	for username, shares := range bettorShares {
		initialProb := // Retrieve the initial probability at which the user placed the bet
		// Calculate payout based on how far the initial probability was from the final resolution
		// This is a simplistic formula; you might want to refine it based on your market model
		payoutMultiplier := math.Abs(finalResolutionProb - initialProb)
		payout := float64(shares) * payoutMultiplier

		positions = append(positions, MarketPosition{
			Username:         username,
			SharesOwned:      shares,
			PurchaseProb:     initialProb,
			CalculatedPayout: payout,
		})
	}


	// Respond with the bets display information
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thingToDisplayAtEnd)
}