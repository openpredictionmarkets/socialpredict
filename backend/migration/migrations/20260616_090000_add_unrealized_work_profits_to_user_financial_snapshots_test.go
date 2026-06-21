package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddUnrealizedWorkProfitsToUserFinancialSnapshotsAddsColumn(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddUserFinancialMetricSnapshots(db); err != nil {
		t.Fatalf("base migration failed: %v", err)
	}
	if err := db.Migrator().DropColumn(&models.UserFinancialMetricSnapshot{}, "UnrealizedWorkIncome"); err != nil {
		t.Fatalf("drop unrealized work income column: %v", err)
	}
	if err := db.Migrator().DropColumn(&models.UserFinancialMetricSnapshot{}, "UnrealizedWorkProfits"); err != nil {
		t.Fatalf("drop unrealized work profits column: %v", err)
	}

	if err := migrations.MigrateAddUnrealizedWorkProfitsToUserFinancialSnapshots(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasColumn(&models.UserFinancialMetricSnapshot{}, "UnrealizedWorkIncome") {
		t.Fatalf("expected UnrealizedWorkIncome column after migration")
	}
	if !db.Migrator().HasColumn(&models.UserFinancialMetricSnapshot{}, "UnrealizedWorkProfits") {
		t.Fatalf("expected UnrealizedWorkProfits column after migration")
	}
}
