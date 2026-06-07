package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialpredict/handlers"
	"socialpredict/handlers/cms/reportingvisibility"
	"socialpredict/internal/domain/users"
	rusers "socialpredict/internal/repository/users"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/security"

	"gorm.io/gorm"
)

func TestPublicGetReturnsReportingVisibilitySettings(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	handler := newTestHandler(t, db)

	req := httptest.NewRequest(http.MethodGet, "/v0/content/reporting-visibility", nil)
	rec := httptest.NewRecorder()

	handler.PublicGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp handlers.SuccessEnvelope[settingsResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Result.SystemMetricsPublic || !resp.Result.GlobalLeaderboardPublic {
		t.Fatalf("expected default public toggles, got %+v", resp.Result)
	}
}

func TestAdminUpdateReportingVisibilitySettings(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	admin := seedAdmin(t, db)
	handler := newTestHandler(t, db)

	hide := false
	payload := updateReq{
		SystemMetricsPublic:     &hide,
		GlobalLeaderboardPublic: &hide,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/v0/admin/content/reporting-visibility", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT(admin.Username))
	rec := httptest.NewRecorder()

	handler.AdminUpdate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var stored models.ReportingVisibilitySettings
	if err := db.Where("slug = ?", "default").First(&stored).Error; err != nil {
		t.Fatalf("load stored reporting visibility settings: %v", err)
	}
	if stored.SystemMetricsPublic || stored.GlobalLeaderboardPublic {
		t.Fatalf("expected both toggles false, got %+v", stored)
	}
	if stored.UpdatedBy != admin.Username {
		t.Fatalf("updated by = %q, want %q", stored.UpdatedBy, admin.Username)
	}
}

func newTestHandler(t *testing.T, db *gorm.DB) *Handler {
	t.Helper()
	svc := reportingvisibility.NewService(reportingvisibility.NewGormRepository(db))
	auth := authsvc.NewAuthService(users.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer))
	return NewHandler(svc, auth)
}

func seedAdmin(t *testing.T, db *gorm.DB) models.User {
	t.Helper()
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	admin := modelstesting.GenerateUser("reporting_admin", 0)
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
