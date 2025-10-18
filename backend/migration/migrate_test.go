package migration_test

import (
	"testing"
	"time"

	"socialpredict/migration"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

func TestRegister_DuplicatePanics(t *testing.T) {
	migration.ClearRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on duplicate migration id, got none")
		}
	}()

	migration.Register("20250101000000", func(db *gorm.DB) error { return nil })
	// Second registration with the same ID should panic.
	migration.Register("20250101000000", func(db *gorm.DB) error { return nil })
}

func TestRun_AppliesInOrder_And_Persists(t *testing.T) {
	migration.ClearRegistry()
	db := modelstesting.NewTestDB(t)

	var calls []string

	// Register intentionally out of order; Run must apply in ascending lexicographic order.
	migration.Register("20250103000000", func(db *gorm.DB) error {
		calls = append(calls, "20250103000000")
		// Touch DB to ensure Up() runs a real operation
		return db.AutoMigrate(&migration.SchemaMigration{})
	})
	migration.Register("20250101000000", func(db *gorm.DB) error {
		calls = append(calls, "20250101000000")
		return nil
	})
	migration.Register("20250102000000", func(db *gorm.DB) error {
		calls = append(calls, "20250102000000")
		return nil
	})

	if err := migration.Run(db); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := []string{"20250101000000", "20250102000000", "20250103000000"}
	if len(calls) != len(want) {
		t.Fatalf("unexpected call count: got %d want %d", len(calls), len(want))
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("order[%d]: got %s want %s", i, calls[i], want[i])
		}
	}

	var rows []migration.SchemaMigration
	if err := db.Order("id asc").Find(&rows).Error; err != nil {
		t.Fatalf("query SchemaMigration: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 SchemaMigration rows, got %d", len(rows))
	}
	for i := range want {
		if rows[i].ID != want[i] {
			t.Fatalf("row[%d].ID: got %s want %s", i, rows[i].ID, want[i])
		}
		if rows[i].AppliedAt.IsZero() {
			t.Fatalf("row[%d].AppliedAt is zero", i)
		}
	}
}

func TestRun_IsIdempotent(t *testing.T) {
	migration.ClearRegistry()
	db := modelstesting.NewFakeDB(t)

	var calls []string
	migration.Register("20250101000000", func(db *gorm.DB) error {
		calls = append(calls, "20250101000000")
		return nil
	})

	// First run applies once.
	if err := migration.Run(db); err != nil {
		t.Fatalf("Run (first): %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("after first run, calls=%d want 1", len(calls))
	}

	// Second run should skip already-applied migration.
	if err := migration.Run(db); err != nil {
		t.Fatalf("Run (second): %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("after second run, calls=%d want 1 (idempotent)", len(calls))
	}

	var rows []migration.SchemaMigration
	if err := db.Find(&rows).Error; err != nil {
		t.Fatalf("query SchemaMigration: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != "20250101000000" {
		t.Fatalf("unexpected SchemaMigration state: %+v", rows)
	}
}

func TestRun_ErrOnNilUp(t *testing.T) {
	migration.ClearRegistry()
	db := modelstesting.NewFakeDB(t)

	migration.Register("20250102000000", nil)

	if err := migration.Run(db); err == nil {
		t.Fatalf("expected error for nil Up(), got nil")
	}
}

func TestRun_PersistsAppliedAt(t *testing.T) {
	migration.ClearRegistry()
	db := modelstesting.NewFakeDB(t)

	migration.Register("20250104000000", func(db *gorm.DB) error { return nil })

	if err := migration.Run(db); err != nil {
		t.Fatalf("Run: %v", err)
	}

	var row migration.SchemaMigration
	if err := db.First(&row, "id = ?", "20250104000000").Error; err != nil {
		t.Fatalf("lookup SchemaMigration: %v", err)
	}
	if row.AppliedAt.IsZero() || time.Since(row.AppliedAt) < 0 {
		t.Fatalf("AppliedAt not set correctly: %+v", row)
	}
}
