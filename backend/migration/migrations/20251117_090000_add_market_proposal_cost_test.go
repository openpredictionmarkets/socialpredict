package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddMarketProposalCostAddsProposalCostColumn(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_ = db.Migrator().DropColumn(&models.Market{}, "ProposalCost")

	if err := migrations.MigrateAddMarketProposalCost(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasColumn(&models.Market{}, "ProposalCost") {
		t.Fatalf("expected ProposalCost column after migration")
	}
}
