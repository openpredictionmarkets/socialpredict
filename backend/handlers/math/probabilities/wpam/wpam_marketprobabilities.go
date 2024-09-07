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

type ProjectedProbability struct {
	Probability float64 `json:"projectedprobability"`
}

// CalculateMarketProbabilitiesWPAM calculates and returns the probability changes based on bets.
func CalculateMarketProbabilitiesWPAM(loadEconomicsConfig setup.EconConfigLoader, marketCreatedAtTime time.Time, bets []models.Bet) []ProbabilityChange {
	// TODO: decision: this technically only needs a portion of the config: loadEconomicsConfig().Economics.MarketCreation. We should consider another abstraction for this
	appConfig := loadEconomicsConfig()
	var probabilityChanges []ProbabilityChange

	// Initial state using values from appConfig
	P_initial := appConfig.Economics.MarketCreation.InitialMarketProbability
	I_initial := appConfig.Economics.MarketCreation.InitialMarketSubsidization
	totalYes := appConfig.Economics.MarketCreation.InitialMarketYes
	totalNo := appConfig.Economics.MarketCreation.InitialMarketNo

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

// ProjectNewProbabilityWPAM determines what the probability would be if a newBet is placed
func ProjectNewProbabilityWPAM(loadEconomicsConfig setup.EconConfigLoader, marketCreatedAtTime time.Time, currentBets []models.Bet, newBet models.Bet) ProjectedProbability {
	updatedBets := append(currentBets, newBet)
	probabilityChanges := CalculateMarketProbabilitiesWPAM(loadEconomicsConfig, marketCreatedAtTime, updatedBets)
	finalProbability := probabilityChanges[len(probabilityChanges)-1].Probability
	return ProjectedProbability{Probability: finalProbability}
}
