package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

func MigrateAddMarketLifecycleStatus(db *gorm.DB) error {
	m := db.Migrator()

	if !m.HasColumn(&models.Market{}, "LifecycleStatus") {
		if err := m.AddColumn(&models.Market{}, "LifecycleStatus"); err != nil {
			return err
		}
	}

	if err := db.Model(&models.Market{}).
		Where("lifecycle_status IS NULL OR lifecycle_status = ''").
		Update("lifecycle_status", "published").Error; err != nil {
		return err
	}

	return db.Model(&models.Market{}).
		Where("is_resolved = ?", true).
		Update("lifecycle_status", "resolved").Error
}

func init() {
	migration.Register("20251103090000", func(db *gorm.DB) error {
		return MigrateAddMarketLifecycleStatus(db)
	})
}
