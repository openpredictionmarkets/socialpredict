package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddMarketDescriptionAmendmentsCreatesTable(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddMarketDescriptionAmendments(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasTable(&models.MarketDescriptionAmendment{}) {
		t.Fatalf("expected market description amendments table")
	}
	for _, column := range []string{"MarketID", "Version", "Body", "BodyFormat", "Status", "CreatedBy", "ApprovedBy", "ApprovedAt", "RejectedBy", "RejectedAt", "RejectionReason", "SubmitReason"} {
		if !db.Migrator().HasColumn(&models.MarketDescriptionAmendment{}, column) {
			t.Fatalf("expected %s column after migration", column)
		}
	}
}

func TestMigrateAddMarketDescriptionAmendmentsIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketDescriptionAmendments(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddMarketDescriptionAmendments(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
