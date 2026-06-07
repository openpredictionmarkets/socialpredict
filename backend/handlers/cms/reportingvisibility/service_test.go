package reportingvisibility

import (
	"testing"

	"socialpredict/models"

	"gorm.io/gorm"
)

type memoryRepo struct {
	item *models.ReportingVisibilitySettings
	err  error
}

func (r *memoryRepo) GetSettings() (*models.ReportingVisibilitySettings, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.item == nil {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *r.item
	return &copy, nil
}

func (r *memoryRepo) SaveSettings(item *models.ReportingVisibilitySettings) error {
	copy := *item
	copy.ID = 1
	r.item = &copy
	return nil
}

func TestServiceGetSettingsDefaultsToPublicReporting(t *testing.T) {
	svc := NewService(&memoryRepo{})

	settings, err := svc.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}
	if !settings.SystemMetricsPublic || !settings.GlobalLeaderboardPublic {
		t.Fatalf("expected default public reporting settings, got %+v", settings)
	}
}

func TestServiceUpdateSettingsPersistsAdminToggles(t *testing.T) {
	repo := &memoryRepo{item: DefaultSettings()}
	repo.item.ID = 1
	svc := NewService(repo)

	hideSystem := false
	showLeaderboard := true
	settings, err := svc.UpdateSettings(UpdateInput{
		SystemMetricsPublic:     &hideSystem,
		GlobalLeaderboardPublic: &showLeaderboard,
		Version:                 1,
		UpdatedBy:               "admin",
	})
	if err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}
	if settings.SystemMetricsPublic {
		t.Fatalf("expected system metrics public=false")
	}
	if !settings.GlobalLeaderboardPublic {
		t.Fatalf("expected global leaderboard public=true")
	}
	if settings.Version != 2 {
		t.Fatalf("version = %d, want 2", settings.Version)
	}
	if settings.UpdatedBy != "admin" {
		t.Fatalf("updated by = %q, want admin", settings.UpdatedBy)
	}
}
