package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddMarketDescriptionAmendments adds append-only amendment storage for
// market contract descriptions. The original markets.description value remains
// version 1; amendment rows start at version 2.
func MigrateAddMarketDescriptionAmendments(db *gorm.DB) error {
	return db.AutoMigrate(&models.MarketDescriptionAmendment{})
}

func init() {
	migration.Register("20260112090000", func(db *gorm.DB) error {
		return MigrateAddMarketDescriptionAmendments(db)
	})
}
