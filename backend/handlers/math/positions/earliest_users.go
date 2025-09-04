package positionsmath

import (
	"log"
	"time"

	"gorm.io/gorm"
)

type UserOrdered struct {
	Username    string
	YesShares   int64
	NoShares    int64
	EarliestBet string
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
