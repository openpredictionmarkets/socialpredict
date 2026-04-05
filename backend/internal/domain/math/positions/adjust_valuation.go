package positionsmath

import (
	"sort"
	"time"
)

// UserHolder carries the ordering data used for deterministic valuation adjustments.
type UserHolder struct {
	Username     string
	RoundedValue int64
	EarliestBet  time.Time
}

type UserValuationAdjuster interface {
	Adjust(
		userValuations map[string]UserValuationResult,
		earliestBets map[string]time.Time,
		targetMarketVolume int64,
	) map[string]UserValuationResult
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

type deterministicValuationAdjuster struct{}

var defaultUserValuationAdjuster UserValuationAdjuster = deterministicValuationAdjuster{}

// AdjustUserValuationsToMarketVolume ensures user values match total market volume,
// distributing rounding delta deterministically. Only users with >0 value are adjusted.
func AdjustUserValuationsToMarketVolume(
	userValuations map[string]UserValuationResult,
	earliestBets map[string]time.Time,
	targetMarketVolume int64,
) map[string]UserValuationResult {
	return defaultUserValuationAdjuster.Adjust(userValuations, earliestBets, targetMarketVolume)
}

func (deterministicValuationAdjuster) Adjust(
	userValuations map[string]UserValuationResult,
	earliestBets map[string]time.Time,
	targetMarketVolume int64,
) map[string]UserValuationResult {
	filtered := filterWinningValuations(userValuations)
	if len(filtered) == 0 {
		return userValuations
	}

	holders, sum := buildUserHolders(filtered, earliestBets)
	return adjustValuations(filtered, holders, sum, targetMarketVolume)
}

// filterWinningValuations returns only users eligible for rounding adjustments.
func filterWinningValuations(all map[string]UserValuationResult) map[string]UserValuationResult {
	filtered := make(map[string]UserValuationResult, len(all))
	for username, val := range all {
		if val.RoundedValue > 0 {
			filtered[username] = val
		}
	}
	return filtered
}

// buildUserHolders builds the deterministic adjustment order and total adjusted value.
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
		holders = append(holders, newUserHolder(username, val.RoundedValue, earliest[username]))
	}
	sort.Sort(ByValBetTimeUsername(holders))
	return holders, sum
}

func newUserHolder(username string, roundedValue int64, earliestBet time.Time) UserHolder {
	return UserHolder{
		Username:     username,
		RoundedValue: roundedValue,
		EarliestBet:  earliestBet,
	}
}

// adjustValuations spreads the rounding delta across the sorted holders.
func adjustValuations(
	userVals map[string]UserValuationResult,
	holders []UserHolder,
	sum int64,
	target int64,
) map[string]UserValuationResult {
	delta, adjustment := valuationDelta(sum, target)
	if delta == 0 {
		return userVals
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

func valuationDelta(currentSum, target int64) (int64, int64) {
	delta := target - currentSum
	if delta < 0 {
		return -delta, -1
	}
	return delta, 1
}
