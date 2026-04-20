package positionsmath

import (
	"sort"
	"time"

	"socialpredict/internal/domain/boundary"
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

type MarketPositionSource interface {
	CalculateMarketPositions(snapshot MarketSnapshot, bets []boundary.Bet) ([]MarketPosition, error)
}

type UserSpendCalculator interface {
	Spend(bets []boundary.Bet, username string) int64
}

type EarliestBetTimeFinder interface {
	EarliestBetTime(bets []boundary.Bet, username string) time.Time
}

type PositionTypeResolver interface {
	Resolve(yesShares, noShares int64) string
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

type LeaderboardCalculator struct {
	positions     MarketPositionSource
	spend         UserSpendCalculator
	earliest      EarliestBetTimeFinder
	positionTypes PositionTypeResolver
}

type defaultSpendCalculator struct{}
type defaultEarliestBetTimeFinder struct{}
type defaultPositionTypeResolver struct{}

var defaultLeaderboardCalculator = LeaderboardCalculator{
	positions:     NewPositionCalculator(),
	spend:         defaultSpendCalculator{},
	earliest:      defaultEarliestBetTimeFinder{},
	positionTypes: defaultPositionTypeResolver{},
}

// CalculateMarketLeaderboard ranks users in a market by profitability.
func CalculateMarketLeaderboard(snapshot MarketSnapshot, bets []boundary.Bet) ([]UserProfitability, error) {
	return defaultLeaderboardCalculator.Calculate(snapshot, bets)
}

func (c LeaderboardCalculator) Calculate(snapshot MarketSnapshot, bets []boundary.Bet) ([]UserProfitability, error) {
	if len(bets) == 0 {
		return []UserProfitability{}, nil
	}
	c = c.withDefaults()

	positions, err := c.positions.CalculateMarketPositions(snapshot, bets)
	if err != nil {
		return nil, err
	}

	leaderboard := buildLeaderboard(positions, bets, c.spend, c.earliest, c.positionTypes)
	sortLeaderboard(leaderboard)
	assignLeaderboardRanks(leaderboard)
	return leaderboard, nil
}

func (c LeaderboardCalculator) withDefaults() LeaderboardCalculator {
	if c.positions == nil {
		c.positions = NewPositionCalculator()
	}
	if c.spend == nil {
		c.spend = defaultSpendCalculator{}
	}
	if c.earliest == nil {
		c.earliest = defaultEarliestBetTimeFinder{}
	}
	if c.positionTypes == nil {
		c.positionTypes = defaultPositionTypeResolver{}
	}
	return c
}

func buildLeaderboard(
	positions []MarketPosition,
	bets []boundary.Bet,
	spend UserSpendCalculator,
	earliest EarliestBetTimeFinder,
	positionTypes PositionTypeResolver,
) []UserProfitability {
	leaderboard := make([]UserProfitability, 0, len(positions))
	for _, position := range positions {
		if hasNoOpenPosition(position) {
			continue
		}
		leaderboard = append(leaderboard, newUserProfitability(position, bets, spend, earliest, positionTypes))
	}
	return leaderboard
}

func hasNoOpenPosition(position MarketPosition) bool {
	return position.YesSharesOwned == 0 && position.NoSharesOwned == 0
}

func newUserProfitability(
	position MarketPosition,
	bets []boundary.Bet,
	spend UserSpendCalculator,
	earliest EarliestBetTimeFinder,
	positionTypes PositionTypeResolver,
) UserProfitability {
	totalSpent := spend.Spend(bets, position.Username)
	return UserProfitability{
		Username:       position.Username,
		CurrentValue:   position.Value,
		TotalSpent:     totalSpent,
		Profit:         position.Value - totalSpent,
		Position:       positionTypes.Resolve(position.YesSharesOwned, position.NoSharesOwned),
		YesSharesOwned: position.YesSharesOwned,
		NoSharesOwned:  position.NoSharesOwned,
		EarliestBet:    earliest.EarliestBetTime(bets, position.Username),
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
func CalculateUserSpend(bets []boundary.Bet, username string) int64 {
	return defaultSpendCalculator{}.Spend(bets, username)
}

func (defaultSpendCalculator) Spend(bets []boundary.Bet, username string) int64 {
	var total int64
	forEachUserBet(bets, username, func(bet boundary.Bet) {
		total += bet.Amount
	})
	return total
}

// GetEarliestBetTime returns the earliest bet timestamp for the user.
func GetEarliestBetTime(bets []boundary.Bet, username string) time.Time {
	return defaultEarliestBetTimeFinder{}.EarliestBetTime(bets, username)
}

func (defaultEarliestBetTimeFinder) EarliestBetTime(bets []boundary.Bet, username string) time.Time {
	var earliest time.Time
	first := true

	forEachUserBet(bets, username, func(bet boundary.Bet) {
		if first || bet.PlacedAt.Before(earliest) {
			earliest = bet.PlacedAt
			first = false
		}
	})

	return earliest
}

// DeterminePositionType identifies whether the user holds YES, NO, or both.
func DeterminePositionType(yesShares, noShares int64) string {
	return defaultPositionTypeResolver{}.Resolve(yesShares, noShares)
}

func (defaultPositionTypeResolver) Resolve(yesShares, noShares int64) string {
	for _, rule := range defaultPositionTypeRules {
		if rule.matches(yesShares, noShares) {
			return rule.position
		}
	}
	return positionTypeNone
}

func forEachUserBet(bets []boundary.Bet, username string, visit func(boundary.Bet)) {
	for _, bet := range bets {
		if bet.Username != username {
			continue
		}
		visit(bet)
	}
}
