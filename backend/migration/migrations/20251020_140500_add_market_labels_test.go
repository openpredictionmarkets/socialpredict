package migrations_test

import (
	"testing"
	"time"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

// MarketV1 mirrors the old schema (no label fields) but uses same table name.
type MarketV1 struct {
	ID                      int64 `gorm:"primaryKey"`
	QuestionTitle           string
	Description             string
	OutcomeType             string
	ResolutionDateTime      time.Time
	FinalResolutionDateTime time.Time
	UTCOffset               int
	IsResolved              bool
	ResolutionResult        string
	InitialProbability      float64
	CreatorUsername         string
}

func (MarketV1) TableName() string { return "markets" }

func seedPreMigrationMarket(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	m := MarketV1{
		ID:                 1,
		QuestionTitle:      "Test Market",
		Description:        "Test Description",
		OutcomeType:        "BINARY",
		ResolutionDateTime: time.Now().Add(24 * time.Hour),
		CreatorUsername:    "alice",
	}
	if err := db.Create(&m).Error; err != nil {
		t.Fatalf("failed to seed v1 market: %v", err)
	}
	return m.ID
}

func TestMigrateAddMarketLabels_AddsColumnsAndBackfills(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	// Simulate v1: drop the new columns if present.
	_ = db.Migrator().DropColumn(&models.Market{}, "YesLabel")
	_ = db.Migrator().DropColumn(&models.Market{}, "NoLabel")

	id := seedPreMigrationMarket(t, db)

	// Run the migration under test.
	if err := migrations.MigrateAddMarketLabels(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Assert columns exist.
	mig := db.Migrator()
	if !mig.HasColumn(&models.Market{}, "YesLabel") {
		t.Fatalf("expected yes_label column after migration")
	}
	if !mig.HasColumn(&models.Market{}, "NoLabel") {
		t.Fatalf("expected no_label column after migration")
	}

	// Assert backfill applied.
	var out models.Market
	if err := db.First(&out, id).Error; err != nil {
		t.Fatalf("load market failed: %v", err)
	}
	if out.YesLabel != "YES" {
		t.Fatalf("expected YesLabel 'YES', got %q", out.YesLabel)
	}
	if out.NoLabel != "NO" {
		t.Fatalf("expected NoLabel 'NO', got %q", out.NoLabel)
	}
}
