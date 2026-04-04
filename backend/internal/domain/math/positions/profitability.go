package positionsmath

import (
	"socialpredict/models"
	"sort"
	"time"
)

const (
	positionTypeYes     = "YES"
	positionTypeNo      = "NO"
	positionTypeNeutral = "NEUTRAL"
	positionTypeNone    = "NONE"
)

type positionTypeRule struct {
	position string
	matches  func(yesShares, noShares int64) bool
}

var defaultPositionTypeRules = []positionTypeRule{
	{
		position: positionTypeYes,
		matches: func(yesShares, noShares int64) bool {
			return yesShares > 0 && noShares == 0
		},
	},
	{
		position: positionTypeNo,
		matches: func(yesShares, noShares int64) bool {
			return noShares > 0 && yesShares == 0
		},
	},
	{
		position: positionTypeNeutral,
		matches: func(yesShares, noShares int64) bool {
			return yesShares > 0 && noShares > 0
		},
	},
}

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
	if len(bets) == 0 {
		return []UserProfitability{}, nil
	}

	positions, err := CalculateMarketPositions_WPAM_DBPM(snapshot, bets)
	if err != nil {
		return nil, err
	}

	leaderboard := buildLeaderboard(positions, bets)
	sortLeaderboard(leaderboard)
	assignLeaderboardRanks(leaderboard)
	return leaderboard, nil
}

func buildLeaderboard(positions []MarketPosition, bets []models.Bet) []UserProfitability {
	leaderboard := make([]UserProfitability, 0, len(positions))
	for _, position := range positions {
		if hasNoOpenPosition(position) {
			continue
		}
		leaderboard = append(leaderboard, newUserProfitability(position, bets))
	}
	return leaderboard
}

func hasNoOpenPosition(position MarketPosition) bool {
	return position.YesSharesOwned == 0 && position.NoSharesOwned == 0
}

func newUserProfitability(position MarketPosition, bets []models.Bet) UserProfitability {
	totalSpent := CalculateUserSpend(bets, position.Username)
	return UserProfitability{
		Username:       position.Username,
		CurrentValue:   position.Value,
		TotalSpent:     totalSpent,
		Profit:         position.Value - totalSpent,
		Position:       DeterminePositionType(position.YesSharesOwned, position.NoSharesOwned),
		YesSharesOwned: position.YesSharesOwned,
		NoSharesOwned:  position.NoSharesOwned,
		EarliestBet:    GetEarliestBetTime(bets, position.Username),
	}
}

func sortLeaderboard(leaderboard []UserProfitability) {
	sort.Slice(leaderboard, func(i, j int) bool {
		if leaderboard[i].Profit == leaderboard[j].Profit {
			return leaderboard[i].EarliestBet.Before(leaderboard[j].EarliestBet)
		}
		return leaderboard[i].Profit > leaderboard[j].Profit
	})
}

func assignLeaderboardRanks(leaderboard []UserProfitability) {
	for i := range leaderboard {
		leaderboard[i].Rank = i + 1
	}
}

// CalculateUserSpend sums a user's total spend (positive buys, negative sells).
func CalculateUserSpend(bets []models.Bet, username string) int64 {
	var total int64
	forEachUserBet(bets, username, func(bet models.Bet) {
		total += bet.Amount
	})
	return total
}

// GetEarliestBetTime returns the earliest bet timestamp for the user.
func GetEarliestBetTime(bets []models.Bet, username string) time.Time {
	var earliest time.Time
	first := true

	forEachUserBet(bets, username, func(bet models.Bet) {
		if first || bet.PlacedAt.Before(earliest) {
			earliest = bet.PlacedAt
			first = false
		}
	})

	return earliest
}

// DeterminePositionType identifies whether the user holds YES, NO, or both.
func DeterminePositionType(yesShares, noShares int64) string {
	for _, rule := range defaultPositionTypeRules {
		if rule.matches(yesShares, noShares) {
			return rule.position
		}
	}
	return positionTypeNone
}

func forEachUserBet(bets []models.Bet, username string, visit func(models.Bet)) {
	for _, bet := range bets {
		if bet.Username != username {
			continue
		}
		visit(bet)
	}
}
