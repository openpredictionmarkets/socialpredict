package migrations_test

import (
	"testing"
	"time"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

type MarketBeforeStewardship struct {
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
	YesLabel                string
	NoLabel                 string
	LifecycleStatus         string
	ProposalCost            int64
}

func (MarketBeforeStewardship) TableName() string { return "markets" }

func TestMigrateAddMarketStewardshipAddsColumnBackfillsAndCreatesAuditTable(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_ = db.Migrator().DropTable(&models.MarketStewardshipAudit{})
	_ = db.Migrator().DropColumn(&models.Market{}, "StewardUsername")

	seed := MarketBeforeStewardship{
		ID:                 77,
		QuestionTitle:      "Needs steward",
		Description:        "Backfill stewardship",
		OutcomeType:        "BINARY",
		ResolutionDateTime: time.Now().Add(24 * time.Hour),
		InitialProbability: 0.5,
		CreatorUsername:    "creator",
		LifecycleStatus:    "published",
	}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("seed old market: %v", err)
	}

	if err := migrations.MigrateAddMarketStewardship(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasColumn(&models.Market{}, "StewardUsername") {
		t.Fatalf("expected steward_username column")
	}
	if !db.Migrator().HasTable(&models.MarketStewardshipAudit{}) {
		t.Fatalf("expected market stewardship audit table")
	}

	var out models.Market
	if err := db.First(&out, seed.ID).Error; err != nil {
		t.Fatalf("load migrated market: %v", err)
	}
	if out.StewardUsername != seed.CreatorUsername {
		t.Fatalf("steward backfill = %q, want %q", out.StewardUsername, seed.CreatorUsername)
	}
}

func TestMigrateAddMarketStewardshipIsIdempotent(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	if err := migrations.MigrateAddMarketStewardship(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddMarketStewardship(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
