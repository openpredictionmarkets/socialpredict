package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddUserFinancialMetricSnapshots adds durable authenticated
// display/read-model snapshots for individual user financial metrics.
func MigrateAddUserFinancialMetricSnapshots(db *gorm.DB) error {
	return db.AutoMigrate(&models.UserFinancialMetricSnapshot{})
}

func init() {
	migration.Register("20260607091000", func(db *gorm.DB) error {
		return MigrateAddUserFinancialMetricSnapshots(db)
	})
}
