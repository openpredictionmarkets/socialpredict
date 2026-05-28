package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

type socialShareSettingsV1 struct {
	gorm.Model
	Slug               string
	SiteName           string
	DefaultDescription string
	DefaultImageURL    string
	ImageAlt           string
	Version            uint
	UpdatedBy          string
}

func (socialShareSettingsV1) TableName() string { return "social_share_settings" }

func TestMigrateAddSocialShareImageEnabledAddsColumnAndBackfillsEnabled(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := db.AutoMigrate(&socialShareSettingsV1{}); err != nil {
		t.Fatalf("create v1 table: %v", err)
	}
	if err := db.Create(&socialShareSettingsV1{
		Slug:               "default",
		SiteName:           "SocialPredict",
		DefaultDescription: "Prediction markets for the social web",
		DefaultImageURL:    "/og/socialpredict-share.png",
		ImageAlt:           "SocialPredict share card",
		Version:            1,
	}).Error; err != nil {
		t.Fatalf("seed v1 settings: %v", err)
	}
	if db.Migrator().HasColumn(&models.SocialShareSettings{}, "ImageEnabled") {
		t.Fatalf("expected v1 table to omit ImageEnabled before migration")
	}

	if err := migrations.MigrateAddSocialShareImageEnabled(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasColumn(&models.SocialShareSettings{}, "ImageEnabled") {
		t.Fatalf("expected ImageEnabled column after migration")
	}
	var out models.SocialShareSettings
	if err := db.Where("slug = ?", "default").First(&out).Error; err != nil {
		t.Fatalf("load migrated settings: %v", err)
	}
	if !out.ImageEnabled {
		t.Fatalf("expected existing settings to be backfilled with ImageEnabled=true")
	}
}
