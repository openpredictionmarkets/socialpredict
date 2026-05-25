package migrations_test

import (
	"testing"
	"time"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

type MarketWithoutLifecycleStatus struct {
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
	YesLabel                string
	NoLabel                 string
	CreatorUsername         string
}

func (MarketWithoutLifecycleStatus) TableName() string { return "markets" }

func TestMigrateAddMarketLifecycleStatusAddsColumnAndBackfills(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_ = db.Migrator().DropColumn(&models.Market{}, "LifecycleStatus")

	open := MarketWithoutLifecycleStatus{
		ID:                 11,
		QuestionTitle:      "Open",
		Description:        "Open market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: time.Now().Add(24 * time.Hour),
		InitialProbability: 0.5,
		CreatorUsername:    "creator",
	}
	resolved := open
	resolved.ID = 12
	resolved.QuestionTitle = "Resolved"
	resolved.IsResolved = true
	resolved.ResolutionResult = "YES"

	if err := db.Create(&open).Error; err != nil {
		t.Fatalf("seed open market: %v", err)
	}
	if err := db.Create(&resolved).Error; err != nil {
		t.Fatalf("seed resolved market: %v", err)
	}

	if err := migrations.MigrateAddMarketLifecycleStatus(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasColumn(&models.Market{}, "LifecycleStatus") {
		t.Fatalf("expected lifecycle_status column after migration")
	}

	var out []models.Market
	if err := db.Order("id").Find(&out).Error; err != nil {
		t.Fatalf("load migrated markets: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 markets, got %d", len(out))
	}
	if out[0].LifecycleStatus != "published" {
		t.Fatalf("open lifecycle_status = %q, want published", out[0].LifecycleStatus)
	}
	if out[1].LifecycleStatus != "resolved" {
		t.Fatalf("resolved lifecycle_status = %q, want resolved", out[1].LifecycleStatus)
	}
}
