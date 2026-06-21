package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddMarketGroupAnswerAdditionsCreatesTableAndGovernanceColumn(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddMarketGroupAnswerAdditions(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasTable(&models.MarketGroupAnswerAddition{}) {
		t.Fatalf("expected market group answer additions table")
	}
	if !db.Migrator().HasIndex(&models.MarketGroupAnswerAddition{}, "idx_market_group_answer_additions_group_status") {
		t.Fatalf("expected group/status index")
	}
	if !db.Migrator().HasIndex(&models.MarketGroupAnswerAddition{}, "idx_market_group_answer_additions_status_created") {
		t.Fatalf("expected status/created index")
	}
	if !db.Migrator().HasColumn(&models.MarketGovernanceSettings{}, "AutoApproveMarketGroupAnswers") {
		t.Fatalf("expected AutoApproveMarketGroupAnswers governance setting")
	}
}

func TestMigrateAddMarketGroupAnswerAdditionsIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketGroupAnswerAdditions(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddMarketGroupAnswerAdditions(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
