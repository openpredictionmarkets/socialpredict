package migrations_test

import (
	"testing"
	"time"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

func seedMinimalMarket(t *testing.T, db *gorm.DB) models.Market {
	t.Helper()
	// Use your helper if present; otherwise a minimal inline seed.
	m := modelstesting.GenerateMarket(1, "alice")
	// Ensure zero-values for new fields (old DB rows would have had no columns)
	m.YesLabel = ""
	m.NoLabel = ""
	if err := db.Create(&m).Error; err != nil {
		t.Fatalf("failed to seed market: %v", err)
	}
	return m
}

func TestMigrateAddMarketLabels_AddsColumnsAndBackfills(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	// Simulate pre-migration DB: create a market before columns exist.
	// (NewFakeDB runs full AutoMigrate on current models; if your test DB
	// already includes columns, force-drop them to mimic v1 schema.)
	_ = db.Migrator().DropColumn(&models.Market{}, "YesLabel")
	_ = db.Migrator().DropColumn(&models.Market{}, "NoLabel")

	// Now seed data as it would exist in v1.x (no labels)
	m := seedMinimalMarket(t, db)

	// Run the migration under test
	if err := migrations.MigrateAddMarketLabels(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Columns should exist
	if !db.Migrator().HasColumn(&models.Market{}, "YesLabel") {
		t.Fatalf("expected yes_label column to exist after migration")
	}
	if !db.Migrator().HasColumn(&models.Market{}, "NoLabel") {
		t.Fatalf("expected no_label column to exist after migration")
	}

	// Backfill should have applied
	var out models.Market
	if err := db.First(&out, m.ID).Error; err != nil {
		t.Fatalf("failed to load market after migration: %v", err)
	}
	if out.YesLabel != "YES" {
		t.Fatalf("expected YesLabel to be backfilled to YES, got %q", out.YesLabel)
	}
	if out.NoLabel != "NO" {
		t.Fatalf("expected NoLabel to be backfilled to NO, got %q", out.NoLabel)
	}
}

func TestMigrateAddMarketLabels_IsIdempotent(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	// Ensure columns exist first time
	if err := migrations.MigrateAddMarketLabels(db); err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	// Run a second time to confirm idempotence (no errors)
	if err := migrations.MigrateAddMarketLabels(db); err != nil {
		t.Fatalf("second run failed (should be idempotent): %v", err)
	}

	// Quick sanity: columns still there
	if !db.Migrator().HasColumn(&models.Market{}, "YesLabel") ||
		!db.Migrator().HasColumn(&models.Market{}, "NoLabel") {
		t.Fatalf("expected columns to remain after second run")
	}

	_ = time.Now() // keep lints happy if needed
}
