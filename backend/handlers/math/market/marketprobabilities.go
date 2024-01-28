package marketmath

import (
	"socialpredict/models"
	"socialpredict/setup"
	"time"
)

type ProbabilityChange struct {
	Probability float64   `json:"probability"`
	Timestamp   time.Time `json:"timestamp"`
}

// AppConfig holds the application-wide configuration
type AppConfig struct {
	InitialMarketProbability   float64
	InitialMarketSubsidization float64
	InitialMarketYes           float64
	InitialMarketNo            float64
	CreateMarketCost           float64
	TraderBonus                float64
	// user stuff
	MaximumDebtAllowed    float64
	InitialAccountBalance float64
	// betting stuff
	MinimumBet    float64
	BetFee        float64
	SellSharesFee float64
}

var appConfig AppConfig

func init() {
	// Load configuration
	config := setup.LoadEconomicsConfig()

	// Populate the appConfig struct
	appConfig = AppConfig{
		// market stuff
		InitialMarketProbability:   config.Economics.MarketCreation.InitialMarketProbability,
		InitialMarketSubsidization: config.Economics.MarketCreation.InitialMarketSubsidization,
		InitialMarketYes:           config.Economics.MarketCreation.InitialMarketYes,
		InitialMarketNo:            config.Economics.MarketCreation.InitialMarketNo,
		CreateMarketCost:           config.Economics.MarketIncentives.CreateMarketCost,
		TraderBonus:                config.Economics.MarketIncentives.TraderBonus,
		// user stuff
		MaximumDebtAllowed:    config.Economics.User.MaximumDebtAllowed,
		InitialAccountBalance: config.Economics.User.InitialAccountBalance,
		// betting stuff
		MinimumBet:    config.Economics.Betting.MinimumBet,
		BetFee:        config.Economics.Betting.BetFee,
		SellSharesFee: config.Economics.Betting.SellSharesFee,
	}
}

// Modify calculateMarketProbabilities to accept bets directly
func CalculateMarketProbabilities(market models.Market, bets []models.Bet) []ProbabilityChange {

	var probabilityChanges []ProbabilityChange

	// Initial state
	P_initial := appConfig.InitialMarketProbability
	I_initial := appConfig.InitialMarketSubsidization
	totalYes := appConfig.InitialMarketYes
	totalNo := appConfig.InitialMarketNo

	// Add initial state
	probabilityChanges = append(probabilityChanges, ProbabilityChange{Probability: P_initial, Timestamp: market.CreatedAt})

	// Calculate probabilities after each bet
	for _, bet := range bets {
		if bet.Outcome == "YES" {
			totalYes += bet.Amount
		} else if bet.Outcome == "NO" {
			totalNo += bet.Amount
		}

		newProbability := (P_initial*I_initial + totalYes) / (I_initial + totalYes + totalNo)
		probabilityChanges = append(probabilityChanges, ProbabilityChange{Probability: newProbability, Timestamp: bet.PlacedAt})
	}

	return probabilityChanges
}
