package positionsmath

import (
	"socialpredict/internal/domain/boundary"
	"socialpredict/internal/domain/math/outcomes/dbpm"
)

// BetPayout keeps a DBPM final payout attached to the source bet row.
type BetPayout struct {
	Bet    boundary.Bet
	Payout int64
}

// CalculateBetPayouts_WPAM_DBPM returns DBPM final payouts before user aggregation.
func CalculateBetPayouts_WPAM_DBPM(snapshot MarketSnapshot, bets []boundary.Bet) []BetPayout {
	calc := NewPositionCalculator()
	sortedBets := calc.sorter.Sort(bets)
	if len(sortedBets) == 0 {
		return nil
	}

	probabilityChanges := calc.probabilities.Calculate(snapshot.CreatedAt, sortedBets)
	yesShares, noShares := dbpm.DivideUpMarketPoolSharesDBPM(sortedBets, probabilityChanges)
	coursePayouts := dbpm.CalculateCoursePayoutsDBPM(sortedBets, probabilityChanges)
	yesFactor, noFactor := dbpm.CalculateNormalizationFactorsDBPM(yesShares, noShares, coursePayouts)
	scaledPayouts := dbpm.CalculateScaledPayoutsDBPM(sortedBets, coursePayouts, yesFactor, noFactor)
	finalPayouts := dbpm.AdjustPayouts(sortedBets, scaledPayouts)

	out := make([]BetPayout, 0, len(sortedBets))
	for i, bet := range sortedBets {
		payout := int64(0)
		if i < len(finalPayouts) {
			payout = finalPayouts[i]
		}
		out = append(out, BetPayout{Bet: bet, Payout: payout})
	}
	return out
}

// CalculateUnlockedSellablePosition_WPAM_DBPM returns the portion of a user's
// current position backed by prior buy rows that have a later buy from another user.
func CalculateUnlockedSellablePosition_WPAM_DBPM(snapshot MarketSnapshot, bets []boundary.Bet, username string, outcome string) (UserMarketPosition, error) {
	current, err := CalculateMarketPositionForUser_WPAM_DBPM(snapshot, bets, username)
	if err != nil {
		return UserMarketPosition{}, err
	}

	currentShares := sharesForPositionOutcome(current, outcome)
	if currentShares <= 0 || current.Value <= 0 {
		return UserMarketPosition{
			TotalSpent:       current.TotalSpent,
			TotalSpentInPlay: current.TotalSpentInPlay,
			IsResolved:       current.IsResolved,
			ResolutionResult: current.ResolutionResult,
		}, nil
	}

	payouts := CalculateBetPayouts_WPAM_DBPM(snapshot, bets)
	unlockedShares := int64(0)
	for i, payout := range payouts {
		if isUnlockedBuy(payouts, i, username, outcome) && payout.Payout > 0 {
			unlockedShares += payout.Payout
		}
	}
	if unlockedShares > currentShares {
		unlockedShares = currentShares
	}
	if unlockedShares <= 0 {
		return UserMarketPosition{
			TotalSpent:       current.TotalSpent,
			TotalSpentInPlay: current.TotalSpentInPlay,
			IsResolved:       current.IsResolved,
			ResolutionResult: current.ResolutionResult,
		}, nil
	}

	valuePerShare := current.Value / currentShares
	sellableValue := unlockedShares * valuePerShare
	if sellableValue > current.Value {
		sellableValue = current.Value
	}

	position := UserMarketPosition{
		Value:            sellableValue,
		TotalSpent:       current.TotalSpent,
		TotalSpentInPlay: current.TotalSpentInPlay,
		IsResolved:       current.IsResolved,
		ResolutionResult: current.ResolutionResult,
	}
	switch outcome {
	case positionTypeYes:
		position.YesSharesOwned = unlockedShares
	case positionTypeNo:
		position.NoSharesOwned = unlockedShares
	}
	return position, nil
}

func isUnlockedBuy(payouts []BetPayout, index int, username string, outcome string) bool {
	if index < 0 || index >= len(payouts) {
		return false
	}
	bet := payouts[index].Bet
	if bet.Username != username || bet.Outcome != outcome || bet.Amount <= 0 {
		return false
	}

	for i := index + 1; i < len(payouts); i++ {
		later := payouts[i].Bet
		if later.Amount > 0 && later.Username != username {
			return true
		}
	}
	return false
}

func sharesForPositionOutcome(position UserMarketPosition, outcome string) int64 {
	switch outcome {
	case positionTypeYes:
		return position.YesSharesOwned
	case positionTypeNo:
		return position.NoSharesOwned
	default:
		return 0
	}
}
