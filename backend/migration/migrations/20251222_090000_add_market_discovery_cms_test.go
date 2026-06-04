package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddMarketDiscoveryCMSCreatesTables(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddMarketDiscoveryCMS(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	for _, table := range []any{
		&models.MarketDiscoveryPage{},
		&models.MarketDiscoverySection{},
		&models.MarketDiscoveryPin{},
	} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("expected table for %T", table)
		}
	}
	if !db.Migrator().HasIndex(&models.MarketDiscoveryPage{}, "idx_market_discovery_pages_slug") {
		t.Fatalf("expected market discovery page slug index")
	}
	if !db.Migrator().HasColumn(&models.MarketDiscoveryPin{}, "target_page_slug") {
		t.Fatalf("expected market discovery pin target_page_slug column")
	}
}

func TestMigrateAddMarketDiscoveryCMSIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketDiscoveryCMS(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddMarketDiscoveryCMS(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
