package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models/modelstesting"
)

func TestMigrateRemoveMarketDiscoverySectionsDropsTableAndColumn(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketDiscoveryCMS(db); err != nil {
		t.Fatalf("setup market discovery cms: %v", err)
	}
	if err := db.Exec("ALTER TABLE market_discovery_pages ADD COLUMN sections_enabled boolean DEFAULT false").Error; err != nil {
		t.Fatalf("add legacy sections_enabled column: %v", err)
	}
	if err := db.Exec(`CREATE TABLE market_discovery_sections (
		id integer primary key,
		page_id integer not null,
		slug text not null,
		title text not null
	)`).Error; err != nil {
		t.Fatalf("add legacy sections table: %v", err)
	}

	if err := migrations.MigrateRemoveMarketDiscoverySections(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	if db.Migrator().HasTable("market_discovery_sections") {
		t.Fatalf("expected market_discovery_sections to be removed")
	}
	if db.Migrator().HasColumn("market_discovery_pages", "sections_enabled") {
		t.Fatalf("expected sections_enabled to be removed")
	}
}

func TestMigrateRemoveMarketDiscoverySectionsIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketDiscoveryCMS(db); err != nil {
		t.Fatalf("setup market discovery cms: %v", err)
	}
	if err := migrations.MigrateRemoveMarketDiscoverySections(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateRemoveMarketDiscoverySections(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
