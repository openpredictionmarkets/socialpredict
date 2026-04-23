package wpam

import (
	"time"

	"socialpredict/internal/domain/boundary"
)

const (
	wpamOutcomeYes = "YES"
	wpamOutcomeNo  = "NO"
)

type ProbabilityChange struct {
	Probability float64   `json:"probability"`
	Timestamp   time.Time `json:"timestamp"`
}

type ProjectedProbability struct {
	Probability float64 `json:"projectedprobability"`
}

type ProbabilityFormula interface {
	Calculate(seeds Seeds, totalYes, totalNo int64) float64
}

type ContributionAccumulator func(*marketContributions, int64)

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

type marketContributions struct {
	yes int64
	no  int64
}

type weightedAverageProbabilityFormula struct{}

var defaultSeedProvider = StaticSeedProvider{
	Value: Seeds{
		InitialProbability:   0.5,
		InitialSubsidization: 1,
	},
}

var defaultContributionAccumulators = map[string]ContributionAccumulator{
	wpamOutcomeYes: func(contributions *marketContributions, amount int64) { contributions.yes += amount },
	wpamOutcomeNo:  func(contributions *marketContributions, amount int64) { contributions.no += amount },
}

// ProbabilityCalculator performs WPAM probability calculations using supplied seeds.
type ProbabilityCalculator struct {
	seeds        SeedProvider
	formula      ProbabilityFormula
	accumulators map[string]ContributionAccumulator
}

type ProbabilityCalculatorOption func(*ProbabilityCalculator)

// WithProbabilityFormula overrides the formula used for probability calculations.
func WithProbabilityFormula(formula ProbabilityFormula) ProbabilityCalculatorOption {
	return func(c *ProbabilityCalculator) {
		if formula != nil {
			c.formula = formula
		}
	}
}

// WithContributionAccumulators overrides outcome contribution handling.
func WithContributionAccumulators(accumulators map[string]ContributionAccumulator) ProbabilityCalculatorOption {
	return func(c *ProbabilityCalculator) {
		if accumulators == nil {
			return
		}
		c.accumulators = accumulators
	}
}

// NewProbabilityCalculator constructs a calculator with the provided seed source.
// If provider is nil, sensible defaults are used.
func NewProbabilityCalculator(provider SeedProvider) ProbabilityCalculator {
	return NewProbabilityCalculatorWithOptions(provider)
}

// NewProbabilityCalculatorWithOptions constructs a calculator with substitutable probability behavior.
func NewProbabilityCalculatorWithOptions(provider SeedProvider, opts ...ProbabilityCalculatorOption) ProbabilityCalculator {
	if provider == nil {
		provider = defaultSeedProvider
	}
	calculator := ProbabilityCalculator{
		seeds:        provider,
		formula:      weightedAverageProbabilityFormula{},
		accumulators: defaultContributionAccumulators,
	}
	for _, opt := range opts {
		opt(&calculator)
	}
	return calculator
}

// Seeds returns the configured seeds for the calculator.
func (c ProbabilityCalculator) Seeds() Seeds {
	return c.withDefaults().seeds.Seeds()
}

// CalculateMarketProbabilitiesWPAM calculates and returns the probability changes based on bets.
func CalculateMarketProbabilitiesWPAM(marketCreatedAtTime time.Time, bets []boundary.Bet) []ProbabilityChange {
	return NewProbabilityCalculator(nil).CalculateMarketProbabilitiesWPAM(marketCreatedAtTime, bets)
}

// CalculateMarketProbabilitiesWPAM calculates and returns the probability changes based on bets using the calculator seeds.
func (c ProbabilityCalculator) CalculateMarketProbabilitiesWPAM(marketCreatedAtTime time.Time, bets []boundary.Bet) []ProbabilityChange {
	calculator := c.withDefaults()
	seeds := calculator.seeds.Seeds()
	var probabilityChanges []ProbabilityChange

	P_initial := seeds.InitialProbability

	probabilityChanges = append(probabilityChanges, ProbabilityChange{Probability: P_initial, Timestamp: marketCreatedAtTime})

	// Calculate probabilities after each bet
	contributions := marketContributions{
		yes: seeds.InitialYesContribution,
		no:  seeds.InitialNoContribution,
	}
	for _, bet := range bets {
		calculator.applyContribution(&contributions, bet)

		newProbability := calculator.formula.Calculate(seeds, contributions.yes, contributions.no)
		probabilityChanges = append(probabilityChanges, ProbabilityChange{Probability: newProbability, Timestamp: bet.PlacedAt})
	}

	return probabilityChanges
}

func ProjectNewProbabilityWPAM(marketCreatedAtTime time.Time, currentBets []boundary.Bet, newBet boundary.Bet) ProjectedProbability {
	return NewProbabilityCalculator(nil).ProjectNewProbabilityWPAM(marketCreatedAtTime, currentBets, newBet)
}

// ProjectNewProbabilityWPAM projects the probability after a new bet using calculator seeds.
func (c ProbabilityCalculator) ProjectNewProbabilityWPAM(marketCreatedAtTime time.Time, currentBets []boundary.Bet, newBet boundary.Bet) ProjectedProbability {
	updatedBets := append(append([]boundary.Bet(nil), currentBets...), newBet)
	probabilityChanges := c.withDefaults().CalculateMarketProbabilitiesWPAM(marketCreatedAtTime, updatedBets)
	finalProbability := probabilityChanges[len(probabilityChanges)-1].Probability

	return ProjectedProbability{Probability: finalProbability}
}

func (c ProbabilityCalculator) withDefaults() ProbabilityCalculator {
	if c.seeds == nil {
		c.seeds = defaultSeedProvider
	}
	if c.formula == nil {
		c.formula = weightedAverageProbabilityFormula{}
	}
	if c.accumulators == nil {
		c.accumulators = defaultContributionAccumulators
	}
	return c
}

func (c ProbabilityCalculator) applyContribution(contributions *marketContributions, bet boundary.Bet) {
	if accumulate, ok := c.accumulators[bet.Outcome]; ok {
		accumulate(contributions, bet.Amount)
	}
}

func (weightedAverageProbabilityFormula) Calculate(seeds Seeds, totalYes, totalNo int64) float64 {
	return (seeds.InitialProbability*float64(seeds.InitialSubsidization) + float64(totalYes)) /
		(float64(seeds.InitialSubsidization) + float64(totalYes) + float64(totalNo))
}
