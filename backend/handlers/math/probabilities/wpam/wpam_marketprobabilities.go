package wpam

import (
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
func CalculateMarketProbabilitiesWPAM(mcl setup.MarketCreationLoader, marketCreatedAtTime time.Time, bets []models.Bet) []ProbabilityChange {
	var probabilityChanges []ProbabilityChange

	mc := mcl.LoadMarketCreation()
	// Initial state using values from appConfig
	P_initial := mc.InitialMarketProbability
	I_initial := mc.InitialMarketSubsidization
	totalYes := mc.InitialMarketYes
	totalNo := mc.InitialMarketNo

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

	return probabilityChanges
}

func ProjectNewProbabilityWPAM(mcl setup.MarketCreationLoader, marketCreatedAtTime time.Time, currentBets []models.Bet, newBet models.Bet) ProjectedProbability {

	updatedBets := append(currentBets, newBet)

	probabilityChanges := CalculateMarketProbabilitiesWPAM(mcl, marketCreatedAtTime, updatedBets)

	finalProbability := probabilityChanges[len(probabilityChanges)-1].Probability

	return ProjectedProbability{Probability: finalProbability}
}
