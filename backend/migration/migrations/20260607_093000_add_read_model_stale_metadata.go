package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddReadModelStaleMetadata adds stale markers to existing display
// snapshots and creates a generic analytics snapshot table for aggregate read
// models. These tables are display-only and must not become transaction truth.
func MigrateAddReadModelStaleMetadata(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.MarketAccountingSnapshot{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&models.UserFinancialMetricSnapshot{}); err != nil {
		return err
	}
	return db.AutoMigrate(&models.AnalyticsReadModelSnapshot{})
}

func init() {
	migration.Register("20260607093000", func(db *gorm.DB) error {
		return MigrateAddReadModelStaleMetadata(db)
	})
}
