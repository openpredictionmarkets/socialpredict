package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddMarketAccountingSnapshotsCreatesTable(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddMarketAccountingSnapshots(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasTable(&models.MarketAccountingSnapshot{}) {
		t.Fatalf("expected market accounting snapshots table")
	}
	for _, column := range []string{
		"MarketID",
		"LastProbability",
		"NetBetVolume",
		"MarketDust",
		"VolumeWithDust",
		"UserCount",
		"BetCount",
		"LastProcessedBetID",
		"LastProcessedBetAt",
		"GeneratedAt",
		"Source",
	} {
		if !db.Migrator().HasColumn(&models.MarketAccountingSnapshot{}, column) {
			t.Fatalf("expected %s column after migration", column)
		}
	}
	if !db.Migrator().HasIndex(&models.MarketAccountingSnapshot{}, "idx_market_accounting_snapshots_market_id") {
		t.Fatalf("expected market id unique index")
	}
}

func TestMigrateAddMarketAccountingSnapshotsIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketAccountingSnapshots(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddMarketAccountingSnapshots(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
