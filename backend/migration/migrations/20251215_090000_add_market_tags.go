package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddMarketTags adds admin-managed market taxonomy tags and the
// market-to-tag assignment table used by moderator proposals and navigation.
func MigrateAddMarketTags(db *gorm.DB) error {
	return db.AutoMigrate(&models.MarketTag{}, &models.MarketTagAssignment{})
}

func init() {
	migration.Register("20251215090000", func(db *gorm.DB) error {
		return MigrateAddMarketTags(db)
	})
}
