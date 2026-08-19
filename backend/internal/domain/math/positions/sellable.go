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
	unlockedShares := deriveRemainingUnlockedShares(payouts, username, outcome)
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

type unlockedLot struct {
	remainingShares int64
}

func deriveRemainingUnlockedShares(payouts []BetPayout, username string, outcome string) int64 {
	lots := make([]unlockedLot, 0)
	unlockedEnd := 0
	consumeHead := 0

	for _, payout := range payouts {
		bet := payout.Bet
		if bet.Amount > 0 && bet.Username == username && bet.Outcome == outcome && payout.Payout > 0 {
			lots = append(lots, unlockedLot{remainingShares: payout.Payout})
		}
		if bet.Amount > 0 && bet.Username != username {
			unlockedEnd = len(lots)
		}
		if bet.Amount < 0 && bet.Username == username && bet.Outcome == outcome {
			consumeHead = consumeUnlockedLots(lots, consumeHead, unlockedEnd, -bet.Amount)
		}
	}

	return sumRemainingUnlockedLots(lots, consumeHead, unlockedEnd)
}

func consumeUnlockedLots(lots []unlockedLot, consumeHead int, unlockedEnd int, shares int64) int {
	if consumeHead < 0 {
		consumeHead = 0
	}
	if unlockedEnd > len(lots) {
		unlockedEnd = len(lots)
	}
	for shares > 0 && consumeHead < unlockedEnd {
		consumed := lots[consumeHead].remainingShares
		if consumed > shares {
			consumed = shares
		}
		lots[consumeHead].remainingShares -= consumed
		shares -= consumed
		if lots[consumeHead].remainingShares == 0 {
			consumeHead++
		}
	}
	return consumeHead
}

func sumRemainingUnlockedLots(lots []unlockedLot, consumeHead int, unlockedEnd int) int64 {
	if consumeHead < 0 {
		consumeHead = 0
	}
	if unlockedEnd > len(lots) {
		unlockedEnd = len(lots)
	}
	total := int64(0)
	for i := consumeHead; i < unlockedEnd; i++ {
		total += lots[i].remainingShares
	}
	return total
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
