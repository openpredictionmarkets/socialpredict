package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddMarketGroups adds parent market-group storage for FEATURE/13
// multiple-choice binary markets. Existing markets remain the canonical child
// trading entities.
func MigrateAddMarketGroups(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.MarketGroup{},
		&models.MarketGroupMember{},
	)
}

func init() {
	migration.Register("20260612090000", func(db *gorm.DB) error {
		return MigrateAddMarketGroups(db)
	})
}
