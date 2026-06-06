package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddMarketGovernanceSettings adds singleton market governance settings
// used by admin-controlled operational policies such as amendment auto-approval.
func MigrateAddMarketGovernanceSettings(db *gorm.DB) error {
	return db.AutoMigrate(&models.MarketGovernanceSettings{})
}

func init() {
	migration.Register("20260119090000", func(db *gorm.DB) error {
		return MigrateAddMarketGovernanceSettings(db)
	})
}
