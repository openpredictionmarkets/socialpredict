package socialshare

import (
	"errors"
	"testing"

	"socialpredict/models"

	"gorm.io/gorm"
)

var tinyPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
	0x42, 0x60, 0x82,
}

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
	svc := NewService(&mockRepository{}, NewFileImageStore(t.TempDir()))

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
	svc := NewService(repo, NewFileImageStore(t.TempDir()))

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

func TestUploadImageStoresImageAndPointsSettingsAtImageRoute(t *testing.T) {
	repo := &mockRepository{}
	store := NewFileImageStore(t.TempDir())
	svc := NewService(repo, store)

	settings, err := svc.UploadImage(UploadImageInput{
		FileName:  "card.png",
		Data:      tinyPNG,
		ImageAlt:  "Uploaded share card",
		UpdatedBy: "admin",
	})
	if err != nil {
		t.Fatalf("UploadImage returned error: %v", err)
	}
	image, err := store.Load()
	if err != nil {
		t.Fatalf("load stored image: %v", err)
	}
	if image.ContentType != "image/png" {
		t.Fatalf("ContentType = %q", image.ContentType)
	}
	if settings.DefaultImageURL != UploadedImageURL {
		t.Fatalf("DefaultImageURL = %q", settings.DefaultImageURL)
	}
	if settings.ImageAlt != "Uploaded share card" {
		t.Fatalf("ImageAlt = %q", settings.ImageAlt)
	}
}

func TestUploadImageRejectsUnsupportedContentType(t *testing.T) {
	svc := NewService(&mockRepository{}, NewFileImageStore(t.TempDir()))

	_, err := svc.UploadImage(UploadImageInput{Data: []byte("not an image")})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestUpdateSettingsRejectsUnsafeImageURL(t *testing.T) {
	svc := NewService(&mockRepository{}, NewFileImageStore(t.TempDir()))

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
