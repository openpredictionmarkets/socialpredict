package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
	"socialpredict/models/modelstesting"
	"socialpredict/security"
	"socialpredict/setup"
	"socialpredict/util"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func buildTestHandler(t *testing.T) http.Handler {
	t.Helper()

	securityService := security.NewSecurityService()
	c := buildCORSFromEnv()
	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}).Methods("GET")

	db := util.GetDB()
	econConfig := setup.EconomicsConfig()
	container := app.BuildApplication(db, econConfig)
	marketsService := container.GetMarketsService()
	usersService := container.GetUsersService()
	analyticsService := container.GetAnalyticsService()
	authService := container.GetAuthService()
	betsService := container.GetBetsService()

	marketsHandler := marketshandlers.NewHandler(marketsService, authService)

	securityMiddleware := securityService.SecurityMiddleware()
	loginSecurityMiddleware := securityService.LoginSecurityMiddleware()

	router.HandleFunc("/v0/home", handlers.HomeHandler).Methods("GET")
	router.Handle("/v0/login", loginSecurityMiddleware(http.HandlerFunc(authsvc.LoginHandler))).Methods("POST")

	router.Handle("/v0/setup", securityMiddleware(http.HandlerFunc(setuphandlers.GetSetupHandler(setup.LoadEconomicsConfig)))).Methods("GET")
	router.Handle("/v0/stats", securityMiddleware(http.HandlerFunc(statshandlers.StatsHandler()))).Methods("GET")
	router.Handle("/v0/system/metrics", securityMiddleware(metricshandlers.GetSystemMetricsHandler(analyticsService))).Methods("GET")
	router.Handle("/v0/global/leaderboard", securityMiddleware(metricshandlers.GetGlobalLeaderboardHandler(analyticsService))).Methods("GET")

	router.Handle("/v0/markets", securityMiddleware(http.HandlerFunc(marketsHandler.ListMarkets))).Methods("GET")
	router.Handle("/v0/markets", securityMiddleware(http.HandlerFunc(marketsHandler.CreateMarket))).Methods("POST")
	router.Handle("/v0/markets/search", securityMiddleware(http.HandlerFunc(marketsHandler.SearchMarkets))).Methods("GET")
	router.Handle("/v0/markets/status/{status}", securityMiddleware(http.HandlerFunc(marketsHandler.ListByStatus))).Methods("GET")
	router.Handle("/v0/markets/{id}", securityMiddleware(http.HandlerFunc(marketsHandler.GetDetails))).Methods("GET")
	router.Handle("/v0/markets/{id}/resolve", securityMiddleware(http.HandlerFunc(marketsHandler.ResolveMarket))).Methods("POST")
	router.Handle("/v0/markets/{id}/leaderboard", securityMiddleware(http.HandlerFunc(marketsHandler.MarketLeaderboard))).Methods("GET")
	router.Handle("/v0/markets/{id}/projection", securityMiddleware(http.HandlerFunc(marketsHandler.ProjectProbability))).Methods("GET")

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
	router.Handle("/v0/marketprojection/{marketId}/{amount}/{outcome}", securityMiddleware(marketshandlers.ProjectNewProbabilityHandler(marketsService))).Methods("GET")
	router.Handle("/v0/marketprojection/{marketId}/{amount}/{outcome}/", securityMiddleware(marketshandlers.ProjectNewProbabilityHandler(marketsService))).Methods("GET")

	router.Handle("/v0/markets/bets/{marketId}", securityMiddleware(betshandlers.MarketBetsHandlerWithService(marketsService))).Methods("GET")
	router.Handle("/v0/markets/positions/{marketId}", securityMiddleware(positionshandlers.MarketPositionsHandlerWithService(marketsService))).Methods("GET")
	router.Handle("/v0/markets/positions/{marketId}/{username}", securityMiddleware(positionshandlers.MarketUserPositionHandlerWithService(marketsService))).Methods("GET")

	router.Handle("/v0/userinfo/{username}", securityMiddleware(usershandlers.GetPublicUserHandler(usersService))).Methods("GET")
	router.Handle("/v0/usercredit/{username}", securityMiddleware(usercredit.GetUserCreditHandler(usersService, econConfig.Economics.User.MaximumDebtAllowed))).Methods("GET")
	router.Handle("/v0/portfolio/{username}", securityMiddleware(publicuser.GetPortfolioHandler(usersService))).Methods("GET")
	router.Handle("/v0/users/{username}/financial", securityMiddleware(usershandlers.GetUserFinancialHandler(usersService))).Methods("GET")

	router.Handle("/v0/privateprofile", securityMiddleware(privateuser.GetPrivateProfileHandler(usersService))).Methods("GET")

	router.Handle("/v0/changepassword", securityMiddleware(usershandlers.ChangePasswordHandler(usersService))).Methods("POST")
	router.Handle("/v0/profilechange/displayname", securityMiddleware(usershandlers.ChangeDisplayNameHandler(usersService))).Methods("POST")
	router.Handle("/v0/profilechange/emoji", securityMiddleware(usershandlers.ChangeEmojiHandler(usersService))).Methods("POST")
	router.Handle("/v0/profilechange/description", securityMiddleware(usershandlers.ChangeDescriptionHandler(usersService))).Methods("POST")
	router.Handle("/v0/profilechange/links", securityMiddleware(usershandlers.ChangePersonalLinksHandler(usersService))).Methods("POST")

	router.Handle("/v0/bet", securityMiddleware(buybetshandlers.PlaceBetHandler(betsService, usersService))).Methods("POST")
	router.Handle("/v0/userposition/{marketId}", securityMiddleware(usershandlers.UserMarketPositionHandlerWithService(marketsService, usersService))).Methods("GET")
	router.Handle("/v0/sell", securityMiddleware(sellbetshandlers.SellPositionHandler(betsService, usersService))).Methods("POST")

	router.Handle("/v0/admin/createuser", securityMiddleware(http.HandlerFunc(adminhandlers.AddUserHandler(setup.EconomicsConfig, authService)))).Methods("POST")

	homepageRepo := homepage.NewGormRepository(db)
	homepageRenderer := homepage.NewDefaultRenderer()
	homepageSvc := homepage.NewService(homepageRepo, homepageRenderer)
	homepageHandler := cmshomehttp.NewHandler(homepageSvc, authService)

	router.HandleFunc("/v0/content/home", homepageHandler.PublicGet).Methods("GET")
	router.Handle("/v0/admin/content/home", securityMiddleware(http.HandlerFunc(homepageHandler.AdminUpdate))).Methods("PUT")

	handler := http.Handler(router)
	if c != nil {
		handler = c.Handler(handler)
	}
	return handler
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
	util.DB = db
	seedServerTestData(t, db)

	handler := buildTestHandler(t)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"health", "/health", http.StatusOK},
		{"home", "/v0/home", http.StatusOK},
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
