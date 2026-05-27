package http

import (
	"bytes"
	"encoding/json"
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
)

func TestPublicGetReturnsDefaultSocialShareSettings(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	svc := socialshare.NewService(socialshare.NewGormRepository(db))
	auth := authsvc.NewAuthService(users.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer))
	handler := NewHandler(svc, auth)

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

	svc := socialshare.NewService(socialshare.NewGormRepository(db))
	auth := authsvc.NewAuthService(users.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer))
	handler := NewHandler(svc, auth)

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
