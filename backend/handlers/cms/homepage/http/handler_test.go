package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialpredict/handlers/cms/homepage"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
)

func TestPublicGet_ReturnsHomepageContent(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	origDB := util.DB
	util.DB = db
	t.Cleanup(func() {
		util.DB = origDB
	})

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
	handler := NewHandler(svc)

	req := httptest.NewRequest("GET", "/v0/content/home", nil)
	rec := httptest.NewRecorder()

	handler.PublicGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp["title"] != item.Title {
		t.Fatalf("expected title %q, got %q", item.Title, resp["title"])
	}
}

func TestAdminUpdate_Success(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	origDB := util.DB
	util.DB = db
	t.Cleanup(func() {
		util.DB = origDB
	})
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	admin := modelstesting.GenerateUser("admin_user", 0)
	admin.UserType = "ADMIN"
	admin.MustChangePassword = false
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("create admin user: %v", err)
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
	handler := NewHandler(svc)

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

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp["title"] != payload.Title {
		t.Fatalf("expected updated title %q, got %q", payload.Title, resp["title"])
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
	origDB := util.DB
	util.DB = db
	t.Cleanup(func() {
		util.DB = origDB
	})
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	repo := homepage.NewGormRepository(db)
	renderer := homepage.NewDefaultRenderer()
	svc := homepage.NewService(repo, renderer)
	handler := NewHandler(svc)

	req := httptest.NewRequest("PUT", "/v0/admin/content/home", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()

	handler.AdminUpdate(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
