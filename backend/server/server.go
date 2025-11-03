package server

import (
	"log"
	"net/http"
	"os"
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
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/security"
	"socialpredict/setup"
	"socialpredict/util"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
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

func Start() {
	// Initialize security service
	securityService := security.NewSecurityService()

	// CORS handler (configurable via env)
	c := buildCORSFromEnv()

	// Initialize mux router
	router := mux.NewRouter()

	// Health endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}).Methods("GET")

	// Initialize domain services
	db := util.GetDB()
	econConfig := setup.EconomicsConfig()
	container := app.BuildApplication(db, econConfig)
	marketsService := container.GetMarketsService()
	usersService := container.GetUsersService()
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
	router.Handle("/v0/login", loginSecurityMiddleware(http.HandlerFunc(authsvc.LoginHandler))).Methods("POST")

	// application setup and stats information
	router.Handle("/v0/setup", securityMiddleware(http.HandlerFunc(setuphandlers.GetSetupHandler(setup.LoadEconomicsConfig)))).Methods("GET")
	router.Handle("/v0/stats", securityMiddleware(http.HandlerFunc(statshandlers.StatsHandler()))).Methods("GET")
	router.Handle("/v0/system/metrics", securityMiddleware(metricshandlers.GetSystemMetricsHandler(analyticsService))).Methods("GET")
	router.Handle("/v0/global/leaderboard", securityMiddleware(http.HandlerFunc(metricshandlers.GetGlobalLeaderboardHandler))).Methods("GET")

	// Markets routes - using new Handler instance
	router.Handle("/v0/markets", securityMiddleware(http.HandlerFunc(marketsHandler.ListMarkets))).Methods("GET")
	router.Handle("/v0/markets", securityMiddleware(http.HandlerFunc(marketsHandler.CreateMarket))).Methods("POST")
	router.Handle("/v0/markets/search", securityMiddleware(http.HandlerFunc(marketsHandler.SearchMarkets))).Methods("GET")
	router.Handle("/v0/markets/status/{status}", securityMiddleware(http.HandlerFunc(marketsHandler.ListByStatus))).Methods("GET")
	router.Handle("/v0/markets/{id}", securityMiddleware(http.HandlerFunc(marketsHandler.GetDetails))).Methods("GET")
	router.Handle("/v0/markets/{id}/resolve", securityMiddleware(http.HandlerFunc(marketsHandler.ResolveMarket))).Methods("POST")
	router.Handle("/v0/markets/{id}/leaderboard", securityMiddleware(http.HandlerFunc(marketsHandler.MarketLeaderboard))).Methods("GET")
	router.Handle("/v0/markets/{id}/projection", securityMiddleware(http.HandlerFunc(marketsHandler.ProjectProbability))).Methods("GET")

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
	router.Handle("/v0/marketprojection/{marketId}/{amount}/{outcome}/", securityMiddleware(marketshandlers.ProjectNewProbabilityHandler(marketsService))).Methods("GET")

	// handle market positions, get trades - using service injection from new locations
	router.Handle("/v0/markets/bets/{marketId}", securityMiddleware(betshandlers.MarketBetsHandlerWithService(marketsService))).Methods("GET")
	router.Handle("/v0/markets/positions/{marketId}", securityMiddleware(positionshandlers.MarketPositionsHandlerWithService(marketsService))).Methods("GET")
	router.Handle("/v0/markets/positions/{marketId}/{username}", securityMiddleware(positionshandlers.MarketUserPositionHandlerWithService(marketsService))).Methods("GET")

	// handle public user stuff
	router.Handle("/v0/userinfo/{username}", securityMiddleware(usershandlers.GetPublicUserHandler(usersService))).Methods("GET")
	router.Handle("/v0/usercredit/{username}", securityMiddleware(usercredit.GetUserCreditHandler(usersService, econConfig.Economics.User.MaximumDebtAllowed))).Methods("GET")
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
	router.Handle("/v0/admin/createuser", securityMiddleware(http.HandlerFunc(adminhandlers.AddUserHandler(setup.EconomicsConfig, authService)))).Methods("POST")

	// homepage content routes
	homepageRepo := homepage.NewGormRepository(db)
	homepageRenderer := homepage.NewDefaultRenderer()
	homepageSvc := homepage.NewService(homepageRepo, homepageRenderer)
	homepageHandler := cmshomehttp.NewHandler(homepageSvc, authService)

	router.HandleFunc("/v0/content/home", homepageHandler.PublicGet).Methods("GET")
	router.Handle("/v0/admin/content/home", securityMiddleware(http.HandlerFunc(homepageHandler.AdminUpdate))).Methods("PUT")

	// Apply CORS middleware if enabled
	handler := http.Handler(router)
	if c != nil {
		handler = c.Handler(handler)
	}

	// Allow BACKEND_PORT to be configured via environment, default to 8080
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
