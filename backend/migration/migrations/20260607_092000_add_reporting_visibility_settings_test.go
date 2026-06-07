package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddReportingVisibilitySettingsCreatesDefaultPublicToggles(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddReportingVisibilitySettings(db); err != nil {
		t.Fatalf("migration returned error: %v", err)
	}

	var settings models.ReportingVisibilitySettings
	if err := db.Where("slug = ?", "default").First(&settings).Error; err != nil {
		t.Fatalf("load default settings: %v", err)
	}
	if !settings.SystemMetricsPublic {
		t.Fatalf("system metrics default should preserve public visibility")
	}
	if !settings.GlobalLeaderboardPublic {
		t.Fatalf("global leaderboard default should preserve public visibility")
	}
	if settings.Version != 1 {
		t.Fatalf("version = %d, want 1", settings.Version)
	}
}
