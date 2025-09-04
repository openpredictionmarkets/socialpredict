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
	if s[i].RoundedValue != s[j].RoundedValue {
		return s[i].RoundedValue > s[j].RoundedValue
	}
	if !s[i].EarliestBet.Equal(s[j].EarliestBet) {
		return s[i].EarliestBet.Before(s[j].EarliestBet)
	}
	return s[i].Username < s[j].Username
}

// AdjustUserValuationsToMarketVolume ensures user values match total market volume,
// distributing rounding delta deterministically. Only users with >0 value are adjusted.
func AdjustUserValuationsToMarketVolume(
	db *gorm.DB,
	marketID uint,
	userValuations map[string]UserValuationResult,
	targetMarketVolume int64,
) (map[string]UserValuationResult, error) {
	// Filter out users with zero valuation
	filtered := filterWinningValuations(userValuations)
	if len(filtered) == 0 {
		return userValuations, nil
	}

	// Fetch earliest bets for ordering
	earliestBets, err := GetAllUserEarliestBetsForMarket(db, marketID)
	if err != nil {
		return nil, err
	}

	// Create sortable holder list
	holders, sum := buildUserHolders(filtered, earliestBets)

	// Apply delta correction
	adjusted := adjustValuations(filtered, holders, sum, targetMarketVolume)

	return adjusted, nil
}

// filterWinningValuations drops users with zero rounded value
func filterWinningValuations(all map[string]UserValuationResult) map[string]UserValuationResult {
	filtered := make(map[string]UserValuationResult)
	for username, val := range all {
		if val.RoundedValue > 0 {
			filtered[username] = val
		}
	}
	return filtered
}

// buildUserHolders prepares sorted holders and computes total value sum
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
	sort.Sort(ByValBetTimeUsername(holders))
	return holders, sum
}

// adjustValuations spreads rounding delta among holders
func adjustValuations(
	userVals map[string]UserValuationResult,
	holders []UserHolder,
	sum int64,
	target int64,
) map[string]UserValuationResult {
	delta := target - sum
	if delta == 0 {
		return userVals
	}

	adjustment := int64(1)
	if delta < 0 {
		adjustment = -1
		delta = -delta
	}

	holderCount := int64(len(holders))
	for i := int64(0); i < delta; i++ {
		holder := holders[i%holderCount]
		val := userVals[holder.Username]
		val.RoundedValue += adjustment
		userVals[holder.Username] = val
	}
	return userVals
}
