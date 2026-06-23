package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddUserFinancialMetricSnapshotsCreatesTable(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddUserFinancialMetricSnapshots(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasTable(&models.UserFinancialMetricSnapshot{}) {
		t.Fatalf("expected user financial metric snapshots table")
	}
	for _, column := range []string{
		"Username",
		"AccountBalance",
		"MaximumDebtAllowed",
		"AmountInPlay",
		"AmountBorrowed",
		"RetainedEarnings",
		"Equity",
		"TradingProfits",
		"WorkProfits",
		"UnrealizedWorkIncome",
		"UnrealizedWorkProfits",
		"TotalProfits",
		"AmountInPlayActive",
		"TotalSpent",
		"TotalSpentInPlay",
		"RealizedProfits",
		"PotentialProfits",
		"RealizedValue",
		"PotentialValue",
		"PositionCount",
		"GeneratedAt",
		"Source",
	} {
		if !db.Migrator().HasColumn(&models.UserFinancialMetricSnapshot{}, column) {
			t.Fatalf("expected %s column after migration", column)
		}
	}
	if !db.Migrator().HasIndex(&models.UserFinancialMetricSnapshot{}, "idx_user_financial_metric_snapshots_username") {
		t.Fatalf("expected username unique index")
	}
}

func TestMigrateAddUserFinancialMetricSnapshotsIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddUserFinancialMetricSnapshots(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddUserFinancialMetricSnapshots(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
