package migrations_test

import (
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestAddCommentsMigration_CreatesTable(t *testing.T) {
	db := modelstesting.NewFakeDB(t) // runs all registered migrations including this one

	m := db.Migrator()

	if !m.HasTable(&models.Comment{}) {
		t.Fatal("expected comments table to exist after migration")
	}

	for _, col := range []string{"MarketID", "Username", "Content"} {
		if !m.HasColumn(&models.Comment{}, col) {
			t.Fatalf("expected comments.%s column to exist", col)
		}
	}
}
