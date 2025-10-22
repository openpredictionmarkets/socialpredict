package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// migrateAddMarketLabels contains the core logic so we can unit test it directly.
func MigrateAddMarketLabels(db *gorm.DB) error {
	m := db.Migrator()

	// 1) Add columns if missing (snake_cased by GORM -> yes_label / no_label)
	if !m.HasColumn(&models.Market{}, "YesLabel") {
		if err := m.AddColumn(&models.Market{}, "YesLabel"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&models.Market{}, "NoLabel") {
		if err := m.AddColumn(&models.Market{}, "NoLabel"); err != nil {
			return err
		}
	}

	// 2) Backfill defaults for existing rows (empty or NULL -> "YES"/"NO")
	if err := db.Model(&models.Market{}).
		Where("yes_label IS NULL OR yes_label = ''").
		Update("yes_label", "YES").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Market{}).
		Where("no_label IS NULL OR no_label = ''").
		Update("no_label", "NO").Error; err != nil {
		return err
	}

	return nil
}

// Register the migration with a timestamp. Adjust the timestamp if you need strict ordering vs. other migrations.
func init() {
	migration.Register("20251020140500", func(db *gorm.DB) error {
		return MigrateAddMarketLabels(db)
	})
}
