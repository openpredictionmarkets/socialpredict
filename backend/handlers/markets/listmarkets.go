package marketshandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/handlers/marketpublicresponse"
	marketmath "socialpredict/handlers/math/market"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/handlers/tradingdata"
	"socialpredict/handlers/users/publicuser"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"
)

// MarketService interface defines methods for market operations
// This will be injected from the domain layer
type MarketService interface {
	ListMarkets() ([]models.Market, error)
}

// DefaultMarketService implements MarketService using existing functionality
// This is a temporary bridge to avoid breaking changes
type DefaultMarketService struct{}

func (s *DefaultMarketService) ListMarkets() ([]models.Market, error) {
	db := util.GetDB()
	var markets []models.Market
	result := db.Order("RANDOM()").Limit(100).Find(&markets)
	if result.Error != nil {
		log.Printf("Error fetching markets: %v", result.Error)
		return nil, result.Error
	}
	return markets, nil
}

// ListMarketsResponse defines the structure for the list markets response
type ListMarketsResponse struct {
	Markets []MarketOverview `json:"markets"`
}

type MarketOverview struct {
	Market          marketpublicresponse.PublicResponseMarket `json:"market"`
	Creator         models.PublicUser                         `json:"creator"`
	LastProbability float64                                   `json:"lastProbability"`
	NumUsers        int                                       `json:"numUsers"`
	TotalVolume     int64                                     `json:"totalVolume"`
}

// listMarketsService holds the service instance
var listMarketsService MarketService = &DefaultMarketService{}

// SetListMarketsService allows injecting a custom service (for testing or new architecture)
func SetListMarketsService(service MarketService) {
	listMarketsService = service
}

// ListMarketsHandler handles the HTTP request for listing markets.
func ListMarketsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("ListMarketsHandler: Request received")
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	markets, err := listMarketsService.ListMarkets()
	if err != nil {
		http.Error(w, "Error fetching markets", http.StatusInternalServerError)
		return
	}

	var marketOverviews []MarketOverview = make([]MarketOverview, 0)
	db := util.GetDB() // Still needed for complex calculations - will be refactored later

	for _, market := range markets {
		bets := tradingdata.GetBetsForMarket(db, uint(market.ID))
		probabilityChanges := wpam.CalculateMarketProbabilitiesWPAM(market.CreatedAt, bets)
		numUsers := models.GetNumMarketUsers(bets)
		marketVolume := marketmath.GetMarketVolume(bets)
		lastProbability := probabilityChanges[len(probabilityChanges)-1].Probability

		creatorInfo := publicuser.GetPublicUserInfo(db, market.CreatorUsername)

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
