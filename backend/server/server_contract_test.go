package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"socialpredict/handlers"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestServerServesOpenAPIDocumentAndSwaggerRedirect(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	db := modelstesting.NewFakeDB(t)
	handler := buildTestHandler(t, db)

	openAPIReq := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	openAPIRec := httptest.NewRecorder()
	handler.ServeHTTP(openAPIRec, openAPIReq)

	if openAPIRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", openAPIRec.Code, openAPIRec.Body.String())
	}
	if got := openAPIRec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/yaml") {
		t.Fatalf("expected yaml content type, got %q", got)
	}
	if !strings.Contains(openAPIRec.Body.String(), "openapi:") {
		t.Fatalf("expected OpenAPI body, got %q", openAPIRec.Body.String())
	}

	swaggerRedirectReq := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	swaggerRedirectRec := httptest.NewRecorder()
	handler.ServeHTTP(swaggerRedirectRec, swaggerRedirectReq)

	if swaggerRedirectRec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected swagger redirect status 301, got %d", swaggerRedirectRec.Code)
	}
	if location := swaggerRedirectRec.Header().Get("Location"); location != "/swagger/" {
		t.Fatalf("expected redirect to /swagger/, got %q", location)
	}

	swaggerReq := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	swaggerRec := httptest.NewRecorder()
	handler.ServeHTTP(swaggerRec, swaggerReq)

	if swaggerRec.Code != http.StatusOK {
		t.Fatalf("expected swagger status 200, got %d", swaggerRec.Code)
	}
	if !strings.Contains(swaggerRec.Body.String(), "swagger") {
		t.Fatalf("expected swagger placeholder body, got %q", swaggerRec.Body.String())
	}
}

func TestServerAllowsChangePasswordWhenPasswordChangeRequired(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	db := modelstesting.NewFakeDB(t)
	user := modelstesting.GenerateUser("mustchange", 1000)
	if err := user.HashPassword("OldPassword123"); err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user.MustChangePassword = true
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	handler := buildTestHandler(t, db)
	req := httptest.NewRequest(http.MethodPost, "/v0/changepassword", bytes.NewBufferString(`{"currentPassword":"OldPassword123","newPassword":"NewPassword123"}`))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT(user.Username))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response struct {
		OK     bool `json:"ok"`
		Result struct {
			Message string `json:"message"`
		} `json:"result"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode change-password response: %v", err)
	}
	if !response.OK || response.Result.Message != "Password changed successfully" {
		t.Fatalf("unexpected response %+v", response)
	}
}

func TestServerServesPublicReportingAndContentRoutesWithoutAuth(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	db := modelstesting.NewFakeDB(t)
	seedServerTestData(t, db)

	content := models.HomepageContent{
		Slug:    "home",
		Title:   "Welcome",
		Format:  "html",
		HTML:    "<p>Hello</p>",
		Version: 1,
	}
	if err := db.Create(&content).Error; err != nil {
		t.Fatalf("seed homepage content: %v", err)
	}

	handler := buildTestHandler(t, db)

	tests := []string{
		"/v0/home",
		"/v0/stats",
		"/v0/system/metrics",
		"/v0/global/leaderboard",
		"/v0/content/home",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
			}

			var response map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			ok, present := response["ok"].(bool)
			if !present || !ok {
				t.Fatalf("expected shared success envelope, got %+v", response)
			}
		})
	}
}

func TestSystemMetricsRouteRemainsApplicationReporting(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	db := modelstesting.NewFakeDB(t)
	seedServerTestData(t, db)

	handler := buildTestHandler(t, db)

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRec := httptest.NewRecorder()
	handler.ServeHTTP(healthRec, healthReq)

	if healthRec.Code != http.StatusOK {
		t.Fatalf("expected /health status 200, got %d", healthRec.Code)
	}
	if got := healthRec.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("expected /health plain-text probe response, got %q", got)
	}
	if body := healthRec.Body.String(); body != "live" {
		t.Fatalf("expected /health body live, got %q", body)
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/v0/system/metrics", nil)
	metricsRec := httptest.NewRecorder()
	handler.ServeHTTP(metricsRec, metricsReq)

	if metricsRec.Code != http.StatusOK {
		t.Fatalf("expected /v0/system/metrics status 200, got %d: %s", metricsRec.Code, metricsRec.Body.String())
	}
	if got := metricsRec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected /v0/system/metrics JSON content type, got %q", got)
	}

	var response handlers.SuccessEnvelope[map[string]any]
	if err := json.Unmarshal(metricsRec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode /v0/system/metrics response: %v", err)
	}
	if !response.OK {
		t.Fatalf("expected success envelope from /v0/system/metrics, got %+v", response)
	}
	if _, ok := response.Result["moneyCreated"]; !ok {
		t.Fatalf("expected moneyCreated metrics payload, got %+v", response.Result)
	}
}
