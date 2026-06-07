package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddMarketGovernanceSettingsCreatesTable(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddMarketGovernanceSettings(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasTable(&models.MarketGovernanceSettings{}) {
		t.Fatalf("expected market governance settings table")
	}
	for _, column := range []string{"AutoApproveDescriptionAmendments", "AutoApproveMarketProposals", "Version", "UpdatedBy"} {
		if !db.Migrator().HasColumn(&models.MarketGovernanceSettings{}, column) {
			t.Fatalf("expected %s column after migration", column)
		}
	}
}

func TestMigrateAddMarketGovernanceSettingsIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketGovernanceSettings(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddMarketGovernanceSettings(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
