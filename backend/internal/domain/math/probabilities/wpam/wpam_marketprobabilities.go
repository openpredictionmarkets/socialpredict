package wpam

import (
	"socialpredict/models"
	"time"
)

type ProbabilityChange struct {
	Probability float64   `json:"probability"`
	Timestamp   time.Time `json:"timestamp"`
}

type ProjectedProbability struct {
	Probability float64 `json:"projectedprobability"`
}

// Seeds captures the initial market parameters needed for WPAM calculations.
type Seeds struct {
	InitialProbability     float64
	InitialSubsidization   int64
	InitialYesContribution int64
	InitialNoContribution  int64
}

var (
	configSeeds Seeds
	seedsSet    bool
)

// SetSeeds configures the initial values for WPAM calculations.
func SetSeeds(seeds Seeds) {
	configSeeds = seeds
	seedsSet = true
}

func mustGetSeeds() Seeds {
	if !seedsSet {
		panic("wpam seeds not configured: call wpam.SetSeeds with initial market parameters")
	}
	return configSeeds
}

// CalculateMarketProbabilitiesWPAM calculates and returns the probability changes based on bets.
func CalculateMarketProbabilitiesWPAM(marketCreatedAtTime time.Time, bets []models.Bet) []ProbabilityChange {
	seeds := mustGetSeeds()
	var probabilityChanges []ProbabilityChange

	P_initial := seeds.InitialProbability
	I_initial := seeds.InitialSubsidization
	totalYes := seeds.InitialYesContribution
	totalNo := seeds.InitialNoContribution

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

func ProjectNewProbabilityWPAM(marketCreatedAtTime time.Time, currentBets []models.Bet, newBet models.Bet) ProjectedProbability {
	_ = mustGetSeeds() // ensure seeds set before proceeding

	updatedBets := append(currentBets, newBet)

	probabilityChanges := CalculateMarketProbabilitiesWPAM(marketCreatedAtTime, updatedBets)

	finalProbability := probabilityChanges[len(probabilityChanges)-1].Probability

	return ProjectedProbability{Probability: finalProbability}
}
