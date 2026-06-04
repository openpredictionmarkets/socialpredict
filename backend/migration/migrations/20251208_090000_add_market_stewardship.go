package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddMarketStewardship adds current-steward storage while preserving
// immutable creator attribution. Existing markets backfill steward to creator.
func MigrateAddMarketStewardship(db *gorm.DB) error {
	m := db.Migrator()

	if !m.HasColumn(&models.Market{}, "StewardUsername") {
		if err := m.AddColumn(&models.Market{}, "StewardUsername"); err != nil {
			return err
		}
	}

	if err := db.Model(&models.Market{}).
		Where("steward_username IS NULL OR steward_username = ''").
		Update("steward_username", gorm.Expr("creator_username")).Error; err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.MarketStewardshipAudit{}); err != nil {
		return err
	}

	return nil
}

func init() {
	migration.Register("20251208090000", func(db *gorm.DB) error {
		return MigrateAddMarketStewardship(db)
	})
}
