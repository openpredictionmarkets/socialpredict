package wpam

import (
	"socialpredict/logging"
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
	InitialMarketSubsidization int64
	InitialMarketYes           int64
	InitialMarketNo            int64
	CreateMarketCost           int64
	TraderBonus                int64
	// user stuff
	MaximumDebtAllowed    int64
	InitialAccountBalance int64
	// betting stuff
	MinimumBet    int64
	BetFee        int64
	SellSharesFee int64
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
// See README/README-MATH-PROB-AND-PAYOUT.md#wpam-formula-for-updating-market-probability
func CalculateMarketProbabilitiesWPAM(marketCreatedAtTime time.Time, bets []models.Bet) []ProbabilityChange {

	var probabilityChanges []ProbabilityChange

	// Initial state
	P_initial := appConfig.InitialMarketProbability
	I_initial := appConfig.InitialMarketSubsidization
	totalYes := appConfig.InitialMarketYes
	totalNo := appConfig.InitialMarketNo

	// Add initial state
	probabilityChanges = append(probabilityChanges, ProbabilityChange{Probability: P_initial, Timestamp: marketCreatedAtTime})

	// Calculate probabilities after each bet
	for _, bet := range bets {
		if bet.Outcome == "YES" {
			totalYes += bet.Amount
		} else if bet.Outcome == "NO" {
			totalNo += bet.Amount
		}

		newProbability := (P_initial*float64(I_initial) + float64(totalYes)) / (float64(I_initial) + float64(totalYes) + float64(totalNo))
		probabilityChanges = append(probabilityChanges, ProbabilityChange{Probability: newProbability, Timestamp: bet.PlacedAt})
	}

	logging.LogAnyType(probabilityChanges, "probabilityChanges")

	return probabilityChanges
}
