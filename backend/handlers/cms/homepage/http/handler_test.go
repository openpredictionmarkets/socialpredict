package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialpredict/handlers"
	"socialpredict/handlers/cms/homepage"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	dusers "socialpredict/internal/domain/users"
	rusers "socialpredict/internal/repository/users"
	"socialpredict/security"
)

func TestPublicGet_ReturnsHomepageContent(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	item := models.HomepageContent{
		Slug:     "home",
		Title:    "Welcome",
		Format:   "html",
		Markdown: "",
		HTML:     "<p>Hello</p>",
		Version:  1,
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed homepage content: %v", err)
	}

	repo := homepage.NewGormRepository(db)
	renderer := homepage.NewDefaultRenderer()
	svc := homepage.NewService(repo, renderer)
	usersSvc := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)
	auth := authsvc.NewAuthService(usersSvc)
	handler := NewHandler(svc, auth)

	req := httptest.NewRequest("GET", "/v0/content/home", nil)
	rec := httptest.NewRecorder()

	handler.PublicGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp handlers.SuccessEnvelope[map[string]interface{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("expected ok=true, got false")
	}
	if resp.Result["title"] != item.Title {
		t.Fatalf("expected title %q, got %q", item.Title, resp.Result["title"])
	}
}

func TestAdminUpdate_Success(t *testing.T) {
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

	item := models.HomepageContent{
		Slug:    "home",
		Title:   "Old title",
		Format:  "html",
		HTML:    "<p>Old</p>",
		Version: 1,
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed homepage content: %v", err)
	}

	repo := homepage.NewGormRepository(db)
	renderer := homepage.NewDefaultRenderer()
	svc := homepage.NewService(repo, renderer)
	usersSvc := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)
	auth := authsvc.NewAuthService(usersSvc)
	handler := NewHandler(svc, auth)

	payload := updateReq{
		Title:   "New title",
		Format:  "html",
		HTML:    "<p>New</p>",
		Version: item.Version,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/v0/admin/content/home", bytes.NewReader(body))
	token := modelstesting.GenerateValidJWT(admin.Username)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.AdminUpdate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp handlers.SuccessEnvelope[map[string]interface{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("expected ok=true, got false")
	}
	if resp.Result["title"] != payload.Title {
		t.Fatalf("expected updated title %q, got %q", payload.Title, resp.Result["title"])
	}
	if resp.Result["markdown"] != "" {
		t.Fatalf("expected cleared markdown, got %#v", resp.Result["markdown"])
	}
	if resp.Result["updatedAt"] == nil {
		t.Fatalf("expected updatedAt in response")
	}

	var stored models.HomepageContent
	if err := db.Where("slug = ?", "home").First(&stored).Error; err != nil {
		t.Fatalf("fetch stored content: %v", err)
	}
	if stored.Title != payload.Title {
		t.Fatalf("expected stored title %q, got %q", payload.Title, stored.Title)
	}
	if stored.UpdatedBy != admin.Username {
		t.Fatalf("expected UpdatedBy %q, got %q", admin.Username, stored.UpdatedBy)
	}
}

func TestAdminUpdate_Unauthorized(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	repo := homepage.NewGormRepository(db)
	renderer := homepage.NewDefaultRenderer()
	svc := homepage.NewService(repo, renderer)
	usersSvc := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)
	auth := authsvc.NewAuthService(usersSvc)
	handler := NewHandler(svc, auth)

	req := httptest.NewRequest("PUT", "/v0/admin/content/home", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()

	handler.AdminUpdate(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if resp.Reason != string(handlers.ReasonInvalidToken) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonInvalidToken, resp.Reason)
	}
}

func TestPublicGet_NotFound(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	repo := homepage.NewGormRepository(db)
	renderer := homepage.NewDefaultRenderer()
	svc := homepage.NewService(repo, renderer)
	usersSvc := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)
	auth := authsvc.NewAuthService(usersSvc)
	handler := NewHandler(svc, auth)

	req := httptest.NewRequest(http.MethodGet, "/v0/content/home", nil)
	rec := httptest.NewRecorder()

	handler.PublicGet(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if resp.Reason != string(handlers.ReasonNotFound) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonNotFound, resp.Reason)
	}
}

func TestAdminUpdate_NonAdminReturnsAuthorizationDenied(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	user := modelstesting.GenerateUser("member_user", 0)
	user.UserType = "USER"
	user.MustChangePassword = false
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Model(&user).Update("must_change_password", false).Error; err != nil {
		t.Fatalf("clear must_change_password: %v", err)
	}

	item := models.HomepageContent{
		Slug:    "home",
		Title:   "Old title",
		Format:  "html",
		HTML:    "<p>Old</p>",
		Version: 1,
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed homepage content: %v", err)
	}

	repo := homepage.NewGormRepository(db)
	renderer := homepage.NewDefaultRenderer()
	svc := homepage.NewService(repo, renderer)
	usersSvc := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)
	auth := authsvc.NewAuthService(usersSvc)
	handler := NewHandler(svc, auth)

	req := httptest.NewRequest("PUT", "/v0/admin/content/home", bytes.NewReader([]byte(`{"title":"Nope","format":"html","html":"<p>Nope</p>","version":1}`)))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT(user.Username))
	rec := httptest.NewRecorder()

	handler.AdminUpdate(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}

	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if resp.Reason != string(handlers.ReasonAuthorizationDenied) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonAuthorizationDenied, resp.Reason)
	}
}
