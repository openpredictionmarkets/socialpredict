package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models/modelstesting"
)

func TestMigrateRemoveMarketDiscoveryIsPublishedDropsColumn(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketDiscoveryCMS(db); err != nil {
		t.Fatalf("setup market discovery cms: %v", err)
	}
	if err := db.Exec("ALTER TABLE market_discovery_pages ADD COLUMN is_published boolean DEFAULT true").Error; err != nil {
		t.Fatalf("add legacy is_published column: %v", err)
	}
	if !db.Migrator().HasColumn("market_discovery_pages", "is_published") {
		t.Fatalf("expected setup schema to include is_published")
	}
	if err := migrations.MigrateRemoveMarketDiscoveryIsPublished(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	if db.Migrator().HasColumn("market_discovery_pages", "is_published") {
		t.Fatalf("expected is_published to be removed")
	}
}

func TestMigrateRemoveMarketDiscoveryIsPublishedIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketDiscoveryCMS(db); err != nil {
		t.Fatalf("setup market discovery cms: %v", err)
	}
	if err := migrations.MigrateRemoveMarketDiscoveryIsPublished(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateRemoveMarketDiscoveryIsPublished(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
