package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"socialpredict/handlers"
	appruntime "socialpredict/internal/app/runtime"
	authsvc "socialpredict/internal/service/auth"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/logger"
	"socialpredict/models/modelstesting"
	"socialpredict/security"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var testOpenAPISpec = []byte("openapi: 3.0.0\ninfo:\n  title: SocialPredict Test API\n")

type recordingShutdowner struct {
	t         *testing.T
	readiness *appruntime.Readiness
	deadline  time.Time
}

func (s *recordingShutdowner) Shutdown(ctx context.Context) error {
	s.t.Helper()

	if s.readiness.Ready() {
		s.t.Fatalf("expected readiness gate closed before server shutdown")
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		s.t.Fatalf("expected shutdown context deadline")
	}
	s.deadline = deadline
	return nil
}

func testSwaggerUIFS() fstest.MapFS {
	return fstest.MapFS{
		"swagger-ui/index.html": &fstest.MapFile{
			Data: []byte("<html>swagger</html>"),
		},
	}
}

func TestShutdownHTTPServerClosesReadinessBeforeDrain(t *testing.T) {
	readiness := appruntime.NewReadiness()
	readiness.MarkReady()
	shutdowner := &recordingShutdowner{t: t, readiness: readiness}
	var slept time.Duration

	start := time.Now()
	err := shutdownHTTPServer(
		shutdowner,
		readiness,
		appruntime.ShutdownConfig{
			ReadinessDrainWindow: 7 * time.Second,
			ShutdownTimeout:      15 * time.Second,
		},
		func(duration time.Duration) {
			if readiness.Ready() {
				t.Fatalf("expected readiness gate closed before drain wait")
			}
			slept = duration
		},
	)
	if err != nil {
		t.Fatalf("shutdownHTTPServer: %v", err)
	}

	if readiness.Ready() {
		t.Fatalf("expected readiness gate to remain closed after shutdown")
	}
	if slept != 7*time.Second {
		t.Fatalf("readiness drain sleep = %v, want 7s", slept)
	}
	if shutdowner.deadline.Before(start.Add(14*time.Second)) || shutdowner.deadline.After(start.Add(16*time.Second)) {
		t.Fatalf("shutdown deadline = %v, want about 15s after %v", shutdowner.deadline, start)
	}
}

func buildTestRouter(t *testing.T, db *gorm.DB) *mux.Router {
	t.Helper()

	econConfig := modelstesting.GenerateEconomicConfig()
	readiness := appruntime.NewReadiness()
	readiness.MarkReady()

	router, err := buildRouter(testOpenAPISpec, testSwaggerUIFS(), db, configsvc.NewStaticService(econConfig), readiness, testSecurityConfig(t))
	if err != nil {
		t.Fatalf("build test router: %v", err)
	}
	return router
}

func buildTestHandler(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()

	econConfig := modelstesting.GenerateEconomicConfig()
	readiness := appruntime.NewReadiness()
	readiness.MarkReady()
	return buildTestHandlerWithConfig(t, db, econConfig, readiness)
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
		{"readyz", "/readyz", http.StatusOK},
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

	readiness := appruntime.NewReadiness()
	readiness.MarkReady()
	handler := buildTestHandlerWithConfig(t, db, config, readiness)

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

func TestBuildHandlerPropagatesRequestIDToPublicRoutes(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	db := modelstesting.NewFakeDB(t)
	handler := buildTestHandler(t, db)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-Id", "external-request-id")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("X-Request-Id"); got != "external-request-id" {
		t.Fatalf("expected propagated request ID header, got %q", got)
	}
}

func TestBuildHandlerUsesSharedMethodNotAllowedWriter(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	db := modelstesting.NewFakeDB(t)
	handler := buildTestHandler(t, db)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
	if got := rec.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("expected Allow GET, got %q", got)
	}

	var response struct {
		OK     bool   `json:"ok"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode method-not-allowed response: %v", err)
	}
	if response.OK || response.Reason != security.RuntimeReasonMethodNotAllowed {
		t.Fatalf("expected method-not-allowed reason, got %+v", response)
	}
}

func TestBuildHandlerFailsClosedWithoutJWTSigningKey(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	readiness := appruntime.NewReadiness()
	readiness.MarkReady()

	_, err := buildHandler(testOpenAPISpec, testSwaggerUIFS(), db, configsvc.NewStaticService(modelstesting.GenerateEconomicConfig()), readiness, appruntime.SecurityConfig{})
	if err == nil {
		t.Fatalf("expected missing JWT signing key error")
	}
}

func TestBuildHandlerUsesRuntimeSecurityPosture(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	db := modelstesting.NewFakeDB(t)
	readiness := appruntime.NewReadiness()
	readiness.MarkReady()
	securityConfig := testSecurityConfig(t)
	securityConfig.CORS.AllowedOrigins = []string{"https://app.example"}
	securityConfig.Headers.StrictTransportSecurity = "max-age=300"

	handler, err := buildHandler(testOpenAPISpec, testSwaggerUIFS(), db, configsvc.NewStaticService(modelstesting.GenerateEconomicConfig()), readiness, securityConfig)
	if err != nil {
		t.Fatalf("build handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "https://app.example")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want runtime origin", got)
	}
	if got := rec.Header().Get("Strict-Transport-Security"); got != "max-age=300" {
		t.Fatalf("Strict-Transport-Security = %q, want runtime HSTS", got)
	}
}

func TestBuildHandlerRequiresRuntimeJWTSigningKey(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	readiness := appruntime.NewReadiness()
	readiness.MarkReady()

	_, err := buildHandler(
		testOpenAPISpec,
		testSwaggerUIFS(),
		db,
		configsvc.NewStaticService(modelstesting.GenerateEconomicConfig()),
		readiness,
		appruntime.SecurityConfig{},
	)
	if err == nil {
		t.Fatalf("expected missing JWT signing key error")
	}
}

func TestBuildHandlerAppliesRuntimeHeaderPosture(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")
	t.Setenv("SECURITY_HSTS_ENABLED", "true")
	t.Setenv("SECURITY_HSTS_MAX_AGE", "99")

	db := modelstesting.NewFakeDB(t)
	handler := buildTestHandler(t, db)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Strict-Transport-Security"); got != "max-age=99" {
		t.Fatalf("expected runtime HSTS header, got %q", got)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected security header posture, got %q", got)
	}
}

func TestInfraHealthAndReadinessRoutesHaveDistinctSemantics(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	db := modelstesting.NewFakeDB(t)
	readiness := appruntime.NewReadiness()
	handler := buildTestHandlerWithConfig(t, db, modelstesting.GenerateEconomicConfig(), readiness)

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRec := httptest.NewRecorder()
	handler.ServeHTTP(healthRec, healthReq)

	if healthRec.Code != http.StatusOK {
		t.Fatalf("expected /health status 200, got %d", healthRec.Code)
	}
	if body := healthRec.Body.String(); body != "live" {
		t.Fatalf("expected /health body live, got %q", body)
	}
	if got := healthRec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected /health cache-control no-store, got %q", got)
	}

	readyReq := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	readyRec := httptest.NewRecorder()
	handler.ServeHTTP(readyRec, readyReq)

	if readyRec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected /readyz status 503, got %d", readyRec.Code)
	}
	if body := readyRec.Body.String(); body != "not ready" {
		t.Fatalf("expected /readyz body not ready, got %q", body)
	}

	readiness.MarkReady()

	readyRec = httptest.NewRecorder()
	handler.ServeHTTP(readyRec, readyReq)

	if readyRec.Code != http.StatusOK {
		t.Fatalf("expected /readyz status 200 after readiness, got %d", readyRec.Code)
	}
	if body := readyRec.Body.String(); body != "ready" {
		t.Fatalf("expected /readyz body ready after readiness, got %q", body)
	}
}

func TestReadyzChecksDatabaseAvailabilityAfterReadinessGateOpens(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	db := modelstesting.NewFakeDB(t)
	readiness := appruntime.NewReadiness()
	readiness.MarkReady()
	handler := buildTestHandlerWithConfig(t, db, modelstesting.GenerateEconomicConfig(), readiness)

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close sql db: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected /readyz status 503 when database is unavailable, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "not ready" {
		t.Fatalf("expected /readyz body not ready when database is unavailable, got %q", body)
	}
}

func TestServerBlocksProtectedProfileRoutesWhenPasswordChangeRequiredWithSharedReason(t *testing.T) {
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

			var envelope handlers.FailureEnvelope
			if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode failure envelope: %v", err)
			}
			if envelope.OK || envelope.Reason != string(handlers.ReasonPasswordChangeRequired) {
				t.Fatalf("expected reason %q, got %+v", handlers.ReasonPasswordChangeRequired, envelope)
			}
		})
	}
}

func TestServerBlocksPrivateActionRoutesWhenPasswordChangeRequiredWithSharedReason(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	db := modelstesting.NewFakeDB(t)
	seedServerTestData(t, db)

	user := modelstesting.GenerateUser("mustchangeactions", 1000)
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
		{name: "place bet", method: http.MethodPost, path: "/v0/bet", body: `{"marketId":1,"amount":10,"outcome":"YES"}`},
		{name: "sell position", method: http.MethodPost, path: "/v0/sell", body: `{"marketId":1,"amount":1,"outcome":"YES"}`},
		{name: "user position", method: http.MethodGet, path: "/v0/userposition/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected status 403, got %d (body: %s)", rec.Code, rec.Body.String())
			}

			var envelope handlers.FailureEnvelope
			if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode failure envelope: %v", err)
			}
			if envelope.OK || envelope.Reason != string(handlers.ReasonPasswordChangeRequired) {
				t.Fatalf("expected reason %q, got %+v", handlers.ReasonPasswordChangeRequired, envelope)
			}
		})
	}
}

func TestServerSetsRequestIDHeaderWhenMissing(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	db := modelstesting.NewFakeDB(t)
	seedServerTestData(t, db)

	handler := buildTestHandler(t, db)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	requestID := rec.Header().Get(logger.RequestIDHeader)
	if requestID == "" {
		t.Fatalf("expected %s response header to be set", logger.RequestIDHeader)
	}
}

func TestServerPreservesIncomingRequestIDHeader(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	db := modelstesting.NewFakeDB(t)
	seedServerTestData(t, db)

	handler := buildTestHandler(t, db)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(logger.RequestIDHeader, "req-test-123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(logger.RequestIDHeader); got != "req-test-123" {
		t.Fatalf("expected preserved request id %q, got %q", "req-test-123", got)
	}
}

func TestRequestLoggingMiddlewareTreatsCanceledRequestsAsClientClosed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v0/markets", nil)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler := logger.RequestLoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected recorder to remain unwritten with default 200, got %d", rec.Code)
	}
}

func buildTestHandlerWithConfig(t *testing.T, db *gorm.DB, econConfig any, readiness ...*appruntime.Readiness) http.Handler {
	t.Helper()

	var gate *appruntime.Readiness
	if len(readiness) > 0 {
		gate = readiness[0]
	} else {
		gate = appruntime.NewReadiness()
		gate.MarkReady()
	}

	handler, err := buildHandler(testOpenAPISpec, testSwaggerUIFS(), db, configsvc.NewStaticService(econConfig), gate, testSecurityConfig(t))
	if err != nil {
		t.Fatalf("build test handler: %v", err)
	}
	return handler
}

func testSecurityConfig(t *testing.T) appruntime.SecurityConfig {
	t.Helper()

	config, err := appruntime.LoadSecurityConfigFromEnv()
	if err != nil {
		t.Fatalf("load test security config: %v", err)
	}
	authsvc.ConfigureJWTSigningKey(config.JWTSigningKey)
	return config
}
