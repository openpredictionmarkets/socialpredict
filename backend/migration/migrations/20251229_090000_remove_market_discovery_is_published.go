package migrations

import (
	"socialpredict/migration"

	"gorm.io/gorm"
)

// MigrateRemoveMarketDiscoveryIsPublished removes the unused CMS draft/publish
// flag from market discovery pages. Market discovery CMS edits publish
// immediately, so this column only added confusion.
func MigrateRemoveMarketDiscoveryIsPublished(db *gorm.DB) error {
	m := db.Migrator()
	if !m.HasTable("market_discovery_pages") || !m.HasColumn("market_discovery_pages", "is_published") {
		return nil
	}
	if db.Dialector.Name() == "postgres" {
		return db.Exec("ALTER TABLE market_discovery_pages DROP COLUMN IF EXISTS is_published").Error
	}
	return db.Exec("ALTER TABLE market_discovery_pages DROP COLUMN is_published").Error
}

func init() {
	migration.Register("20251229090000", func(db *gorm.DB) error {
		return MigrateRemoveMarketDiscoveryIsPublished(db)
	})
}
