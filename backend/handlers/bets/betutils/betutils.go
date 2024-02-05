package betutils

import (
	"errors"
	"socialpredict/models"
	"socialpredict/setup"
	"time"

	"gorm.io/gorm"
)

// CheckMarketStatus checks if the market is resolved or closed.
// It returns an error if the market is not suitable for placing a bet.
func CheckMarketStatus(db *gorm.DB, marketID int64) error {
	var market models.Market
	if result := db.First(&market, marketID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("market not found")
		}
		return errors.New("error fetching market")
	}

	if market.IsResolved {
		return errors.New("cannot place a bet on a resolved market")
	}

	if time.Now().After(market.ResolutionDateTime) {
		return errors.New("cannot place a bet on a closed market")
	}

	return nil
}

// AppConfig holds the application-wide configuration
type AppConfig struct {
	InitialMarketProbability   float64
	InitialMarketSubsidization float64
	// user stuff
	MaximumDebtAllowed    float64
	InitialAccountBalance float64
	// betting stuff
	MinimumBet    float64
	BetFee        float64
	SellSharesFee float64
}

var Appconfig AppConfig

func init() {
	// Load configuration
	config := setup.LoadEconomicsConfig()

	// Populate the appConfig struct
	Appconfig = AppConfig{
		// market stuff
		InitialMarketProbability:   config.Economics.MarketCreation.InitialMarketProbability,
		InitialMarketSubsidization: config.Economics.MarketCreation.InitialMarketSubsidization,
		// user stuff
		MaximumDebtAllowed:    config.Economics.User.MaximumDebtAllowed,
		InitialAccountBalance: config.Economics.User.InitialAccountBalance,
		// betting stuff
		MinimumBet:    config.Economics.Betting.MinimumBet,
		BetFee:        config.Economics.Betting.BetFee,
		SellSharesFee: config.Economics.Betting.SellSharesFee,
	}
}
