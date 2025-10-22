package migrations_test

import (
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestCoreModelsMigration_CreatesTablesAndColumns(t *testing.T) {
	db := modelstesting.NewFakeDB(t) // runs global migrations incl. this file's init-registered one

	m := db.Migrator()

	// Tables
	for _, tbl := range []any{&models.User{}, &models.Market{}, &models.Bet{}, &models.HomepageContent{}} {
		if !m.HasTable(tbl) {
			t.Fatalf("expected table for %T to exist", tbl)
		}
	}

	// Critical market columns (v2 change)
	if !m.HasColumn(&models.Market{}, "YesLabel") {
		t.Fatalf("expected markets.yes_label column to exist")
	}
	if !m.HasColumn(&models.Market{}, "NoLabel") {
		t.Fatalf("expected markets.no_label column to exist")
	}
}
