package socialshare

import (
	"errors"
	"testing"

	"socialpredict/models"

	"gorm.io/gorm"
)

type mockRepository struct {
	item    *models.SocialShareSettings
	saveErr error
	getErr  error
}

func (m *mockRepository) GetBySlug(slug string) (*models.SocialShareSettings, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.item == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return m.item, nil
}

func (m *mockRepository) Save(item *models.SocialShareSettings) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.item = item
	return nil
}

func TestGetSettingsReturnsDefaultsWhenMissing(t *testing.T) {
	svc := NewService(&mockRepository{})

	settings, err := svc.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}
	if settings.SiteName != DefaultSiteName {
		t.Fatalf("SiteName = %q", settings.SiteName)
	}
	if settings.DefaultImageURL != DefaultImageURL {
		t.Fatalf("DefaultImageURL = %q", settings.DefaultImageURL)
	}
}

func TestUpdateSettingsCreatesDefaultRowAndValidatesURL(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)

	settings, err := svc.UpdateSettings(UpdateInput{
		SiteName:           "KConfs",
		DefaultDescription: "A useful public share description for SocialPredict markets.",
		DefaultImageURL:    "/assets/share-card.png",
		ImageAlt:           "KConfs share card",
		UpdatedBy:          "admin",
	})
	if err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}
	if settings.Slug != settingsSlug {
		t.Fatalf("Slug = %q", settings.Slug)
	}
	if settings.Version != 1 {
		t.Fatalf("Version = %d, want 1", settings.Version)
	}
	if repo.item == nil || repo.item.UpdatedBy != "admin" {
		t.Fatalf("expected saved row with UpdatedBy admin, got %+v", repo.item)
	}
}

func TestUpdateSettingsRejectsUnsafeImageURL(t *testing.T) {
	svc := NewService(&mockRepository{})

	_, err := svc.UpdateSettings(UpdateInput{
		SiteName:           "SocialPredict",
		DefaultDescription: "A useful public share description.",
		DefaultImageURL:    "javascript:alert(1)",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestUpdateSettingsPropagatesRepositoryErrors(t *testing.T) {
	wantErr := errors.New("database unavailable")
	svc := NewService(&mockRepository{getErr: wantErr})

	_, err := svc.UpdateSettings(UpdateInput{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}
