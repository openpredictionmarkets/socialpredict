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

// SeedProvider supplies seeds for probability calculations.
type SeedProvider interface {
	Seeds() Seeds
}

// StaticSeedProvider returns a fixed seed configuration.
type StaticSeedProvider struct {
	Value Seeds
}

func (p StaticSeedProvider) Seeds() Seeds { return p.Value }

// ProbabilityCalculator performs WPAM probability calculations using supplied seeds.
type ProbabilityCalculator struct {
	seeds SeedProvider
}

// NewProbabilityCalculator constructs a calculator with the provided seed source.
// If provider is nil, sensible defaults are used.
func NewProbabilityCalculator(provider SeedProvider) ProbabilityCalculator {
	if provider == nil {
		provider = StaticSeedProvider{
			Value: Seeds{
				InitialProbability:   0.5,
				InitialSubsidization: 1,
			},
		}
	}
	return ProbabilityCalculator{seeds: provider}
}

// Seeds returns the configured seeds for the calculator.
func (c ProbabilityCalculator) Seeds() Seeds {
	if c.seeds == nil {
		return Seeds{}
	}
	return c.seeds.Seeds()
}

// CalculateMarketProbabilitiesWPAM calculates and returns the probability changes based on bets.
func CalculateMarketProbabilitiesWPAM(marketCreatedAtTime time.Time, bets []models.Bet) []ProbabilityChange {
	return NewProbabilityCalculator(nil).CalculateMarketProbabilitiesWPAM(marketCreatedAtTime, bets)
}

// CalculateMarketProbabilitiesWPAM calculates and returns the probability changes based on bets using the calculator seeds.
func (c ProbabilityCalculator) CalculateMarketProbabilitiesWPAM(marketCreatedAtTime time.Time, bets []models.Bet) []ProbabilityChange {
	seeds := c.seeds.Seeds()
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
	return NewProbabilityCalculator(nil).ProjectNewProbabilityWPAM(marketCreatedAtTime, currentBets, newBet)
}

// ProjectNewProbabilityWPAM projects the probability after a new bet using calculator seeds.
func (c ProbabilityCalculator) ProjectNewProbabilityWPAM(marketCreatedAtTime time.Time, currentBets []models.Bet, newBet models.Bet) ProjectedProbability {
	updatedBets := append(currentBets, newBet)
	probabilityChanges := c.CalculateMarketProbabilitiesWPAM(marketCreatedAtTime, updatedBets)
	finalProbability := probabilityChanges[len(probabilityChanges)-1].Probability

	return ProjectedProbability{Probability: finalProbability}
}
