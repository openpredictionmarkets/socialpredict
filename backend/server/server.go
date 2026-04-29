package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"socialpredict/handlers"
	adminhandlers "socialpredict/handlers/admin"
	betshandlers "socialpredict/handlers/bets"
	buybetshandlers "socialpredict/handlers/bets/buying"
	sellbetshandlers "socialpredict/handlers/bets/selling"
	"socialpredict/handlers/cms/homepage"
	cmshomehttp "socialpredict/handlers/cms/homepage/http"
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
	appruntime "socialpredict/internal/app/runtime"
	authsvc "socialpredict/internal/service/auth"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/logger"
	"socialpredict/security"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gorm.io/gorm"
)

// CORS helpers configured via environment variables

func getListEnv(key, def string) []string { // default empty - allows any string, splits on comma
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		val = def
	}
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getBoolEnv(key string, def bool) bool { // default false - allows any string to be false except specific true values
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return def
	}
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func getIntEnv(key string, def int) int { // default 0 - allows any string to be int, otherwise default
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return def
}

func buildCORSFromEnv() *cors.Cors {
	if !getBoolEnv("CORS_ENABLED", true) {
		return nil
	}
	origins := getListEnv("CORS_ALLOW_ORIGINS", "*")
	methods := getListEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	headers := getListEnv("CORS_ALLOW_HEADERS", "Content-Type,Authorization")
	expose := getListEnv("CORS_EXPOSE_HEADERS", "")
	allowCreds := getBoolEnv("CORS_ALLOW_CREDENTIALS", false)
	maxAge := getIntEnv("CORS_MAX_AGE", 600)

	return cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   methods,
		AllowedHeaders:   headers,
		ExposedHeaders:   expose,
		AllowCredentials: allowCreds,
		MaxAge:           maxAge,
	})
}

func buildHandler(openAPISpec []byte, swaggerUIFS fs.FS, db *gorm.DB, configService configsvc.Service, readiness *appruntime.Readiness) (http.Handler, error) {
	router, err := buildRouter(openAPISpec, swaggerUIFS, db, configService, readiness)
	if err != nil {
		return nil, err
	}

	handler := http.Handler(router)
	if c := buildCORSFromEnv(); c != nil {
		handler = c.Handler(handler)
	}
	handler = security.RequestBoundaryMiddleware()(handler)

	return handler, nil
}

func buildRouter(openAPISpec []byte, swaggerUIFS fs.FS, db *gorm.DB, configService configsvc.Service, readiness *appruntime.Readiness) (*mux.Router, error) {
	if configService == nil {
		return nil, fmt.Errorf("config init: configuration service unavailable")
	}

	router := mux.NewRouter()
	router.MethodNotAllowedHandler = methodNotAllowedHandler(router)
	if err := registerInfraRoutes(router, openAPISpec, swaggerUIFS, db, readiness); err != nil {
		return nil, err
	}

	registerApplicationRoutes(router, db, configService, security.NewSecurityService())
	return router, nil
}

func methodNotAllowedHandler(router *mux.Router) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allow := strings.Join(allowedMethodsForRequest(router, r), ", "); allow != "" {
			w.Header().Set("Allow", allow)
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
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

func registerInfraRoutes(router *mux.Router, openAPISpec []byte, swaggerUIFS fs.FS, db *gorm.DB, readiness *appruntime.Readiness) error {
	router.HandleFunc("/health", healthHandler).Methods("GET")
	router.Handle("/readyz", readinessHandler(db, readiness)).Methods("GET")

	// OpenAPI spec endpoint
	router.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		_, _ = w.Write(openAPISpec)
	}).Methods("GET")

	// Swagger UI endpoints
	// Redirect /swagger -> /swagger/
	router.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	// File server rooted at swagger-ui/
	uiFS, err := fs.Sub(swaggerUIFS, "swagger-ui")
	if err != nil {
		return fmt.Errorf("failed to set up swagger-ui FS: %w", err)
	}
	swaggerHandler := http.FileServer(http.FS(uiFS))
	router.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", swaggerHandler))

	return nil
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeProbeResponse(w, http.StatusOK, "ok")
}

func readinessHandler(db *gorm.DB, readiness *appruntime.Readiness) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if readiness == nil || !readiness.Ready() {
			writeProbeResponse(w, http.StatusServiceUnavailable, "not ready")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), readinessProbeTimeout)
		defer cancel()
		if err := appruntime.CheckDBReadiness(ctx, db); err != nil {
			writeProbeResponse(w, http.StatusServiceUnavailable, "not ready")
			return
		}

		writeProbeResponse(w, http.StatusOK, "ready")
	})
}

func writeProbeResponse(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func registerApplicationReportingRoutes(router *mux.Router, db *gorm.DB, configService configsvc.Service, reportingService applicationReportingService, securityMiddleware func(http.Handler) http.Handler) {
	// These /v0/ reporting routes stay application-owned. Future tracing or metrics
	// export work belongs in request-boundary/runtime wiring, not in health probes.
	router.Handle("/v0/stats", securityMiddleware(statshandlers.StatsHandler(db, configService))).Methods("GET")
	router.Handle("/v0/system/metrics", securityMiddleware(metricshandlers.GetSystemMetricsHandler(reportingService))).Methods("GET")
	router.Handle("/v0/global/leaderboard", securityMiddleware(metricshandlers.GetGlobalLeaderboardHandler(reportingService))).Methods("GET")
}

func registerApplicationRoutes(router *mux.Router, db *gorm.DB, configService configsvc.Service, securityService *security.SecurityService) {
	container := app.BuildApplicationWithConfigService(db, configService)
	marketsService := container.GetMarketsService()
	usersService := container.GetUsersService()
	usersRepo := container.GetUsersRepository()
	analyticsService := container.GetAnalyticsService()
	authService := container.GetAuthService()

	// Create Handler instances
	marketsHandler := marketshandlers.NewHandler(marketsService, authService)

	// Define endpoint handlers using Gorilla Mux router
	// This defines all functions starting with /api/

	// Apply security middleware to all routes
	securityMiddleware := securityService.SecurityMiddleware()
	loginSecurityMiddleware := securityService.LoginSecurityMiddleware()

	router.HandleFunc("/v0/home", handlers.HomeHandler).Methods("GET")
	router.Handle("/v0/login", loginSecurityMiddleware(authsvc.LoginHandler(usersRepo))).Methods("POST")

	// application setup information
	router.Handle("/v0/setup", securityMiddleware(http.HandlerFunc(setuphandlers.GetSetupHandler(container.GetConfigService())))).Methods("GET")
	router.Handle("/v0/setup/frontend", securityMiddleware(http.HandlerFunc(setuphandlers.GetFrontendSetupHandler(container.GetConfigService())))).Methods("GET")
	registerApplicationReportingRoutes(router, db, configService, analyticsService, securityMiddleware)

	// Markets routes - using new Handler instance
	router.Handle("/v0/markets", securityMiddleware(http.HandlerFunc(marketsHandler.ListMarkets))).Methods("GET")
	router.Handle("/v0/markets", securityMiddleware(http.HandlerFunc(marketsHandler.CreateMarket))).Methods("POST")
	router.Handle("/v0/markets/search", securityMiddleware(http.HandlerFunc(marketsHandler.SearchMarkets))).Methods("GET")
	router.Handle("/v0/markets/status/{status}", securityMiddleware(http.HandlerFunc(marketsHandler.ListByStatus))).Methods("GET")
	router.Handle("/v0/markets/status", securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rWithStatus := mux.SetURLVars(r, map[string]string{"status": "all"})
		marketsHandler.ListByStatus(w, rWithStatus)
	}))).Methods("GET")

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
	router.Handle("/v0/markets/{id}/leaderboard", securityMiddleware(http.HandlerFunc(marketsHandler.MarketLeaderboard))).Methods("GET")
	router.Handle("/v0/markets/{id}/projection", securityMiddleware(http.HandlerFunc(marketsHandler.ProjectProbability))).Methods("GET")
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

	// handle private user stuff, display sensitive profile information to customize
	router.Handle("/v0/privateprofile", securityMiddleware(privateuser.GetPrivateProfileHandler(usersService))).Methods("GET")

	// changing profile stuff - apply security middleware
	router.Handle("/v0/changepassword", securityMiddleware(usershandlers.ChangePasswordHandler(usersService))).Methods("POST")
	router.Handle("/v0/profilechange/displayname", securityMiddleware(usershandlers.ChangeDisplayNameHandler(usersService))).Methods("POST")
	router.Handle("/v0/profilechange/emoji", securityMiddleware(usershandlers.ChangeEmojiHandler(usersService))).Methods("POST")
	router.Handle("/v0/profilechange/description", securityMiddleware(usershandlers.ChangeDescriptionHandler(usersService))).Methods("POST")
	router.Handle("/v0/profilechange/links", securityMiddleware(usershandlers.ChangePersonalLinksHandler(usersService))).Methods("POST")

	// handle private user actions such as make a bet, sell positions, get user position
	router.Handle("/v0/bet", securityMiddleware(buybetshandlers.PlaceBetHandler(container.GetBetsService(), container.GetUsersService()))).Methods("POST")
	router.Handle("/v0/userposition/{marketId}", securityMiddleware(usershandlers.UserMarketPositionHandlerWithService(marketsService, usersService))).Methods("GET")
	router.Handle("/v0/sell", securityMiddleware(sellbetshandlers.SellPositionHandler(container.GetBetsService(), container.GetUsersService()))).Methods("POST")

	// admin stuff - apply security middleware
	router.Handle("/v0/admin/createuser", securityMiddleware(http.HandlerFunc(adminhandlers.AddUserHandler(db, container.GetConfigService(), authService)))).Methods("POST")

	// homepage content routes
	homepageRepo := homepage.NewGormRepository(db)
	homepageRenderer := homepage.NewDefaultRenderer()
	homepageSvc := homepage.NewService(homepageRepo, homepageRenderer)
	homepageHandler := cmshomehttp.NewHandler(homepageSvc, authService)

	router.HandleFunc("/v0/content/home", homepageHandler.PublicGet).Methods("GET")
	router.Handle("/v0/admin/content/home", securityMiddleware(http.HandlerFunc(homepageHandler.AdminUpdate))).Methods("PUT")
}

func Start(openAPISpec []byte, swaggerUIFS embed.FS, db *gorm.DB, configService configsvc.Service, readiness *appruntime.Readiness) {
	handler, err := buildHandler(openAPISpec, swaggerUIFS, db, configService, readiness)
	if err != nil {
		logger.Fatal("server", "http handler initialization failed", err, logger.Operation("buildHandler"))
	}

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
		if readiness != nil {
			readiness.MarkNotReady()
		}
		logger.Info("server", "shutdown signal received", logger.Operation("Shutdown"), logger.Address(address), logger.String("signal", shutdownSignal.String()))

		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownContext); err != nil {
			logger.Fatal("server", "graceful shutdown failed", err, logger.Operation("Shutdown"), logger.Address(address))
		}

		if err := <-serveErrs; err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server", "http server exited unexpectedly during shutdown", err, logger.Operation("ListenAndServe"), logger.Address(address))
		}

		logger.Info("server", "HTTP server shutdown complete", logger.Operation("Shutdown"), logger.Address(address))
	}
}
