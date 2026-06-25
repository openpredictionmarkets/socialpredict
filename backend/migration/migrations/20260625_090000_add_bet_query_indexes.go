package migrations

import (
	"socialpredict/migration"

	"gorm.io/gorm"
)

// MigrateAddBetQueryIndexes adds composite indexes for canonical bet-history
// replay and common market/user lookup paths. These indexes add write/storage
// cost on each bet insert, but keep high-volume reads scoped to the intended
// market or user boundary.
func MigrateAddBetQueryIndexes(db *gorm.DB) error {
	indexStatements := []string{
		"CREATE INDEX IF NOT EXISTS idx_bets_market_id_placed_at_id ON bets (market_id, placed_at, id)",
		"CREATE INDEX IF NOT EXISTS idx_bets_market_id_username ON bets (market_id, username)",
		"CREATE INDEX IF NOT EXISTS idx_bets_username_market_id_placed_at_id ON bets (username, market_id, placed_at, id)",
	}

	for _, statement := range indexStatements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func init() {
	migration.Register("20260625090000", func(db *gorm.DB) error {
		return MigrateAddBetQueryIndexes(db)
	})
}
