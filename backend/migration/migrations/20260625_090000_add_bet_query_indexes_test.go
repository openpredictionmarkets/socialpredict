package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddBetQueryIndexesCreatesIndexes(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := db.AutoMigrate(&models.User{}, &models.Market{}, &models.Bet{}); err != nil {
		t.Fatalf("auto migrate core models: %v", err)
	}

	if err := migrations.MigrateAddBetQueryIndexes(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	for _, indexName := range []string{
		"idx_bets_market_id_placed_at_id",
		"idx_bets_market_id_username",
		"idx_bets_username_market_id_placed_at_id",
		"idx_bets_username_placed_at_id",
	} {
		if !db.Migrator().HasIndex(&models.Bet{}, indexName) {
			t.Fatalf("expected index %s after migration", indexName)
		}
	}
}

func TestMigrateAddBetQueryIndexesIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := db.AutoMigrate(&models.User{}, &models.Market{}, &models.Bet{}); err != nil {
		t.Fatalf("auto migrate core models: %v", err)
	}

	if err := migrations.MigrateAddBetQueryIndexes(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddBetQueryIndexes(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
