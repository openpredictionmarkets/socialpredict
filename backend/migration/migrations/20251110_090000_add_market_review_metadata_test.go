package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddMarketReviewMetadataAddsReviewColumns(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	for _, column := range []string{"ApprovedBy", "ApprovedAt", "RejectedBy", "RejectedAt", "RejectionReason"} {
		_ = db.Migrator().DropColumn(&models.Market{}, column)
	}

	if err := migrations.MigrateAddMarketReviewMetadata(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	for _, column := range []string{"ApprovedBy", "ApprovedAt", "RejectedBy", "RejectedAt", "RejectionReason"} {
		if !db.Migrator().HasColumn(&models.Market{}, column) {
			t.Fatalf("expected %s column after migration", column)
		}
	}
}
