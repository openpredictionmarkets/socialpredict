package migrations

import (
	"socialpredict/migration"

	"gorm.io/gorm"
)

// MigrateRemoveMarketDiscoverySections removes the unused FEATURE/09 section
// scaffold. Market discovery now uses TOP topic pins plus page-level market
// pins, so retaining sections would keep dead CMS functionality around.
func MigrateRemoveMarketDiscoverySections(db *gorm.DB) error {
	m := db.Migrator()
	if m.HasTable("market_discovery_sections") {
		if err := m.DropTable("market_discovery_sections"); err != nil {
			return err
		}
	}
	if !m.HasTable("market_discovery_pages") || !m.HasColumn("market_discovery_pages", "sections_enabled") {
		return nil
	}
	if db.Dialector.Name() == "postgres" {
		return db.Exec("ALTER TABLE market_discovery_pages DROP COLUMN IF EXISTS sections_enabled").Error
	}
	return db.Exec("ALTER TABLE market_discovery_pages DROP COLUMN sections_enabled").Error
}

func init() {
	migration.Register("20260105090000", func(db *gorm.DB) error {
		return MigrateRemoveMarketDiscoverySections(db)
	})
}
