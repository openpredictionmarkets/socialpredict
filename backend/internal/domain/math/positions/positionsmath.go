package positionsmath

import (
	"sort"
	"time"

	marketmath "socialpredict/internal/domain/math/market"
	"socialpredict/internal/domain/math/outcomes/dbpm"
	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models"
)

// holds the number of YES and NO shares owned by all users in a market
type MarketPosition struct {
	Username         string `json:"username"`
	MarketID         uint   `json:"marketId"`
	NoSharesOwned    int64  `json:"noSharesOwned"`
	YesSharesOwned   int64  `json:"yesSharesOwned"`
	Value            int64  `json:"value"`
	TotalSpent       int64  `json:"totalSpent"`       // Total amount user spent in this market
	TotalSpentInPlay int64  `json:"totalSpentInPlay"` // Amount spent in unresolved markets only
	IsResolved       bool   `json:"isResolved"`       // From market.IsResolved
	ResolutionResult string `json:"resolutionResult"` // From market.ResolutionResult
}

// UserMarketPosition holds the number of YES and NO shares owned by a user in a market.
type UserMarketPosition struct {
	NoSharesOwned    int64  `json:"noSharesOwned"`
	YesSharesOwned   int64  `json:"yesSharesOwned"`
	Value            int64  `json:"value"`
	TotalSpent       int64  `json:"totalSpent"`
	TotalSpentInPlay int64  `json:"totalSpentInPlay"`
	IsResolved       bool   `json:"isResolved"`
	ResolutionResult string `json:"resolutionResult"`
}

// MarketSnapshot captures the minimal market context needed for position calculations.
type MarketSnapshot struct {
	ID               int64
	CreatedAt        time.Time
	IsResolved       bool
	ResolutionResult string
}

// ProbabilityProvider abstracts probability timeline calculations.
type ProbabilityProvider interface {
	Calculate(createdAt time.Time, bets []models.Bet) []wpam.ProbabilityChange
	Current(changes []wpam.ProbabilityChange) float64
}

// PayoutModel defines how to compute market positions and payouts.
type PayoutModel interface {
	DivideShares(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) (int64, int64)
	CoursePayouts(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) []dbpm.CourseBetPayout
	NormalizationFactors(yesShares, noShares int64, coursePayouts []dbpm.CourseBetPayout) (float64, float64)
	ScaledPayouts(bets []models.Bet, coursePayouts []dbpm.CourseBetPayout, yesFactor, noFactor float64) []int64
	AdjustFinalPayouts(bets []models.Bet, scaledPayouts []int64) []int64
	AggregateUserPayouts(bets []models.Bet, finalPayouts []int64) []dbpm.DBPMMarketPosition
	NetAggregateMarketPositions(positions []dbpm.DBPMMarketPosition) []dbpm.DBPMMarketPosition
}

// PositionCalculator encapsulates dependencies used to compute positions.
type PositionCalculator struct {
	probabilities ProbabilityProvider
	payouts       PayoutModel
}

// NewPositionCalculator creates a calculator with default WPAM/DBPM components.
func NewPositionCalculator() PositionCalculator {
	return PositionCalculator{
		probabilities: defaultProbabilityProvider{},
		payouts:       defaultPayoutModel{},
	}
}

// CalculateMarketPositions_WPAM_DBPM summarizes positions for a given market using WPAM/DBPM math.
func CalculateMarketPositions_WPAM_DBPM(snapshot MarketSnapshot, bets []models.Bet) ([]MarketPosition, error) {
	return NewPositionCalculator().CalculateMarketPositions(snapshot, bets)
}

// CalculateMarketPositions runs the position calculation using the calculator's injected strategies.
func (c PositionCalculator) CalculateMarketPositions(snapshot MarketSnapshot, bets []models.Bet) ([]MarketPosition, error) {
	marketIDUint := uint(snapshot.ID)

	sortedBets := sortBetsChronologically(bets)

	allProbabilityChangesOnMarket := c.probabilities.Calculate(snapshot.CreatedAt, sortedBets)
	netPositions := c.calculateNetPositions(sortedBets, allProbabilityChangesOnMarket)

	userPositionMap := mapUserPositions(netPositions)
	currentProbability := c.probabilities.Current(allProbabilityChangesOnMarket)
	totalVolume := marketmath.GetMarketVolume(sortedBets)
	earliestBets := computeEarliestBets(sortedBets)

	valuations, err := CalculateRoundedUserValuationsFromUserMarketPositions(
		userPositionMap,
		currentProbability,
		totalVolume,
		snapshot.IsResolved,
		snapshot.ResolutionResult,
		earliestBets,
	)
	if err != nil {
		return nil, err
	}

	userBetTotals := aggregateUserBetTotals(sortedBets, snapshot.IsResolved)
	displayPositions := assembleDisplayPositions(netPositions, valuations, userBetTotals, snapshot, marketIDUint)

	return displayPositions, nil
}

func computeEarliestBets(bets []models.Bet) map[string]time.Time {
	earliest := make(map[string]time.Time)
	for _, bet := range bets {
		if existing, ok := earliest[bet.Username]; !ok || bet.PlacedAt.Before(existing) {
			earliest[bet.Username] = bet.PlacedAt
		}
	}
	return earliest
}

func sortBetsChronologically(bets []models.Bet) []models.Bet {
	sortedBets := make([]models.Bet, len(bets))
	copy(sortedBets, bets)
	sort.Slice(sortedBets, func(i, j int) bool {
		return sortedBets[i].PlacedAt.Before(sortedBets[j].PlacedAt)
	})
	return sortedBets
}

func calculateNetPositions(sortedBets []models.Bet, probabilityChanges []wpam.ProbabilityChange) []dbpm.DBPMMarketPosition {
	return NewPositionCalculator().calculateNetPositions(sortedBets, probabilityChanges)
}

func (c PositionCalculator) calculateNetPositions(sortedBets []models.Bet, probabilityChanges []wpam.ProbabilityChange) []dbpm.DBPMMarketPosition {
	S_YES, S_NO := c.payouts.DivideShares(sortedBets, probabilityChanges)
	coursePayouts := c.payouts.CoursePayouts(sortedBets, probabilityChanges)
	F_YES, F_NO := c.payouts.NormalizationFactors(S_YES, S_NO, coursePayouts)
	scaledPayouts := c.payouts.ScaledPayouts(sortedBets, coursePayouts, F_YES, F_NO)
	finalPayouts := c.payouts.AdjustFinalPayouts(sortedBets, scaledPayouts)
	aggreatedPositions := c.payouts.AggregateUserPayouts(sortedBets, finalPayouts)
	return c.payouts.NetAggregateMarketPositions(aggreatedPositions)
}

func mapUserPositions(netPositions []dbpm.DBPMMarketPosition) map[string]UserMarketPosition {
	userPositionMap := make(map[string]UserMarketPosition)
	for _, p := range netPositions {
		userPositionMap[p.Username] = UserMarketPosition{
			YesSharesOwned: p.YesSharesOwned,
			NoSharesOwned:  p.NoSharesOwned,
		}
	}
	return userPositionMap
}

func aggregateUserBetTotals(sortedBets []models.Bet, isResolved bool) map[string]struct {
	TotalSpent       int64
	TotalSpentInPlay int64
} {
	userBetTotals := make(map[string]struct {
		TotalSpent       int64
		TotalSpentInPlay int64
	})

	for _, bet := range sortedBets {
		totals := userBetTotals[bet.Username]
		totals.TotalSpent += bet.Amount
		if !isResolved {
			totals.TotalSpentInPlay += bet.Amount
		}
		userBetTotals[bet.Username] = totals
	}
	return userBetTotals
}

func assembleDisplayPositions(
	netPositions []dbpm.DBPMMarketPosition,
	valuations map[string]UserValuationResult,
	userBetTotals map[string]struct {
		TotalSpent       int64
		TotalSpentInPlay int64
	},
	snapshot MarketSnapshot,
	marketIDUint uint,
) []MarketPosition {
	var (
		displayPositions []MarketPosition
		seenUsers        = make(map[string]bool)
	)

	for _, p := range netPositions {
		val := valuations[p.Username]
		betTotals := userBetTotals[p.Username]
		displayPositions = append(displayPositions, MarketPosition{
			Username:         p.Username,
			MarketID:         marketIDUint,
			YesSharesOwned:   p.YesSharesOwned,
			NoSharesOwned:    p.NoSharesOwned,
			Value:            val.RoundedValue,
			TotalSpent:       betTotals.TotalSpent,
			TotalSpentInPlay: betTotals.TotalSpentInPlay,
			IsResolved:       snapshot.IsResolved,
			ResolutionResult: snapshot.ResolutionResult,
		})
		seenUsers[p.Username] = true
	}

	for username, totals := range userBetTotals {
		if seenUsers[username] {
			continue
		}

		displayPositions = append(displayPositions, MarketPosition{
			Username:         username,
			MarketID:         marketIDUint,
			YesSharesOwned:   0,
			NoSharesOwned:    0,
			Value:            valuations[username].RoundedValue,
			TotalSpent:       totals.TotalSpent,
			TotalSpentInPlay: totals.TotalSpentInPlay,
			IsResolved:       snapshot.IsResolved,
			ResolutionResult: snapshot.ResolutionResult,
		})
	}

	return displayPositions
}

// CalculateMarketPositionForUser_WPAM_DBPM fetches and summarizes the position for a given user in a specific market.
func CalculateMarketPositionForUser_WPAM_DBPM(snapshot MarketSnapshot, bets []models.Bet, username string) (UserMarketPosition, error) {
	marketPositions, err := CalculateMarketPositions_WPAM_DBPM(snapshot, bets)
	if err != nil {
		return UserMarketPosition{}, err
	}

	for _, position := range marketPositions {
		if position.Username == username {
			return UserMarketPosition{
				NoSharesOwned:    position.NoSharesOwned,
				YesSharesOwned:   position.YesSharesOwned,
				Value:            position.Value,
				TotalSpent:       position.TotalSpent,
				TotalSpentInPlay: position.TotalSpentInPlay,
				IsResolved:       position.IsResolved,
				ResolutionResult: position.ResolutionResult,
			}, nil
		}
	}

	return UserMarketPosition{}, nil
}

// CalculateAllUserMarketPositions_WPAM_DBPM is deprecated. Prefer computing positions via
// CalculateMarketPositions_WPAM_DBPM after fetching market snapshots and bet histories from a repository.
