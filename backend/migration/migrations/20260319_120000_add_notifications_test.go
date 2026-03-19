package migrations_test

import (
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestAddNotificationsMigration_CreatesTable(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	m := db.Migrator()

	if !m.HasTable(&models.Notification{}) {
		t.Fatal("expected notifications table to exist after migration")
	}

	for _, col := range []string{"Username", "Type", "MarketID", "Message", "IsRead"} {
		if !m.HasColumn(&models.Notification{}, col) {
			t.Fatalf("expected notifications.%s column to exist", col)
		}
	}
}
