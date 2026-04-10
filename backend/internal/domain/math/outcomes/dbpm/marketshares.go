package dbpm

import (
	"math"
	marketmath "socialpredict/internal/domain/math/market"
	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models"
)

const (
	dbpmOutcomeYes = "YES"
	dbpmOutcomeNo  = "NO"
)

type poolShareWeight func(probability float64) float64
type normalizationAccumulator func(*normalizationSums, float64)
type normalizationFactorSelector func(normalizationFactors) float64
type payoutAggregator func(*DBPMMarketPosition, int64)
type exposureAccumulator func(*sideExposure, int64)
type probabilitySelector interface {
	Current([]wpam.ProbabilityChange) float64
}

type volumeCalculator interface {
	Volume([]models.Bet) int64
}

type coursePayoutCalculator interface {
	CoursePayouts([]models.Bet, []wpam.ProbabilityChange) []CourseBetPayout
}

type normalizationFactorCalculator interface {
	NormalizationFactors(int64, int64, []CourseBetPayout) (float64, float64)
}

type scaledPayoutCalculator interface {
	ScaledPayouts([]models.Bet, []CourseBetPayout, float64, float64) []int64
}

type excessCalculator interface {
	Excess([]models.Bet, []int64) int64
}

type payoutAdjuster interface {
	AdjustPositive([]int64, int64) []int64
	AdjustNegative([]int64, int64) []int64
	Adjust([]models.Bet, []int64) []int64
}

type userPayoutAggregator interface {
	Aggregate([]models.Bet, []int64) []DBPMMarketPosition
}

type marketPositionNetter interface {
	Net([]DBPMMarketPosition) []DBPMMarketPosition
}

type singleCreditAllocator interface {
	Allocate([]models.Bet) (int64, int64)
}

type normalizationSums struct {
	yes float64
	no  float64
}

type normalizationFactors struct {
	yes float64
	no  float64
}

type sideExposure struct {
	yes int64
	no  int64
}

var dbpmPoolShareWeights = map[string]poolShareWeight{
	dbpmOutcomeYes: func(probability float64) float64 { return probability },
	dbpmOutcomeNo:  func(probability float64) float64 { return 1 - probability },
}

var dbpmNormalizationAccumulators = map[string]normalizationAccumulator{
	dbpmOutcomeYes: func(sums *normalizationSums, payout float64) { sums.yes += payout },
	dbpmOutcomeNo:  func(sums *normalizationSums, payout float64) { sums.no += payout },
}

var dbpmNormalizationSelectors = map[string]normalizationFactorSelector{
	dbpmOutcomeYes: func(factors normalizationFactors) float64 { return factors.yes },
	dbpmOutcomeNo:  func(factors normalizationFactors) float64 { return factors.no },
}

var dbpmPayoutAggregators = map[string]payoutAggregator{
	dbpmOutcomeYes: func(position *DBPMMarketPosition, payout int64) { position.YesSharesOwned += payout },
	dbpmOutcomeNo:  func(position *DBPMMarketPosition, payout int64) { position.NoSharesOwned += payout },
}

var dbpmExposureAccumulators = map[string]exposureAccumulator{
	dbpmOutcomeYes: func(exposure *sideExposure, amount int64) { exposure.yes += amount },
	dbpmOutcomeNo:  func(exposure *sideExposure, amount int64) { exposure.no += amount },
}

type wpamProbabilitySelector struct{}

func (wpamProbabilitySelector) Current(changes []wpam.ProbabilityChange) float64 {
	return wpam.GetCurrentProbability(changes)
}

type marketVolumeCalculator struct{}

func (marketVolumeCalculator) Volume(bets []models.Bet) int64 {
	return marketmath.GetMarketVolume(bets)
}

type MarketShareCalculator struct {
	probabilities probabilitySelector
	volumes       volumeCalculator
	coursePayouts coursePayoutCalculator
	normalizers   normalizationFactorCalculator
	scalers       scaledPayoutCalculator
	excesses      excessCalculator
	adjustments   payoutAdjuster
	aggregators   userPayoutAggregator
	netter        marketPositionNetter
	allocator     singleCreditAllocator
}

var defaultMarketShareCalculator = MarketShareCalculator{
	probabilities: wpamProbabilitySelector{},
	volumes:       marketVolumeCalculator{},
	coursePayouts: defaultCoursePayoutCalculator{},
	normalizers:   defaultNormalizationFactorCalculator{},
	scalers:       defaultScaledPayoutCalculator{},
	excesses:      defaultExcessCalculator{volumes: marketVolumeCalculator{}},
	adjustments:   defaultPayoutAdjuster{excesses: defaultExcessCalculator{volumes: marketVolumeCalculator{}}},
	aggregators:   defaultUserPayoutAggregator{},
	netter:        defaultMarketPositionNetter{},
	allocator:     defaultSingleCreditAllocator{},
}

type defaultCoursePayoutCalculator struct{}
type defaultNormalizationFactorCalculator struct{}
type defaultScaledPayoutCalculator struct{}
type defaultExcessCalculator struct {
	volumes volumeCalculator
}
type defaultPayoutAdjuster struct {
	excesses excessCalculator
}
type defaultUserPayoutAggregator struct{}
type defaultMarketPositionNetter struct{}
type defaultSingleCreditAllocator struct{}

// holds betting payout information
type CourseBetPayout struct {
	Payout  float64
	Outcome string
}

type DBPMMarketPosition struct {
	Username       string
	NoSharesOwned  int64
	YesSharesOwned int64
}

// DivideUpMarketPoolSharesDBPM divides the market pool into YES and NO pools based on the resolution probability.
// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
func DivideUpMarketPoolSharesDBPM(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) (int64, int64) {
	return defaultMarketShareCalculator.DivideShares(bets, probabilityChanges)
}

func (c MarketShareCalculator) DivideShares(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) (int64, int64) {
	if len(probabilityChanges) == 0 {
		return 0, 0
	}
	c = c.withDefaults()

	// Get the last probability change, which is the resolution probability
	currentProbability := c.probabilities.Current(probabilityChanges)

	// Get the total share pool as a float for precision
	// Do not include the initial market subsidization in volume until market hits final resolution
	totalSharePool := float64(c.volumes.Volume(bets))

	// Initial condition, shares set to zero
	yesShares := int64(0)
	noShares := int64(0)

	// Check case where there is only one bet
	if c.volumes.Volume(bets) == 1 {
		yesShares, noShares = c.allocator.Allocate(bets)
	} else {
		yesShares, noShares = dividePoolShares(totalSharePool, currentProbability)
	}

	// Return calculated shares
	return yesShares, noShares
}

func (c MarketShareCalculator) withDefaults() MarketShareCalculator {
	if c.probabilities == nil {
		c.probabilities = wpamProbabilitySelector{}
	}
	if c.volumes == nil {
		c.volumes = marketVolumeCalculator{}
	}
	if c.coursePayouts == nil {
		c.coursePayouts = defaultCoursePayoutCalculator{}
	}
	if c.normalizers == nil {
		c.normalizers = defaultNormalizationFactorCalculator{}
	}
	if c.scalers == nil {
		c.scalers = defaultScaledPayoutCalculator{}
	}
	if c.excesses == nil {
		c.excesses = defaultExcessCalculator{volumes: c.volumes}
	}
	if c.adjustments == nil {
		c.adjustments = defaultPayoutAdjuster{excesses: c.excesses}
	}
	if c.aggregators == nil {
		c.aggregators = defaultUserPayoutAggregator{}
	}
	if c.netter == nil {
		c.netter = defaultMarketPositionNetter{}
	}
	if c.allocator == nil {
		c.allocator = defaultSingleCreditAllocator{}
	}
	return c
}

func dividePoolShares(totalSharePool, probability float64) (int64, int64) {
	var totals sideExposure
	for outcome, weight := range dbpmPoolShareWeights {
		allocateExposure(&totals, outcome, int64(math.Round(totalSharePool*weight(probability))))
	}
	return totals.yes, totals.no
}

// CalculateCoursePayoutsDBPM calculates the course payout for each bet in the market,
// separating the payouts for YES and NO outcomes.
// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
func CalculateCoursePayoutsDBPM(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) []CourseBetPayout {
	return defaultMarketShareCalculator.CoursePayouts(bets, probabilityChanges)
}

func (c MarketShareCalculator) CoursePayouts(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) []CourseBetPayout {
	return c.withDefaults().coursePayouts.CoursePayouts(bets, probabilityChanges)
}

func (defaultCoursePayoutCalculator) CoursePayouts(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) []CourseBetPayout {
	if len(probabilityChanges) == 0 {
		return nil
	}

	var coursePayouts []CourseBetPayout

	// Get the current (final) probability for the market
	currentProbability := probabilityChanges[len(probabilityChanges)-1].Probability

	// Iterate over each bet to calculate its course payout
	for i, bet := range bets {
		// Probability at which the bet was placed is the bet index+1
		// The probability index is always the length of the bet index+1 because of the initial probability
		betProbabilityAtTimePlaced := probabilityChanges[i+1].Probability

		coursePaymentForBet := math.Abs(currentProbability-betProbabilityAtTimePlaced) * float64(bet.Amount)

		// Append the calculated payout to the result
		coursePayouts = append(coursePayouts, CourseBetPayout{Payout: coursePaymentForBet, Outcome: bet.Outcome})
	}

	return coursePayouts
}

// F_YES calculates the normalization factor for "YES" by dividing the total stake by the cumulative payout for "YES".
// F_NO calculates the normalization factor for "NO" by dividing the total stake by the cumulative payout for "NO".
// Return absolute values of normalization factors to ensure non-negative values for further calculations.
func CalculateNormalizationFactorsDBPM(yesShares int64, noShares int64, coursePayouts []CourseBetPayout) (float64, float64) {
	return defaultMarketShareCalculator.NormalizationFactors(yesShares, noShares, coursePayouts)
}

func (c MarketShareCalculator) NormalizationFactors(yesShares int64, noShares int64, coursePayouts []CourseBetPayout) (float64, float64) {
	return c.withDefaults().normalizers.NormalizationFactors(yesShares, noShares, coursePayouts)
}

func (defaultNormalizationFactorCalculator) NormalizationFactors(yesShares int64, noShares int64, coursePayouts []CourseBetPayout) (float64, float64) {
	var payoutSums normalizationSums
	for _, payout := range coursePayouts {
		accumulateNormalizationPayout(&payoutSums, payout)
	}

	return normalizationFactor(yesShares, payoutSums.yes), normalizationFactor(noShares, payoutSums.no)
}

// CalculateFinalPayouts calculates the final payouts for each bet, adjusted by normalization factors.
func CalculateScaledPayoutsDBPM(allBetsOnMarket []models.Bet, coursePayouts []CourseBetPayout, yesNormalizationFactor, noNormalizationFactor float64) []int64 {
	return defaultMarketShareCalculator.ScaledPayouts(allBetsOnMarket, coursePayouts, yesNormalizationFactor, noNormalizationFactor)
}

func (c MarketShareCalculator) ScaledPayouts(allBetsOnMarket []models.Bet, coursePayouts []CourseBetPayout, yesNormalizationFactor, noNormalizationFactor float64) []int64 {
	return c.withDefaults().scalers.ScaledPayouts(allBetsOnMarket, coursePayouts, yesNormalizationFactor, noNormalizationFactor)
}

func (defaultScaledPayoutCalculator) ScaledPayouts(allBetsOnMarket []models.Bet, coursePayouts []CourseBetPayout, yesNormalizationFactor, noNormalizationFactor float64) []int64 {
	scaledPayouts := make([]int64, len(allBetsOnMarket))
	factors := normalizationFactors{yes: yesNormalizationFactor, no: noNormalizationFactor}

	for i, payout := range coursePayouts {
		scaledPayouts[i] = int64(math.Round(scalePayout(payout, factors)))
	}

	return scaledPayouts
}

// calculateExcess determines the amount of credits unaccounted for by comparing calculated scaledPayouts to availablePool
func calculateExcess(bets []models.Bet, scaledPayouts []int64) int64 {
	return defaultMarketShareCalculator.Excess(bets, scaledPayouts)
}

func (c MarketShareCalculator) Excess(bets []models.Bet, scaledPayouts []int64) int64 {
	return c.withDefaults().excesses.Excess(bets, scaledPayouts)
}

func (c defaultExcessCalculator) Excess(bets []models.Bet, scaledPayouts []int64) int64 {
	if c.volumes == nil {
		c.volumes = marketVolumeCalculator{}
	}
	var sumScaledPayouts int64
	for _, payout := range scaledPayouts {
		sumScaledPayouts += payout
	}
	availablePool := c.volumes.Volume(bets)
	return sumScaledPayouts - availablePool
}

// Adjust scaled payouts if excess is greater than 0
// This  should not be possible given how the preceeding pipeline works, but we adjust for it anyway.
func adjustForPositiveExcess(scaledPayouts []int64, excess int64) []int64 {
	return defaultMarketShareCalculator.AdjustPositiveExcess(scaledPayouts, excess)
}

func (c MarketShareCalculator) AdjustPositiveExcess(scaledPayouts []int64, excess int64) []int64 {
	return c.withDefaults().adjustments.AdjustPositive(scaledPayouts, excess)
}

func (defaultPayoutAdjuster) AdjustPositive(scaledPayouts []int64, excess int64) []int64 {
	// No adjustment needed if no payouts or excess is non-positive
	if excess <= 0 || len(scaledPayouts) == 0 {
		return scaledPayouts
	}

	numBets := int64(len(scaledPayouts)) // Total number of bets
	absoluteExcess := excess             // No need to negate since it's already positive

	// Calculate the base reduction for each bet and the leftover remainder
	baseReduction := absoluteExcess / numBets
	totalReduction := baseReduction * numBets
	remainderReduction := absoluteExcess - totalReduction

	// Apply the base reduction to all payouts
	for betIndex := range scaledPayouts {
		scaledPayouts[betIndex] -= baseReduction
	}

	// Apply the remainder reduction to the newest bets
	for betIndex := int64(len(scaledPayouts)) - 1; remainderReduction > 0; betIndex-- {
		scaledPayouts[betIndex] -= 1
		remainderReduction--
	}

	return scaledPayouts
}

func adjustForNegativeExcess(scaledPayouts []int64, excess int64) []int64 {
	return defaultMarketShareCalculator.AdjustNegativeExcess(scaledPayouts, excess)
}

func (c MarketShareCalculator) AdjustNegativeExcess(scaledPayouts []int64, excess int64) []int64 {
	return c.withDefaults().adjustments.AdjustNegative(scaledPayouts, excess)
}

func (defaultPayoutAdjuster) AdjustNegative(scaledPayouts []int64, excess int64) []int64 {
	// No adjustment needed if no payouts or excess is non-negative
	if excess >= 0 || len(scaledPayouts) == 0 {
		return scaledPayouts
	}

	numBets := int64(len(scaledPayouts)) // Total number of bets
	absoluteExcess := -excess            // Convert excess to positive for allocation

	// Calculate the base addition for each bet and the leftover remainder
	// int64 will apply floor division
	baseAddition := int64(absoluteExcess / numBets)
	totalAddition := baseAddition * numBets
	remainderAddition := absoluteExcess - totalAddition

	// Apply the base addition to all payouts
	for betIndex := range scaledPayouts {
		scaledPayouts[betIndex] += baseAddition
	}

	// Apply the remainder addition to the earliest bets
	for betIndex := int64(0); betIndex < remainderAddition; betIndex++ {
		scaledPayouts[betIndex] += 1
	}

	return scaledPayouts
}

// AdjustPayouts reconciles the additional or lacking funds from the betting pool by adjusting the payouts to past bets
func AdjustPayouts(bets []models.Bet, scaledPayouts []int64) []int64 {
	return defaultMarketShareCalculator.AdjustPayouts(bets, scaledPayouts)
}

func (c MarketShareCalculator) AdjustPayouts(bets []models.Bet, scaledPayouts []int64) []int64 {
	return c.withDefaults().adjustments.Adjust(bets, scaledPayouts)
}

func (a defaultPayoutAdjuster) Adjust(bets []models.Bet, scaledPayouts []int64) []int64 {
	if a.excesses == nil {
		a.excesses = defaultExcessCalculator{volumes: marketVolumeCalculator{}}
	}
	excess := a.excesses.Excess(bets, scaledPayouts)
	if excess > 0 {
		scaledPayouts = a.AdjustPositive(scaledPayouts, excess)
	} else if excess < 0 {
		scaledPayouts = a.AdjustNegative(scaledPayouts, excess)
	}

	return scaledPayouts
}

// AggregateUserPayouts aggregates YES and NO payouts for each user.
func AggregateUserPayoutsDBPM(bets []models.Bet, finalPayouts []int64) []DBPMMarketPosition {
	return defaultMarketShareCalculator.AggregateUserPayouts(bets, finalPayouts)
}

func (c MarketShareCalculator) AggregateUserPayouts(bets []models.Bet, finalPayouts []int64) []DBPMMarketPosition {
	return c.withDefaults().aggregators.Aggregate(bets, finalPayouts)
}

func (defaultUserPayoutAggregator) Aggregate(bets []models.Bet, finalPayouts []int64) []DBPMMarketPosition {
	userPayouts := make(map[string]*DBPMMarketPosition)

	for i, bet := range bets {
		payout := finalPayouts[i]

		// Initialize the user's market position if it doesn't exist
		if _, exists := userPayouts[bet.Username]; !exists {
			userPayouts[bet.Username] = &DBPMMarketPosition{Username: bet.Username}
		}

		aggregatePositionByOutcome(userPayouts[bet.Username], bet.Outcome, payout)
	}

	// Convert map to slice for output
	var positions []DBPMMarketPosition
	for _, pos := range userPayouts {
		// Check and adjust negative shares to 0
		if pos.YesSharesOwned < 0 {
			pos.YesSharesOwned = 0
		}
		if pos.NoSharesOwned < 0 {
			pos.NoSharesOwned = 0
		}
		positions = append(positions, *pos)
	}

	return positions
}

// Function to normalize market positions such that for each user,
// only one of YesSharesOwned or NoSharesOwned is greater than 0,
// with the other being 0, and the value is the net difference.
func NetAggregateMarketPositions(positions []DBPMMarketPosition) []DBPMMarketPosition {
	return defaultMarketShareCalculator.NetPositions(positions)
}

func (c MarketShareCalculator) NetPositions(positions []DBPMMarketPosition) []DBPMMarketPosition {
	return c.withDefaults().netter.Net(positions)
}

func (defaultMarketPositionNetter) Net(positions []DBPMMarketPosition) []DBPMMarketPosition {
	var normalizedPositions []DBPMMarketPosition

	for _, position := range positions {
		var normalizedPosition DBPMMarketPosition
		normalizedPosition.Username = position.Username

		if position.YesSharesOwned > position.NoSharesOwned {
			normalizedPosition.YesSharesOwned = position.YesSharesOwned - position.NoSharesOwned
			normalizedPosition.NoSharesOwned = 0
		} else {
			normalizedPosition.NoSharesOwned = position.NoSharesOwned - position.YesSharesOwned
			normalizedPosition.YesSharesOwned = 0
		}

		normalizedPositions = append(normalizedPositions, normalizedPosition)
	}

	return normalizedPositions
}

// SingleCreditYesNoAllocator assigns the remaining credit/share to YES or NO, based on net position.
func singleCreditYesNoAllocator(bets []models.Bet) (yesShares int64, noShares int64) {
	return defaultMarketShareCalculator.AllocateSingleCredit(bets)
}

func (c MarketShareCalculator) AllocateSingleCredit(bets []models.Bet) (yesShares int64, noShares int64) {
	return c.withDefaults().allocator.Allocate(bets)
}

func (defaultSingleCreditAllocator) Allocate(bets []models.Bet) (yesShares int64, noShares int64) {
	var exposure sideExposure
	for _, bet := range bets {
		allocateExposure(&exposure, bet.Outcome, bet.Amount)
	}
	if exposure.yes > exposure.no {
		return 1, 0
	} else if exposure.no > exposure.yes {
		return 0, 1
	}
	// If equal or ambiguous, assign to neither (fallback)
	return 0, 0
}

func accumulateNormalizationPayout(sums *normalizationSums, payout CourseBetPayout) {
	if accumulator, ok := dbpmNormalizationAccumulators[payout.Outcome]; ok {
		accumulator(sums, payout.Payout)
	}
}

func normalizationFactor(shares int64, payoutSum float64) float64 {
	if payoutSum <= 0 {
		return 0
	}
	return math.Abs(float64(shares) / payoutSum)
}

func scalePayout(payout CourseBetPayout, factors normalizationFactors) float64 {
	selector, ok := dbpmNormalizationSelectors[payout.Outcome]
	if !ok {
		return 0
	}
	return payout.Payout * selector(factors)
}

func aggregatePositionByOutcome(position *DBPMMarketPosition, outcome string, payout int64) {
	if aggregate, ok := dbpmPayoutAggregators[outcome]; ok {
		aggregate(position, payout)
	}
}

func allocateExposure(exposure *sideExposure, outcome string, amount int64) {
	if accumulate, ok := dbpmExposureAccumulators[outcome]; ok {
		accumulate(exposure, amount)
	}
}
