package wpam

import (
	"log"
	"socialpredict/models"
	"socialpredict/setup"
	"time"
)

type ProbabilityChange struct {
	Probability float64   `json:"probability"`
	Timestamp   time.Time `json:"timestamp"`
}

// appConfig holds the loaded application configuration accessible within the package
var appConfig *setup.EconomicConfig

func init() {
	// Load configuration
	var err error
	appConfig, err = setup.LoadEconomicsConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
}

// CalculateMarketProbabilitiesWPAM calculates and returns the probability changes based on bets.
func CalculateMarketProbabilitiesWPAM(conf *setup.EconomicConfig, marketCreatedAtTime time.Time, bets []models.Bet) []ProbabilityChange {
	var probabilityChanges []ProbabilityChange

	// Initial state using values from appConfig
	P_initial := appConfig.Economics.MarketCreation.InitialMarketProbability
	I_initial := appConfig.Economics.MarketCreation.InitialMarketSubsidization
	totalYes := appConfig.Economics.MarketCreation.InitialMarketYes
	totalNo := appConfig.Economics.MarketCreation.InitialMarketNo

	// Add initial state
	probabilityChanges = append(probabilityChanges, ProbabilityChange{Probability: P_initial, Timestamp: marketCreatedAtTime})

	// Calculate probabilities after each bet
	for _, bet := range bets {
		if bet.Outcome == "YES" {
			totalYes += bet.Amount
		} else if bet.Outcome == "NO" {
			totalNo += bet.Amount
		}

		newProbability := calcProbability(P_initial, I_initial, totalYes, totalNo)
		probabilityChanges = append(probabilityChanges, ProbabilityChange{Probability: newProbability, Timestamp: bet.PlacedAt})
	}

	return probabilityChanges
}

// calcProbability calculates the overall probability given the initial market conditions and the current bet allocation
func calcProbability(initialProbability float64, initialInvestment int64, totalYes int64, totalNo int64) float64 {
	res := (initialProbability*float64(initialInvestment) + float64(totalYes)) / (float64(initialInvestment) + float64(totalYes) + float64(totalNo))
	log.Printf("res: %f, prob: %f, inv: %d, yes: %d, no: %d", res, initialProbability, initialInvestment, totalYes, totalNo)
	return res
}
