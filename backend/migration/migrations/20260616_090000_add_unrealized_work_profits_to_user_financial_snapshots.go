package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddUnrealizedWorkProfitsToUserFinancialSnapshots adds display-only
// forecast fields for unresolved stewarded market work income and profit.
func MigrateAddUnrealizedWorkProfitsToUserFinancialSnapshots(db *gorm.DB) error {
	return db.AutoMigrate(&models.UserFinancialMetricSnapshot{})
}

func init() {
	migration.Register("20260616090000", func(db *gorm.DB) error {
		return MigrateAddUnrealizedWorkProfitsToUserFinancialSnapshots(db)
	})
}
