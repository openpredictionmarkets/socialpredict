package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddMarketAccountingSnapshots adds durable display/read-model snapshots
// for market accounting values. Raw bets remain the transaction source of truth.
func MigrateAddMarketAccountingSnapshots(db *gorm.DB) error {
	return db.AutoMigrate(&models.MarketAccountingSnapshot{})
}

func init() {
	migration.Register("20260607090000", func(db *gorm.DB) error {
		return MigrateAddMarketAccountingSnapshots(db)
	})
}
