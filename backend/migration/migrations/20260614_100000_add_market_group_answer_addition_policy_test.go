package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddMarketGroupAnswerAdditionPolicyCreatesColumn(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddMarketGroupAnswerAdditionPolicy(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasColumn(&models.MarketGroup{}, "AutoApproveAnswerAdditions") {
		t.Fatalf("expected market_groups.auto_approve_answer_additions column after migration")
	}
}

func TestMigrateAddMarketGroupAnswerAdditionPolicyIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddMarketGroupAnswerAdditionPolicy(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddMarketGroupAnswerAdditionPolicy(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
