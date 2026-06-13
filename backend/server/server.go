package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"socialpredict/handlers"
	adminhandlers "socialpredict/handlers/admin"
	"socialpredict/handlers/authhttp"
	betshandlers "socialpredict/handlers/bets"
	buybetshandlers "socialpredict/handlers/bets/buying"
	sellbetshandlers "socialpredict/handlers/bets/selling"
	"socialpredict/handlers/cms/homepage"
	cmshomehttp "socialpredict/handlers/cms/homepage/http"
	"socialpredict/handlers/cms/marketdiscovery"
	cmsdiscoveryhttp "socialpredict/handlers/cms/marketdiscovery/http"
	"socialpredict/handlers/cms/reportingvisibility"
	cmsreportinghttp "socialpredict/handlers/cms/reportingvisibility/http"
	"socialpredict/handlers/cms/socialshare"
	cmssocialhttp "socialpredict/handlers/cms/socialshare/http"
	marketshandlers "socialpredict/handlers/markets"
	metricshandlers "socialpredict/handlers/metrics"
	positionshandlers "socialpredict/handlers/positions"
	setuphandlers "socialpredict/handlers/setup"
	statshandlers "socialpredict/handlers/stats"
	usershandlers "socialpredict/handlers/users"
	usercredit "socialpredict/handlers/users/credit"
	privateuser "socialpredict/handlers/users/privateuser"
	publicuser "socialpredict/handlers/users/publicuser"
	"socialpredict/internal/app"
	"socialpredict/internal/app/readmodelinvalidation"
	appruntime "socialpredict/internal/app/runtime"
	dmarkets "socialpredict/internal/domain/markets"
	readmodelrepo "socialpredict/internal/repository/readmodels"
	authsvc "socialpredict/internal/service/auth"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/logger"
	"socialpredict/models"
	"socialpredict/security"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gorm.io/gorm"
)

func buildCORS(config appruntime.CORSConfig) *cors.Cors {
	if !config.Enabled {
		return nil
	}

	return cors.New(cors.Options{
		AllowedOrigins:   config.AllowedOrigins,
		AllowedMethods:   config.AllowedMethods,
		AllowedHeaders:   config.AllowedHeaders,
		ExposedHeaders:   config.ExposedHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	})
}

func buildHandler(openAPISpec []byte, swaggerUIFS fs.FS, db *gorm.DB, configService configsvc.Service, readiness *appruntime.Readiness, securityConfig appruntime.SecurityConfig) (http.Handler, error) {
	operationalMetrics := appruntime.NewOperationalMetrics()
	router, err := buildRouter(openAPISpec, swaggerUIFS, db, configService, readiness, securityConfig, operationalMetrics)
	if err != nil {
		return nil, err
	}

	handler := http.Handler(router)
	if c := buildCORS(securityConfig.CORS); c != nil {
		handler = c.Handler(handler)
	}
	handler = security.SecurityHeadersMiddleware(securityConfig.Headers)(handler)
	// Canonical production request boundary: request IDs, panic recovery, and
	// completion logging live here. Keep logger.RequestLoggingMiddleware out of
	// server wiring to avoid duplicate request IDs and completion logs.
	handler = security.RequestBoundaryMiddlewareWithProxyTrust(securityConfig.TrustProxyHeaders)(handler)
	handler = operationalMetrics.Middleware(handler)

	return handler, nil
}

func buildRouter(openAPISpec []byte, swaggerUIFS fs.FS, db *gorm.DB, configService configsvc.Service, readiness *appruntime.Readiness, securityConfig appruntime.SecurityConfig, operationalMetrics *appruntime.OperationalMetrics) (*mux.Router, error) {
	if configService == nil {
		return nil, fmt.Errorf("config init: configuration service unavailable")
	}
	if len(securityConfig.JWTSigningKey) == 0 {
		return nil, fmt.Errorf("security init: JWT signing key unavailable")
	}

	router := mux.NewRouter()
	router.MethodNotAllowedHandler = methodNotAllowedHandler(router)
	if err := registerInfraRoutes(router, openAPISpec, swaggerUIFS, db, readiness, operationalMetrics); err != nil {
		return nil, err
	}

	registerApplicationRoutes(router, db, configService, securityConfig)
	return router, nil
}

func methodNotAllowedHandler(router *mux.Router) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allow := strings.Join(allowedMethodsForRequest(router, r), ", "); allow != "" {
			w.Header().Set("Allow", allow)
		}
		security.WriteMethodNotAllowed(w)
	})
}

func allowedMethodsForRequest(router *mux.Router, r *http.Request) []string {
	if router == nil || r == nil {
		return nil
	}

	methodSet := make(map[string]struct{})

	_ = router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		methods, err := route.GetMethods()
		if err != nil || len(methods) == 0 {
			return nil
		}

		for _, method := range methods {
			candidate := r.Clone(r.Context())
			candidate.Method = method

			var match mux.RouteMatch
			if route.Match(candidate, &match) {
				for _, matchedMethod := range methods {
					methodSet[matchedMethod] = struct{}{}
				}
				break
			}
		}

		return nil
	})

	if len(methodSet) == 0 {
		return nil
	}

	allowed := make([]string, 0, len(methodSet))
	for method := range methodSet {
		allowed = append(allowed, method)
	}
	sort.Strings(allowed)
	return allowed
}

const readinessProbeTimeout = 2 * time.Second

type applicationReportingService interface {
	metricshandlers.SystemMetricsService
	metricshandlers.GlobalLeaderboardService
}

func registerInfraRoutes(router *mux.Router, openAPISpec []byte, swaggerUIFS fs.FS, db *gorm.DB, readiness *appruntime.Readiness, operationalMetrics *appruntime.OperationalMetrics) error {
	probe := appruntime.NewServingProbe(db, readiness)
	router.Handle("/health", livenessHandler(probe)).Methods("GET")
	router.Handle("/readyz", readinessHandler(probe)).Methods("GET")
	router.Handle("/ops/status", operationalStatusHandler(probe, db, operationalMetrics)).Methods("GET")

	// OpenAPI spec endpoint
	router.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		_, _ = w.Write(openAPISpec)
	}).Methods("GET")

	// Swagger UI endpoints
	// Redirect /swagger -> /swagger/
	router.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	}).Methods("GET")
	// File server rooted at swagger-ui/
	uiFS, err := fs.Sub(swaggerUIFS, "swagger-ui")
	if err != nil {
		return fmt.Errorf("failed to set up swagger-ui FS: %w", err)
	}
	swaggerHandler := http.FileServer(http.FS(uiFS))
	router.PathPrefix("/swagger/").Handler(swaggerUIHeaders(http.StripPrefix("/swagger/", swaggerHandler))).Methods("GET")

	return nil
}

type operationalStatusResponse struct {
	Live                 bool                      `json:"live"`
	Ready                bool                      `json:"ready"`
	RequestFailuresTotal uint64                    `json:"requestFailuresTotal"`
	DBPool               appruntime.DBPoolSnapshot `json:"dbPool"`
}

func swaggerUIHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Swagger UI's bundled runtime uses dynamic function evaluation. Keep
		// that CSP exception scoped to the docs UI instead of the whole API.
		w.Header().Set("Content-Security-Policy", "default-src 'self'; "+
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data: https:; "+
			"connect-src 'self'; "+
			"font-src 'self'; "+
			"object-src 'none'; "+
			"media-src 'self'; "+
			"frame-src 'none'; "+
			"worker-src 'none'; "+
			"child-src 'none'; "+
			"form-action 'self'")
		next.ServeHTTP(w, r)
	})
}

func livenessHandler(probe appruntime.ServingProbe) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if !probe.Live() {
			writeProbeResponse(w, http.StatusServiceUnavailable, "not live")
			return
		}

		writeProbeResponse(w, http.StatusOK, "live")
	})
}

func readinessHandler(probe appruntime.ServingProbe) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readinessProbeTimeout)
		defer cancel()
		if err := probe.Ready(ctx); err != nil {
			writeProbeResponse(w, http.StatusServiceUnavailable, "not ready")
			return
		}

		writeProbeResponse(w, http.StatusOK, "ready")
	})
}

func operationalStatusHandler(probe appruntime.ServingProbe, db *gorm.DB, operationalMetrics *appruntime.OperationalMetrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readinessProbeTimeout)
		defer cancel()

		snapshot := operationalMetrics.Snapshot()
		response := operationalStatusResponse{
			Live:                 probe.Live(),
			Ready:                probe.Ready(ctx) == nil,
			RequestFailuresTotal: snapshot.RequestFailuresTotal,
			DBPool:               appruntime.SnapshotDBPool(db),
		}

		status := http.StatusOK
		if !response.Live || !response.Ready {
			status = http.StatusServiceUnavailable
		}

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(response)
	})
}

func writeProbeResponse(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

type reportingVisibilityService interface {
	GetSettings() (*models.ReportingVisibilitySettings, error)
}

func registerApplicationReportingRoutes(router *mux.Router, configService configsvc.Service, statsService statshandlers.FinancialStatsService, reportingService applicationReportingService, visibility reportingVisibilityService, auth authsvc.Authenticator, securityMiddleware func(http.Handler) http.Handler) {
	// These /v0/ reporting routes stay application-owned. Future tracing or metrics
	// export work belongs in request-boundary/runtime wiring, not in health probes.
	router.Handle("/v0/stats", securityMiddleware(statshandlers.StatsHandler(statsService, configService))).Methods("GET")
	router.Handle("/v0/system/metrics", securityMiddleware(reportingVisibilityGate(visibility, auth, func(s *models.ReportingVisibilitySettings) bool {
		return s == nil || s.SystemMetricsPublic
	}, metricshandlers.GetSystemMetricsHandler(reportingService)))).Methods("GET")
	router.Handle("/v0/global/leaderboard", securityMiddleware(reportingVisibilityGate(visibility, auth, func(s *models.ReportingVisibilitySettings) bool {
		return s == nil || s.GlobalLeaderboardPublic
	}, metricshandlers.GetGlobalLeaderboardHandler(reportingService)))).Methods("GET")
}

func reportingVisibilityGate(visibility reportingVisibilityService, auth authsvc.Authenticator, isPublic func(*models.ReportingVisibilitySettings) bool, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		settings := reportingvisibility.DefaultSettings()
		if visibility != nil {
			loaded, err := visibility.GetSettings()
			if err != nil {
				_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
				return
			}
			settings = loaded
		}
		if isPublic == nil || isPublic(settings) {
			next.ServeHTTP(w, r)
			return
		}
		if auth == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		if _, authErr := auth.CurrentUser(r); authErr != nil {
			_ = authhttp.WriteFailure(w, authErr)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requirePasswordChangeCleared(auth authsvc.Authenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		if _, authErr := auth.CurrentUser(r); authErr != nil {
			_ = authhttp.WriteFailure(w, authErr)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type marketDiscoveryStaleMarker interface {
	MarkMarketDiscoverySnapshotsStale(ctx context.Context, reason string) error
}

type responseStatusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseStatusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseStatusRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(data)
}

func markDiscoveryStaleOnSuccess(marker marketDiscoveryStaleMarker, reason string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &responseStatusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		if marker != nil && status >= 200 && status < 300 {
			_ = marker.MarkMarketDiscoverySnapshotsStale(r.Context(), reason)
		}
	})
}

func registerApplicationRoutes(router *mux.Router, db *gorm.DB, configService configsvc.Service, securityConfig appruntime.SecurityConfig) {
	container := app.BuildApplicationWithConfigAndJWTSigningKey(db, configService, securityConfig.JWTSigningKey)
	marketsService := container.GetMarketsService()
	usersService := container.GetUsersService()
	usersRepo := container.GetUsersRepository()
	analyticsService := container.GetAnalyticsService()
	authService := container.GetAuthService()
	requestSecurityService := container.GetSecurityService()
	readModelSnapshotRepo := readmodelrepo.NewGormRepository(db)
	readModelInvalidator := readmodelinvalidation.New(marketsService, analyticsService, readModelSnapshotRepo)

	// Create Handler instances
	marketsHandler := marketshandlers.NewHandler(marketsService, authService, requestSecurityService)
	marketsHandler.SetReadModelInvalidator(readModelInvalidator)

	// Define endpoint handlers using Gorilla Mux router
	// This defines all functions starting with /api/

	// Apply security middleware to all routes
	securityService := security.NewRuntimeSecurityService(securityConfig.RateLimit, securityConfig.Headers)
	securityMiddleware := securityService.SecurityMiddleware()
	loginSecurityMiddleware := securityService.LoginSecurityMiddleware()
	privateActionMiddleware := func(next http.Handler) http.Handler {
		return securityMiddleware(requirePasswordChangeCleared(authService, next))
	}

	router.HandleFunc("/v0/home", handlers.HomeHandler).Methods("GET")
	router.Handle("/v0/login", loginSecurityMiddleware(authsvc.LoginHandler(usersRepo, requestSecurityService, securityConfig.JWTSigningKey))).Methods("POST")

	// application setup information
	router.Handle("/v0/setup", securityMiddleware(http.HandlerFunc(setuphandlers.GetSetupHandler(container.GetConfigService())))).Methods("GET")
	router.Handle("/v0/setup/frontend", securityMiddleware(http.HandlerFunc(setuphandlers.GetFrontendSetupHandler(container.GetConfigService())))).Methods("GET")
	reportingVisibilityRepo := reportingvisibility.NewGormRepository(db)
	reportingVisibilitySvc := reportingvisibility.NewService(reportingVisibilityRepo)
	reportingVisibilityHandler := cmsreportinghttp.NewHandler(reportingVisibilitySvc, authService)
	registerApplicationReportingRoutes(router, configService, analyticsService, analyticsService, reportingVisibilitySvc, authService, securityMiddleware)

	// CMS routes and services
	homepageRepo := homepage.NewGormRepository(db)
	homepageRenderer := homepage.NewDefaultRenderer()
	homepageSvc := homepage.NewService(homepageRepo, homepageRenderer)
	homepageHandler := cmshomehttp.NewHandler(homepageSvc, authService)

	marketDiscoveryRepo := marketdiscovery.NewGormRepository(db)
	marketDiscoverySvc := marketdiscovery.NewService(marketDiscoveryRepo)
	marketDiscoveryHandler := cmsdiscoveryhttp.NewHandler(marketDiscoverySvc, authService)
	marketDiscoveryHandler.SetReadModelInvalidator(readModelSnapshotRepo)
	marketDiscoveryReadModelHandler := marketshandlers.NewMarketDiscoveryReadModelHandler(marketsService, marketDiscoverySvc, readModelSnapshotRepo)

	socialShareRepo := socialshare.NewGormRepository(db)
	socialShareSvc := socialshare.NewService(socialShareRepo)
	socialShareHandler := cmssocialhttp.NewHandler(socialShareSvc, authService)
	socialShareConfigProvider := func(_ context.Context, fallback dmarkets.ShareMetadataConfig) dmarkets.ShareMetadataConfig {
		settings, err := socialShareSvc.GetSettings()
		if err != nil || settings == nil {
			return fallback
		}
		if strings.TrimSpace(settings.SiteName) != "" {
			fallback.SiteName = settings.SiteName
		}
		if strings.TrimSpace(settings.DefaultImageURL) != "" {
			fallback.DefaultImageURL = settings.DefaultImageURL
		}
		fallback.DisableImage = !settings.ImageEnabled
		if strings.TrimSpace(settings.ImageAlt) != "" {
			fallback.DefaultImageAlt = settings.ImageAlt
		}
		if strings.TrimSpace(settings.DefaultDescription) != "" {
			fallback.DefaultDescription = settings.DefaultDescription
		}
		return fallback
	}

	// Markets routes - using new Handler instance
	router.Handle("/markets/{id:[0-9]+}", securityMiddleware(marketshandlers.MarketShareShellHandler(marketsService, shareMetadataConfig(securityConfig.Share), socialShareConfigProvider))).Methods("GET")
	router.Handle("/v0/markets", securityMiddleware(http.HandlerFunc(marketsHandler.ListMarkets))).Methods("GET")
	router.Handle("/v0/read/market-discovery/{slug}", securityMiddleware(http.HandlerFunc(marketDiscoveryReadModelHandler.Get))).Methods("GET")
	router.Handle("/v0/markets", securityMiddleware(http.HandlerFunc(marketsHandler.CreateMarket))).Methods("POST")
	router.Handle("/v0/market-groups", securityMiddleware(http.HandlerFunc(marketsHandler.CreateMarketGroup))).Methods("POST")
	router.Handle("/v0/market-groups/{id}/bets", securityMiddleware(http.HandlerFunc(marketsHandler.MarketGroupBets))).Methods("GET")
	router.Handle("/v0/market-groups/{id}/positions", securityMiddleware(http.HandlerFunc(marketsHandler.MarketGroupPositions))).Methods("GET")
	router.Handle("/v0/market-groups/{id}/leaderboard", securityMiddleware(http.HandlerFunc(marketsHandler.MarketGroupLeaderboard))).Methods("GET")
	router.Handle("/v0/market-groups/{id}/resolve", securityMiddleware(http.HandlerFunc(marketsHandler.ResolveMarketGroup))).Methods("POST")
	router.Handle("/v0/market-groups/{id}", securityMiddleware(http.HandlerFunc(marketsHandler.GetMarketGroup))).Methods("GET")
	router.Handle("/v0/markets/search", securityMiddleware(http.HandlerFunc(marketsHandler.SearchMarkets))).Methods("GET")
	router.Handle("/v0/markets/status/{status}", securityMiddleware(http.HandlerFunc(marketsHandler.ListByStatus))).Methods("GET")
	router.Handle("/v0/markets/status", securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rWithStatus := mux.SetURLVars(r, map[string]string{"status": "all"})
		marketsHandler.ListByStatus(w, rWithStatus)
	}))).Methods("GET")
	router.Handle("/v0/read/markets/{id}/summary", securityMiddleware(http.HandlerFunc(marketsHandler.MarketSummaryReadModel))).Methods("GET")

	// Legacy routes for backward compatibility — rewrite to new handler with status query
	router.Handle("/v0/markets/active", securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Set("status", "active")
		r.URL.RawQuery = q.Encode()
		marketsHandler.ListMarkets(w, r)
	}))).Methods("GET")
	router.Handle("/v0/markets/closed", securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Set("status", "closed")
		r.URL.RawQuery = q.Encode()
		marketsHandler.ListMarkets(w, r)
	}))).Methods("GET")
	router.Handle("/v0/markets/resolved", securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Set("status", "resolved")
		r.URL.RawQuery = q.Encode()
		marketsHandler.ListMarkets(w, r)
	}))).Methods("GET")
	router.Handle("/v0/markets/{id}", securityMiddleware(http.HandlerFunc(marketsHandler.GetDetails))).Methods("GET")
	router.Handle("/v0/markets/{id}/resolve", securityMiddleware(http.HandlerFunc(marketsHandler.ResolveMarket))).Methods("POST")
	router.Handle("/v0/markets/{id}/description-amendments", privateActionMiddleware(http.HandlerFunc(marketsHandler.ProposeDescriptionAmendment))).Methods("POST")
	router.Handle("/v0/markets/{id}/leaderboard", securityMiddleware(http.HandlerFunc(marketsHandler.MarketLeaderboard))).Methods("GET")
	router.Handle("/v0/markets/{id}/projection", securityMiddleware(http.HandlerFunc(marketsHandler.ProjectProbability))).Methods("GET")
	router.Handle("/v0/market-tags", securityMiddleware(marketshandlers.ListMarketTagsHandler(marketsService))).Methods("GET")
	router.Handle("/v0/marketprojection/{marketId}/{amount}/{outcome}", securityMiddleware(marketshandlers.ProjectNewProbabilityHandler(marketsService))).Methods("GET")
	router.Handle("/v0/marketprojection/{marketId}/{amount}/{outcome}/", securityMiddleware(marketshandlers.ProjectNewProbabilityHandler(marketsService))).Methods("GET")

	// handle market positions, get trades - using service injection from new locations
	router.Handle("/v0/markets/bets/{marketId}", securityMiddleware(betshandlers.MarketBetsHandlerWithService(marketsService))).Methods("GET")
	router.Handle("/v0/markets/positions/{marketId}", securityMiddleware(positionshandlers.MarketPositionsHandlerWithService(marketsService))).Methods("GET")
	router.Handle("/v0/markets/positions/{marketId}/{username}", securityMiddleware(positionshandlers.MarketUserPositionHandlerWithService(marketsService))).Methods("GET")

	// handle public user stuff
	router.Handle("/v0/userinfo/{username}", securityMiddleware(usershandlers.GetPublicUserHandler(usersService))).Methods("GET")
	router.Handle("/v0/usercredit/{username}", securityMiddleware(usercredit.GetUserCreditHandler(usersService, configService.Economics().User.MaximumDebtAllowed))).Methods("GET")
	router.Handle("/v0/portfolio/{username}", securityMiddleware(publicuser.GetPortfolioHandler(usersService))).Methods("GET")
	router.Handle("/v0/users/{username}/financial", securityMiddleware(usershandlers.GetUserFinancialHandler(usersService))).Methods("GET")
	router.Handle("/v0/read/users/{username}/financial-summary", securityMiddleware(usershandlers.GetUserFinancialReadModelHandler(analyticsService, authService))).Methods("GET")
	router.Handle("/v0/users/{username}/owned-markets", securityMiddleware(marketshandlers.ListUserOwnedMarketsHandler(marketsService, authService))).Methods("GET")

	// handle private user stuff, display sensitive profile information to customize
	router.Handle("/v0/privateprofile", securityMiddleware(privateuser.GetPrivateProfileHandler(usersService))).Methods("GET")
	router.Handle("/v0/profile/markets", securityMiddleware(marketshandlers.ListMyLifecycleMarketsHandler(marketsService, authService))).Methods("GET")
	router.Handle("/v0/profile/market-description-amendments", securityMiddleware(http.HandlerFunc(marketsHandler.ListMyDescriptionAmendments))).Methods("GET")

	// changing profile stuff - apply security middleware
	router.Handle("/v0/changepassword", securityMiddleware(usershandlers.ChangePasswordHandler(usersService))).Methods("POST")
	router.Handle("/v0/profilechange/displayname", securityMiddleware(usershandlers.ChangeDisplayNameHandler(usersService))).Methods("POST")
	router.Handle("/v0/profilechange/emoji", securityMiddleware(usershandlers.ChangeEmojiHandler(usersService))).Methods("POST")
	router.Handle("/v0/profilechange/description", securityMiddleware(usershandlers.ChangeDescriptionHandler(usersService))).Methods("POST")
	router.Handle("/v0/profilechange/links", securityMiddleware(usershandlers.ChangePersonalLinksHandler(usersService))).Methods("POST")

	// handle private user actions such as make a bet, sell positions, get user position
	router.Handle("/v0/bet", privateActionMiddleware(buybetshandlers.PlaceBetHandlerWithInvalidator(container.GetBetsService(), container.GetUsersService(), readModelInvalidator))).Methods("POST")
	router.Handle("/v0/userposition/{marketId}", privateActionMiddleware(usershandlers.UserMarketPositionHandlerWithService(marketsService, usersService))).Methods("GET")
	router.Handle("/v0/sell/quote", privateActionMiddleware(sellbetshandlers.SellQuoteHandler(container.GetBetsService(), container.GetUsersService()))).Methods("POST")
	router.Handle("/v0/sell", privateActionMiddleware(sellbetshandlers.SellPositionHandlerWithInvalidator(container.GetBetsService(), container.GetUsersService(), readModelInvalidator))).Methods("POST")

	// admin stuff - apply security middleware
	router.Handle("/v0/admin/createuser", securityMiddleware(http.HandlerFunc(adminhandlers.AddUserHandler(usersService, container.GetConfigService(), authService, requestSecurityService)))).Methods("POST")
	router.Handle("/v0/admin/users", securityMiddleware(adminhandlers.ListAdminUsersHandler(usersService, authService))).Methods("GET")
	router.Handle("/v0/admin/users/{username}/role", securityMiddleware(adminhandlers.UpdateAdminUserRoleHandler(usersService, authService))).Methods("PATCH")
	router.Handle("/v0/admin/moderators/{username}/suspension", securityMiddleware(adminhandlers.UpdateAdminModeratorSuspensionHandler(usersService, authService, time.Now))).Methods("PATCH")
	router.Handle("/v0/admin/markets", securityMiddleware(adminhandlers.ListReviewMarketsHandler(marketsService, authService))).Methods("GET")
	router.Handle("/v0/admin/markets/{id}/approve", securityMiddleware(markDiscoveryStaleOnSuccess(readModelSnapshotRepo, "market_status_changed", adminhandlers.ApproveMarketHandler(marketsService, authService)))).Methods("PATCH")
	router.Handle("/v0/admin/markets/{id}/reject", securityMiddleware(markDiscoveryStaleOnSuccess(readModelSnapshotRepo, "market_status_changed", adminhandlers.RejectMarketHandler(marketsService, authService)))).Methods("PATCH")
	router.Handle("/v0/admin/market-groups/{id}/approve", securityMiddleware(markDiscoveryStaleOnSuccess(readModelSnapshotRepo, "market_status_changed", adminhandlers.ApproveMarketGroupHandler(marketsService, authService)))).Methods("PATCH")
	router.Handle("/v0/admin/market-groups/{id}/reject", securityMiddleware(markDiscoveryStaleOnSuccess(readModelSnapshotRepo, "market_status_changed", adminhandlers.RejectMarketGroupHandler(marketsService, authService)))).Methods("PATCH")
	router.Handle("/v0/admin/market-groups/{id}/steward", securityMiddleware(markDiscoveryStaleOnSuccess(readModelSnapshotRepo, "market_steward_changed", adminhandlers.ReassignMarketGroupStewardHandler(marketsService, authService)))).Methods("PATCH")
	router.Handle("/v0/admin/markets/{id}/steward", securityMiddleware(adminhandlers.ReassignMarketStewardHandler(marketsService, authService))).Methods("PATCH")
	router.Handle("/v0/admin/markets/{id}/tags", securityMiddleware(markDiscoveryStaleOnSuccess(readModelSnapshotRepo, "market_tags_changed", adminhandlers.UpdateMarketTagsHandler(marketsService, authService)))).Methods("PATCH")
	router.Handle("/v0/admin/market-description-amendments", securityMiddleware(adminhandlers.ListMarketDescriptionAmendmentsHandler(marketsService, authService))).Methods("GET")
	router.Handle("/v0/admin/market-description-amendments/settings", securityMiddleware(adminhandlers.GetMarketGovernanceSettingsHandler(marketsService, authService))).Methods("GET")
	router.Handle("/v0/admin/market-description-amendments/settings", securityMiddleware(adminhandlers.UpdateMarketGovernanceSettingsHandler(marketsService, authService))).Methods("PUT")
	router.Handle("/v0/admin/market-description-amendments/{id}", securityMiddleware(adminhandlers.ReviewMarketDescriptionAmendmentHandler(marketsService, authService))).Methods("PATCH")
	router.Handle("/v0/admin/market-tags", securityMiddleware(adminhandlers.ListAdminMarketTagsHandler(marketsService, authService))).Methods("GET")
	router.Handle("/v0/admin/market-tags", securityMiddleware(markDiscoveryStaleOnSuccess(readModelSnapshotRepo, "tag_catalog_changed", adminhandlers.CreateAdminMarketTagHandler(marketsService, authService)))).Methods("POST")
	router.Handle("/v0/admin/market-tags/{slug}", securityMiddleware(markDiscoveryStaleOnSuccess(readModelSnapshotRepo, "tag_catalog_changed", adminhandlers.UpdateAdminMarketTagHandler(marketsService, authService)))).Methods("PATCH")

	router.HandleFunc("/v0/content/home", homepageHandler.PublicGet).Methods("GET")
	router.Handle("/v0/admin/content/home", securityMiddleware(http.HandlerFunc(homepageHandler.AdminUpdate))).Methods("PUT")
	router.HandleFunc("/v0/content/market-discovery/{slug}", marketDiscoveryHandler.PublicGet).Methods("GET")
	router.Handle("/v0/admin/content/market-discovery/{slug}", securityMiddleware(http.HandlerFunc(marketDiscoveryHandler.AdminUpdate))).Methods("PUT")
	router.Handle("/v0/admin/content/market-discovery/{slug}/pins", securityMiddleware(http.HandlerFunc(marketDiscoveryHandler.AdminReplacePins))).Methods("PUT")
	router.HandleFunc("/v0/content/social-share", socialShareHandler.PublicGet).Methods("GET")
	router.HandleFunc("/v0/content/social-share/image", socialShareHandler.PublicImage).Methods("GET", "HEAD")
	router.HandleFunc("/api/v0/content/social-share/image", socialShareHandler.PublicImage).Methods("GET", "HEAD")
	router.Handle("/v0/admin/content/social-share", securityMiddleware(http.HandlerFunc(socialShareHandler.AdminUpdate))).Methods("PUT")
	router.Handle("/v0/admin/content/social-share/image", securityMiddleware(http.HandlerFunc(socialShareHandler.AdminUploadImage))).Methods("POST")
	router.Handle("/v0/content/reporting-visibility", securityMiddleware(http.HandlerFunc(reportingVisibilityHandler.PublicGet))).Methods("GET")
	router.Handle("/v0/admin/content/reporting-visibility", securityMiddleware(http.HandlerFunc(reportingVisibilityHandler.AdminUpdate))).Methods("PUT")
}

func shareMetadataConfig(config appruntime.ShareConfig) dmarkets.ShareMetadataConfig {
	return dmarkets.ShareMetadataConfig{
		PublicBaseURL:      config.PublicBaseURL,
		DefaultImageURL:    config.DefaultImageURL,
		DefaultImageAlt:    socialshare.DefaultImageAlt,
		SiteName:           config.SiteName,
		DefaultDescription: socialshare.DefaultDescription,
	}
}

type gracefulShutdowner interface {
	Shutdown(context.Context) error
}

func shutdownHTTPServer(server gracefulShutdowner, readiness *appruntime.Readiness, config appruntime.ShutdownConfig, sleep func(time.Duration)) error {
	if readiness != nil {
		readiness.MarkNotReady()
	}
	if sleep == nil {
		sleep = time.Sleep
	}
	config = appruntime.NormalizeShutdownConfig(config)
	sleep(config.ReadinessDrainWindow)

	shutdownContext, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	return server.Shutdown(shutdownContext)
}

func Start(openAPISpec []byte, swaggerUIFS embed.FS, db *gorm.DB, configService configsvc.Service, readiness *appruntime.Readiness, securityConfig appruntime.SecurityConfig, shutdownConfig appruntime.ShutdownConfig) {
	authsvc.ConfigureJWTSigningKey(securityConfig.JWTSigningKey)
	handler, err := buildHandler(openAPISpec, swaggerUIFS, db, configService, readiness, securityConfig)
	if err != nil {
		logger.Fatal("server", "http handler initialization failed", err, logger.Operation("buildHandler"))
	}
	shutdownConfig = appruntime.NormalizeShutdownConfig(shutdownConfig)

	// Allow BACKEND_PORT to be configured via environment, default to 8080
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}

	address := ":" + port
	server := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	serveErrs := make(chan error, 1)
	go func() {
		serveErrs <- server.ListenAndServe()
	}()

	logger.Info("server", "HTTP server listening", logger.Operation("Start"), logger.Address(address))

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(shutdownSignals)

	select {
	case err := <-serveErrs:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server", "http server exited unexpectedly", err, logger.Operation("ListenAndServe"), logger.Address(address))
		}
		logger.Info("server", "HTTP server stopped", logger.Operation("ListenAndServe"), logger.Address(address))
	case shutdownSignal := <-shutdownSignals:
		logger.Info(
			"server",
			"shutdown signal received",
			logger.Operation("Shutdown"),
			logger.Address(address),
			logger.String("signal", shutdownSignal.String()),
			logger.String("readinessDrainWindow", shutdownConfig.ReadinessDrainWindow.String()),
			logger.String("shutdownTimeout", shutdownConfig.ShutdownTimeout.String()),
		)

		if err := shutdownHTTPServer(server, readiness, shutdownConfig, time.Sleep); err != nil {
			logger.Fatal("server", "graceful shutdown failed", err, logger.Operation("Shutdown"), logger.Address(address))
		}

		if err := <-serveErrs; err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server", "http server exited unexpectedly during shutdown", err, logger.Operation("ListenAndServe"), logger.Address(address))
		}

		logger.Info("server", "HTTP server shutdown complete", logger.Operation("Shutdown"), logger.Address(address))
	}
}
