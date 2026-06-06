package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddBetDustCreatesColumn(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddBetDust(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasColumn(&models.Bet{}, "Dust") {
		t.Fatalf("expected Dust column after migration")
	}
}

func TestMigrateAddBetDustIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddBetDust(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddBetDust(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
