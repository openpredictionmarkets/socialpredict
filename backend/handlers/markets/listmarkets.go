package marketshandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/handlers/marketpublicresponse"
	marketmath "socialpredict/handlers/math/market"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/handlers/tradingdata"
	usersHandlers "socialpredict/handlers/users"
	"socialpredict/models"
	"socialpredict/setup"
	"socialpredict/util"
	"strconv"

	"gorm.io/gorm"
)

// ListMarketsResponse defines the structure for the list markets response
type ListMarketsResponse struct {
	Markets []MarketOverview `json:"markets"`
}

type MarketOverview struct {
	Market          marketpublicresponse.PublicResponseMarket `json:"market"`
	Creator         usersHandlers.PublicUserType              `json:"creator"`
	LastProbability float64                                   `json:"lastProbability"`
	NumUsers        int                                       `json:"numUsers"`
	TotalVolume     int64                                     `json:"totalVolume"`
}

// ListMarketsHandler handles the HTTP request for listing markets.
func ListMarketsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("ListMarketsHandler: Request received")
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	db := util.GetDB()
	markets, err := ListMarkets(db)
	if err != nil {
		http.Error(w, "Error fetching markets", http.StatusInternalServerError)
		return
	}

	var marketOverviews []MarketOverview
	for _, market := range markets {
		bets := tradingdata.GetBetsForMarket(db, uint(market.ID))
		probabilityChanges := wpam.CalculateMarketProbabilitiesWPAM(setup.MustLoadEconomicsConfig, market.CreatedAt, bets)
		numUsers := usersHandlers.GetNumMarketUsers(bets)
		marketVolume := marketmath.GetMarketVolume(bets)
		lastProbability := probabilityChanges[len(probabilityChanges)-1].Probability

		creatorInfo := usersHandlers.GetPublicUserInfo(db, market.CreatorUsername)

		// return the PublicResponse type with information about the market
		marketIDStr := strconv.FormatUint(uint64(market.ID), 10)
		publicResponseMarket, err := marketpublicresponse.GetPublicResponseMarketByID(db, marketIDStr)
		if err != nil {
			http.Error(w, "Invalid market ID", http.StatusBadRequest)
			return
		}

		marketOverview := MarketOverview{
			Market:          publicResponseMarket,
			Creator:         creatorInfo,
			LastProbability: lastProbability,
			NumUsers:        numUsers,
			TotalVolume:     marketVolume,
		}
		marketOverviews = append(marketOverviews, marketOverview)
	}

	response := ListMarketsResponse{
		Markets: marketOverviews,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ListMarkets fetches a random list of all markets from the database.
func ListMarkets(db *gorm.DB) ([]models.Market, error) {
	var markets []models.Market
	result := db.Order("RANDOM()").Limit(100).Find(&markets) // Set a reasonable limit
	if result.Error != nil {
		log.Printf("Error fetching markets: %v", result.Error)
		return nil, result.Error
	}

	return markets, nil
}
