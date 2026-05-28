package http

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialpredict/handlers"
	"socialpredict/handlers/cms/socialshare"
	"socialpredict/internal/domain/users"
	rusers "socialpredict/internal/repository/users"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/security"

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
	0x00, 0x00, 0x00, 0x49, 0x45, 0x44, 0xae,
	0x42, 0x60, 0x82,
}

func TestPublicGetReturnsDefaultSocialShareSettings(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	handler := newTestHandler(t, db)

	req := httptest.NewRequest(http.MethodGet, "/v0/content/social-share", nil)
	rec := httptest.NewRecorder()

	handler.PublicGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp handlers.SuccessEnvelope[map[string]interface{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Result["siteName"] != socialshare.DefaultSiteName {
		t.Fatalf("expected default siteName, got %#v", resp.Result["siteName"])
	}
}

func TestAdminUpdateSocialShareSettings(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	admin := seedAdmin(t, db)
	handler := newTestHandler(t, db)

	payload := updateReq{
		SiteName:           "KConfs",
		DefaultDescription: "Share SocialPredict markets with a complete public preview card.",
		DefaultImageURL:    "https://kconfs.com/og/socialpredict-card.png",
		ImageAlt:           "SocialPredict market share card",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/v0/admin/content/social-share", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT(admin.Username))
	rec := httptest.NewRecorder()

	handler.AdminUpdate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var stored models.SocialShareSettings
	if err := db.Where("slug = ?", "default").First(&stored).Error; err != nil {
		t.Fatalf("load stored social share settings: %v", err)
	}
	if stored.SiteName != payload.SiteName || stored.UpdatedBy != admin.Username {
		t.Fatalf("unexpected stored settings: %+v", stored)
	}
}

func TestAdminUploadImageStoresPublicImageAndUpdatesSettings(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	admin := seedAdmin(t, db)
	handler := newTestHandler(t, db)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("imageAlt", "Uploaded OpenGraph image")
	part, err := writer.CreateFormFile("image", "card.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	_, _ = part.Write(tinyPNG)
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v0/admin/content/social-share/image", &body)
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT(admin.Username))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	handler.AdminUploadImage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var stored models.SocialShareSettings
	if err := db.Where("slug = ?", "default").First(&stored).Error; err != nil {
		t.Fatalf("load stored settings: %v", err)
	}
	if stored.DefaultImageURL != socialshare.UploadedImageURL {
		t.Fatalf("DefaultImageURL = %q", stored.DefaultImageURL)
	}

	imageReq := httptest.NewRequest(http.MethodGet, "/v0/content/social-share/image", nil)
	imageRec := httptest.NewRecorder()
	handler.PublicImage(imageRec, imageReq)
	if imageRec.Code != http.StatusOK {
		t.Fatalf("expected image status 200, got %d: %s", imageRec.Code, imageRec.Body.String())
	}
	if got := imageRec.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("Content-Type = %q", got)
	}
	if !bytes.Equal(imageRec.Body.Bytes(), tinyPNG) {
		t.Fatalf("served image bytes do not match upload")
	}
}

func newTestHandler(t *testing.T, db *gorm.DB) *Handler {
	t.Helper()
	svc := socialshare.NewService(socialshare.NewGormRepository(db), socialshare.NewFileImageStore(t.TempDir()))
	auth := authsvc.NewAuthService(users.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer))
	return NewHandler(svc, auth)
}

func seedAdmin(t *testing.T, db *gorm.DB) models.User {
	t.Helper()
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	admin := modelstesting.GenerateUser("admin_user", 0)
	admin.UserType = "ADMIN"
	admin.MustChangePassword = false
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("create admin user: %v", err)
	}
	if err := db.Model(&admin).Update("must_change_password", false).Error; err != nil {
		t.Fatalf("clear must_change_password: %v", err)
	}
	return admin
}
