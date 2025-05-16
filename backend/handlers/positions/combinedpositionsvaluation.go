package positions

import (
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/handlers/tradingdata"
	"socialpredict/models"

	"socialpredict/handlers/math/market/marketmath"

	"gorm.io/gorm"
)

type UserPositionWithValuation struct {
	Username       string `json:"username"`
	YesSharesOwned int64  `json:"yesSharesOwned"`
	NoSharesOwned  int64  `json:"noSharesOwned"`
	Value          int64  `json:"value"`
}

func CalculateUserPositionWithValuationResponse(db *gorm.DB, marketIdStr string) ([]UserPositionWithValuation, error) {

	var market models.Market
	if err := db.First(&market, marketIdStr).Error; err != nil {
		return nil, err
	}

	// Step 1: Get user positions
	marketDBPMPositions, err := CalculateMarketPositions_WPAM_DBPM(db, marketIdStr)

	// Step 2: Current WPAM probability)
	currentProbability, err := wpam.GetCurrentProbabilityFromMarketAndBets(db, market)
	if err {
		return nil, err
	}

	// Step 4: Calculate valuations
	valuations := CalculateRoundedUserValuationsFromUserMarketPositions(
		marketDBPMPositions,
		currentProbability,
		marketmath.GetMarketVolume(tradingdata.GetBetsForMarket(db, uint(market.ID))),
	)

	// Step 5: Merge
	var result []UserPositionWithValuation
	for username, pos := range marketDBPMPositions {
		val := valuations[username]
		result = append(result, UserPositionWithValuation{
			Username:       username,
			YesSharesOwned: pos.YesSharesOwned,
			NoSharesOwned:  pos.NoSharesOwned,
			Value:          val.RoundedValue,
		})
	}

	return result, nil
}
