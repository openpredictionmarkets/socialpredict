package positionsmath

import (
	"sort"
	"time"

	"gorm.io/gorm"
)

// UserHolder is for sorting only—combines valuation and earliest bet.
type UserHolder struct {
	Username     string
	RoundedValue int64
	EarliestBet  time.Time
}

type ByValBetTimeUsername []UserHolder

func (s ByValBetTimeUsername) Len() int      { return len(s) }
func (s ByValBetTimeUsername) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByValBetTimeUsername) Less(i, j int) bool {
	return ranksBefore(s[i], s[j])
}

// AdjustUserValuationsToMarketVolume ensures user values match total market volume,
// distributing rounding delta deterministically. Only users with >0 value are adjusted.
func AdjustUserValuationsToMarketVolume(
	db *gorm.DB,
	marketID uint,
	userValuations map[string]UserValuationResult,
	targetMarketVolume int64,
) (map[string]UserValuationResult, error) {
	filtered := filterWinningValuations(userValuations)
	if len(filtered) == 0 {
		return userValuations, nil
	}

	earliestBets, err := GetAllUserEarliestBetsForMarket(db, marketID)
	if err != nil {
		return nil, err
	}

	holders, sum := buildUserHolders(filtered, earliestBets)
	return adjustValuations(filtered, holders, sum, targetMarketVolume), nil
}

// filterWinningValuations keeps only users eligible for deterministic delta adjustment.
func filterWinningValuations(all map[string]UserValuationResult) map[string]UserValuationResult {
	filtered := make(map[string]UserValuationResult)
	for username, val := range all {
		if val.RoundedValue > 0 {
			filtered[username] = val
		}
	}
	return filtered
}

// buildUserHolders converts user values into sortable holders and returns their total value.
func buildUserHolders(
	userVals map[string]UserValuationResult,
	earliest map[string]time.Time,
) ([]UserHolder, int64) {
	var (
		holders []UserHolder
		sum     int64
	)
	for username, val := range userVals {
		sum += val.RoundedValue
		holders = append(holders, UserHolder{
			Username:     username,
			RoundedValue: val.RoundedValue,
			EarliestBet:  earliest[username],
		})
	}
	sortUserHolders(holders)
	return holders, sum
}

func sortUserHolders(holders []UserHolder) {
	sort.Sort(ByValBetTimeUsername(holders))
}

// adjustValuations applies the market-level rounding delta to eligible users.
func adjustValuations(
	userVals map[string]UserValuationResult,
	holders []UserHolder,
	sum int64,
	target int64,
) map[string]UserValuationResult {
	adjusted := cloneUserValuations(userVals)
	delta, adjustment := normalizeDelta(target - sum)
	if delta == 0 || len(holders) == 0 {
		return adjusted
	}

	holderCount := int64(len(holders))
	for i := int64(0); i < delta; i++ {
		holder := holders[i%holderCount]
		val := adjusted[holder.Username]
		val.RoundedValue += adjustment
		adjusted[holder.Username] = val
	}
	return adjusted
}

func cloneUserValuations(userVals map[string]UserValuationResult) map[string]UserValuationResult {
	cloned := make(map[string]UserValuationResult, len(userVals))
	for username, valuation := range userVals {
		cloned[username] = valuation
	}
	return cloned
}

func normalizeDelta(delta int64) (int64, int64) {
	if delta == 0 {
		return 0, 0
	}
	if delta < 0 {
		return -delta, -1
	}
	return delta, 1
}

func ranksBefore(left, right UserHolder) bool {
	if left.RoundedValue != right.RoundedValue {
		return left.RoundedValue > right.RoundedValue
	}
	if !left.EarliestBet.Equal(right.EarliestBet) {
		return left.EarliestBet.Before(right.EarliestBet)
	}
	return left.Username < right.Username
}
