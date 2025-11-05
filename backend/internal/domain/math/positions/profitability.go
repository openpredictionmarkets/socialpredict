package positionsmath

import (
	"socialpredict/models"
	"sort"
	"time"
)

// UserProfitability represents a user's profitability data for a specific market.
type UserProfitability struct {
	Username       string    `json:"username"`
	CurrentValue   int64     `json:"currentValue"`
	TotalSpent     int64     `json:"totalSpent"`
	Profit         int64     `json:"profit"`
	Position       string    `json:"position"` // "YES", "NO", "NEUTRAL"
	YesSharesOwned int64     `json:"yesSharesOwned"`
	NoSharesOwned  int64     `json:"noSharesOwned"`
	EarliestBet    time.Time `json:"earliestBet"`
	Rank           int       `json:"rank"`
}

// CalculateMarketLeaderboard ranks users in a market by profitability.
func CalculateMarketLeaderboard(snapshot MarketSnapshot, bets []models.Bet) ([]UserProfitability, error) {
	positions, err := CalculateMarketPositions_WPAM_DBPM(snapshot, bets)
	if err != nil {
		return nil, err
	}

	if len(bets) == 0 {
		return []UserProfitability{}, nil
	}

	var leaderboard []UserProfitability

	for _, position := range positions {
		if position.YesSharesOwned == 0 && position.NoSharesOwned == 0 {
			continue
		}

		totalSpent := CalculateUserSpend(bets, position.Username)
		profit := position.Value - totalSpent
		positionType := DeterminePositionType(position.YesSharesOwned, position.NoSharesOwned)
		earliestBet := GetEarliestBetTime(bets, position.Username)

		leaderboard = append(leaderboard, UserProfitability{
			Username:       position.Username,
			CurrentValue:   position.Value,
			TotalSpent:     totalSpent,
			Profit:         profit,
			Position:       positionType,
			YesSharesOwned: position.YesSharesOwned,
			NoSharesOwned:  position.NoSharesOwned,
			EarliestBet:    earliestBet,
		})
	}

	sort.Slice(leaderboard, func(i, j int) bool {
		if leaderboard[i].Profit == leaderboard[j].Profit {
			return leaderboard[i].EarliestBet.Before(leaderboard[j].EarliestBet)
		}
		return leaderboard[i].Profit > leaderboard[j].Profit
	})

	for i := range leaderboard {
		leaderboard[i].Rank = i + 1
	}

	return leaderboard, nil
}

// CalculateUserSpend sums a user's total spend (positive buys, negative sells).
func CalculateUserSpend(bets []models.Bet, username string) int64 {
	var total int64
	for _, bet := range bets {
		if bet.Username == username {
			total += bet.Amount
		}
	}
	return total
}

// GetEarliestBetTime returns the earliest bet timestamp for the user.
func GetEarliestBetTime(bets []models.Bet, username string) time.Time {
	var earliest time.Time
	first := true

	for _, bet := range bets {
		if bet.Username != username {
			continue
		}
		if first || bet.PlacedAt.Before(earliest) {
			earliest = bet.PlacedAt
			first = false
		}
	}

	return earliest
}

// DeterminePositionType identifies whether the user holds YES, NO, or both.
func DeterminePositionType(yesShares, noShares int64) string {
	switch {
	case yesShares > 0 && noShares == 0:
		return "YES"
	case noShares > 0 && yesShares == 0:
		return "NO"
	case yesShares > 0 && noShares > 0:
		return "NEUTRAL"
	default:
		return "NONE"
	}
}
