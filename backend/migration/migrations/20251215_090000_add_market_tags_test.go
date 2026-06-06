package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddMarketTagsCreatesTagTables(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddMarketTags(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasTable(&models.MarketTag{}) {
		t.Fatalf("expected market tags table")
	}
	if !db.Migrator().HasTable(&models.MarketTagAssignment{}) {
		t.Fatalf("expected market tag assignments table")
	}
	if !db.Migrator().HasIndex(&models.MarketTag{}, "idx_market_tags_slug") {
		t.Fatalf("expected unique slug index")
	}
}

func TestMigrateAddMarketTagsIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketTags(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddMarketTags(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
