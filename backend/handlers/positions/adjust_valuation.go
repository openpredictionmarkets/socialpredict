package positions

import (
	"log"
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

type UserOrdered struct {
	Username    string
	YesShares   int64
	NoShares    int64
	EarliestBet string
}

// ByValBetTimeUsername: value DESC, then earliest bet ASC, then username ASC
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

// AdjustUserValuationsToMarketVolume ensures the sum of user valuations matches the market volume
// by distributing delta units one at a time among holders, sorted deterministically.
func AdjustUserValuationsToMarketVolume(
	db *gorm.DB,
	marketID uint,
	userValuations map[string]UserValuationResult,
	targetMarketVolume int64,
) (map[string]UserValuationResult, error) {
	// Get earliest bet times for all users in the market, in one efficient query
	earliestBetMap, err := GetAllUserEarliestBetsForMarket(db, marketID)
	if err != nil {
		return nil, err
	}

	var sumOfRoundedValues int64
	holders := make([]UserHolder, 0, len(userValuations))
	for username, valuation := range userValuations {
		sumOfRoundedValues += valuation.RoundedValue
		holders = append(holders, UserHolder{
			Username:     username,
			RoundedValue: valuation.RoundedValue,
			EarliestBet:  earliestBetMap[username], // Always valid since map is built from market bets
		})
	}

	delta := targetMarketVolume - sumOfRoundedValues
	if delta == 0 {
		return userValuations, nil
	}

	sort.Sort(ByValBetTimeUsername(holders))

	adjustment := int64(1)
	if delta < 0 {
		adjustment = -1
		delta = -delta
	}

	holderCount := int64(len(holders))
	for distributed := int64(0); distributed < delta; distributed++ {
		holder := holders[distributed%holderCount]
		current := userValuations[holder.Username]
		current.RoundedValue += adjustment
		userValuations[holder.Username] = current
	}

	return userValuations, nil
}

// GetAllUserEarliestBetsForMarket returns a map of usernames to their earliest bet timestamp in the given market.
func GetAllUserEarliestBetsForMarket(db *gorm.DB, marketID uint) (map[string]time.Time, error) {
	var ordered []UserOrdered
	err := db.Raw(`
		SELECT username,
			SUM(CASE WHEN outcome = 'YES' THEN amount ELSE 0 END) as yes_shares,
			SUM(CASE WHEN outcome = 'NO' THEN amount ELSE 0 END) as no_shares,
			MIN(placed_at) as earliest_bet
		FROM bets
		WHERE market_id = ?
		GROUP BY username
		ORDER BY (SUM(CASE WHEN outcome = 'YES' THEN amount ELSE 0 END) +
				SUM(CASE WHEN outcome = 'NO' THEN amount ELSE 0 END)) DESC,
				earliest_bet ASC,
				username ASC
	`, marketID).Scan(&ordered).Error
	if err != nil {
		return nil, err
	}

	m := make(map[string]time.Time)
	for _, b := range ordered {
		var t time.Time
		var err error
		layouts := []string{
			time.RFC3339Nano,                   // try RFC first
			"2006-01-02 15:04:05.999999999",    // sqlite default (no TZ)
			"2006-01-02 15:04:05.999999-07:00", // your case!
			"2006-01-02 15:04:05",              // fallback, no fraction or tz
		}
		for _, layout := range layouts {
			t, err = time.Parse(layout, b.EarliestBet)
			if err == nil {
				break
			}
		}
		if err != nil {
			log.Printf("Could not parse time %q for user %s: %v", b.EarliestBet, b.Username, err)
			continue
		}
		m[b.Username] = t
	}
	return m, nil
}
