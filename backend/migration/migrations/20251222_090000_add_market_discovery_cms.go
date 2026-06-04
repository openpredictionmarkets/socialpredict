package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddMarketDiscoveryCMS adds CMS-owned market discovery page,
// section, and pin tables for FEATURE/09 taxonomy navigation.
func MigrateAddMarketDiscoveryCMS(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.MarketDiscoveryPage{},
		&models.MarketDiscoverySection{},
		&models.MarketDiscoveryPin{},
	)
}

func init() {
	migration.Register("20251222090000", func(db *gorm.DB) error {
		return MigrateAddMarketDiscoveryCMS(db)
	})
}
