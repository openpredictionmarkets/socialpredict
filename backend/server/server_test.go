package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	configsvc "socialpredict/internal/service/config"
	"socialpredict/models/modelstesting"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var testOpenAPISpec = []byte("openapi: 3.0.0\ninfo:\n  title: SocialPredict Test API\n")

func testSwaggerUIFS() fstest.MapFS {
	return fstest.MapFS{
		"swagger-ui/index.html": &fstest.MapFile{
			Data: []byte("<html>swagger</html>"),
		},
	}
}

func buildTestRouter(t *testing.T, db *gorm.DB) *mux.Router {
	t.Helper()

	econConfig := modelstesting.GenerateEconomicConfig()
	router, err := buildRouter(testOpenAPISpec, testSwaggerUIFS(), db, configsvc.NewStaticService(econConfig))
	if err != nil {
		t.Fatalf("build test router: %v", err)
	}
	return router
}

func buildTestHandler(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()

	econConfig := modelstesting.GenerateEconomicConfig()
	return buildTestHandlerWithConfig(t, db, econConfig)
}

func seedServerTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	creator := modelstesting.GenerateUser("creator", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	market := modelstesting.GenerateMarket(1, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}
}

func TestServerRegistersAndServesCoreRoutes(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	db := modelstesting.NewFakeDB(t)
	seedServerTestData(t, db)

	handler := buildTestHandler(t, db)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"health", "/health", http.StatusOK},
		{"home", "/v0/home", http.StatusOK},
		{"setup frontend", "/v0/setup/frontend", http.StatusOK},
		{"markets", "/v0/markets?status=ACTIVE", http.StatusOK},
		{"userinfo", "/v0/userinfo/creator", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d (body: %s)", tt.wantStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestServerSetupRoutesUseInjectedConfigService(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	db := modelstesting.NewFakeDB(t)
	config := modelstesting.GenerateEconomicConfig()
	config.Economics.MarketIncentives.CreateMarketCost = 77
	config.Frontend.Charts.SigFigs = 1

	handler := buildTestHandlerWithConfig(t, db, config)

	setupReq := httptest.NewRequest(http.MethodGet, "/v0/setup", nil)
	setupRec := httptest.NewRecorder()
	handler.ServeHTTP(setupRec, setupReq)
	if setupRec.Code != http.StatusOK {
		t.Fatalf("expected setup status 200, got %d with body %s", setupRec.Code, setupRec.Body.String())
	}

	var economics struct {
		MarketIncentives struct {
			CreateMarketCost int64 `json:"CreateMarketCost"`
		} `json:"MarketIncentives"`
	}
	if err := json.Unmarshal(setupRec.Body.Bytes(), &economics); err != nil {
		t.Fatalf("decode /v0/setup response: %v", err)
	}
	if economics.MarketIncentives.CreateMarketCost != 77 {
		t.Fatalf("expected injected setup createMarketCost 77, got %d", economics.MarketIncentives.CreateMarketCost)
	}

	frontendReq := httptest.NewRequest(http.MethodGet, "/v0/setup/frontend", nil)
	frontendRec := httptest.NewRecorder()
	handler.ServeHTTP(frontendRec, frontendReq)
	if frontendRec.Code != http.StatusOK {
		t.Fatalf("expected frontend setup status 200, got %d with body %s", frontendRec.Code, frontendRec.Body.String())
	}

	var frontend struct {
		Charts struct {
			SigFigs int `json:"sigFigs"`
		} `json:"charts"`
	}
	if err := json.Unmarshal(frontendRec.Body.Bytes(), &frontend); err != nil {
		t.Fatalf("decode /v0/setup/frontend response: %v", err)
	}
	if frontend.Charts.SigFigs != 2 {
		t.Fatalf("expected clamped frontend sig figs 2, got %d", frontend.Charts.SigFigs)
	}
}

func TestServerBlocksProtectedProfileRoutesWhenPasswordChangeRequired(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	db := modelstesting.NewFakeDB(t)
	seedServerTestData(t, db)

	user := modelstesting.GenerateUser("mustchange", 1000)
	user.MustChangePassword = true
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	handler := buildTestHandler(t, db)
	token := modelstesting.GenerateValidJWT(user.Username)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "private profile", method: http.MethodGet, path: "/v0/privateprofile"},
		{name: "change display name", method: http.MethodPost, path: "/v0/profilechange/displayname", body: `{"displayName":"New Name"}`},
		{name: "change emoji", method: http.MethodPost, path: "/v0/profilechange/emoji", body: `{"emoji":":)"}`},
		{name: "change description", method: http.MethodPost, path: "/v0/profilechange/description", body: `{"description":"updated description"}`},
		{name: "change links", method: http.MethodPost, path: "/v0/profilechange/links", body: `{"personalLink1":"https://example.com"}`},
		{name: "user position", method: http.MethodGet, path: "/v0/userposition/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Authorization", "Bearer "+token)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Fatalf("expected status 403, got %d (body: %s)", rr.Code, rr.Body.String())
			}
		})
	}
}

func buildTestHandlerWithConfig(t *testing.T, db *gorm.DB, econConfig any) http.Handler {
	t.Helper()

	handler, err := buildHandler(testOpenAPISpec, testSwaggerUIFS(), db, configsvc.NewStaticService(econConfig))
	if err != nil {
		t.Fatalf("build test handler: %v", err)
	}
	return handler
}
