package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddReadModelStaleMetadataCreatesSnapshotColumnsAndTable(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddMarketAccountingSnapshots(db); err != nil {
		t.Fatalf("base market accounting migration: %v", err)
	}
	if err := migrations.MigrateAddUserFinancialMetricSnapshots(db); err != nil {
		t.Fatalf("base user financial migration: %v", err)
	}
	if err := migrations.MigrateAddReadModelStaleMetadata(db); err != nil {
		t.Fatalf("MigrateAddReadModelStaleMetadata returned error: %v", err)
	}

	for _, tc := range []struct {
		model  interface{}
		column string
	}{
		{&models.MarketAccountingSnapshot{}, "is_stale"},
		{&models.MarketAccountingSnapshot{}, "stale_reason"},
		{&models.MarketAccountingSnapshot{}, "marked_stale_at"},
		{&models.UserFinancialMetricSnapshot{}, "is_stale"},
		{&models.UserFinancialMetricSnapshot{}, "stale_reason"},
		{&models.UserFinancialMetricSnapshot{}, "marked_stale_at"},
	} {
		if !db.Migrator().HasColumn(tc.model, tc.column) {
			t.Fatalf("expected column %s on %T", tc.column, tc.model)
		}
	}

	if !db.Migrator().HasTable(&models.AnalyticsReadModelSnapshot{}) {
		t.Fatalf("expected analytics read-model snapshot table")
	}
}
