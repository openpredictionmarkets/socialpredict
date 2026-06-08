package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddMarketProposalAutoApprovalCreatesColumn(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddMarketGovernanceSettings(db); err != nil {
		t.Fatalf("base migration failed: %v", err)
	}
	if err := migrations.MigrateAddMarketProposalAutoApproval(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasColumn(&models.MarketGovernanceSettings{}, "AutoApproveMarketProposals") {
		t.Fatalf("expected AutoApproveMarketProposals column after migration")
	}
}

func TestMigrateAddMarketProposalAutoApprovalIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketProposalAutoApproval(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddMarketProposalAutoApproval(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
