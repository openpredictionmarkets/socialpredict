package server

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"socialpredict/handlers"
	adminhandlers "socialpredict/handlers/admin"
	betshandlers "socialpredict/handlers/bets"
	buybetshandlers "socialpredict/handlers/bets/buying"
	sellbetshandlers "socialpredict/handlers/bets/selling"
	marketshandlers "socialpredict/handlers/markets"
	metricshandlers "socialpredict/handlers/metrics"
	positions "socialpredict/handlers/positions"
	setuphandlers "socialpredict/handlers/setup"
	statshandlers "socialpredict/handlers/stats"
	usershandlers "socialpredict/handlers/users"
	usercredit "socialpredict/handlers/users/credit"
	privateuser "socialpredict/handlers/users/privateuser"
	"socialpredict/handlers/users/publicuser"
	"socialpredict/middleware"
	"socialpredict/security"
	"socialpredict/setup"

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

	// Define endpoint handlers using Gorilla Mux router
	// This defines all functions starting with /api/

	// Apply security middleware to all routes
	securityMiddleware := securityService.SecurityMiddleware()
	loginSecurityMiddleware := securityService.LoginSecurityMiddleware()

	router.HandleFunc("/v0/home", handlers.HomeHandler).Methods("GET")
	router.Handle("/v0/login", loginSecurityMiddleware(http.HandlerFunc(middleware.LoginHandler))).Methods("POST")

	// application setup and stats information
	router.Handle("/v0/setup", securityMiddleware(http.HandlerFunc(setuphandlers.GetSetupHandler(setup.LoadEconomicsConfig)))).Methods("GET")
	router.Handle("/v0/stats", securityMiddleware(http.HandlerFunc(statshandlers.StatsHandler()))).Methods("GET")
	router.Handle("/v0/system/metrics", securityMiddleware(http.HandlerFunc(metricshandlers.GetSystemMetricsHandler))).Methods("GET")
	router.Handle("/v0/global/leaderboard", securityMiddleware(http.HandlerFunc(metricshandlers.GetGlobalLeaderboardHandler))).Methods("GET")

	// markets display, market information
	router.Handle("/v0/markets", securityMiddleware(http.HandlerFunc(marketshandlers.ListMarketsHandler))).Methods("GET")
	router.Handle("/v0/markets/search", securityMiddleware(http.HandlerFunc(marketshandlers.SearchMarketsHandler))).Methods("GET")
	router.Handle("/v0/markets/active", securityMiddleware(http.HandlerFunc(marketshandlers.ListActiveMarketsHandler))).Methods("GET")
	router.Handle("/v0/markets/closed", securityMiddleware(http.HandlerFunc(marketshandlers.ListClosedMarketsHandler))).Methods("GET")
	router.Handle("/v0/markets/resolved", securityMiddleware(http.HandlerFunc(marketshandlers.ListResolvedMarketsHandler))).Methods("GET")
	router.Handle("/v0/markets/{marketId}", securityMiddleware(http.HandlerFunc(marketshandlers.MarketDetailsHandler))).Methods("GET")
	router.Handle("/v0/marketprojection/{marketId}/{amount}/{outcome}/", securityMiddleware(http.HandlerFunc(marketshandlers.ProjectNewProbabilityHandler))).Methods("GET")

	// handle market positions, get trades
	router.Handle("/v0/markets/bets/{marketId}", securityMiddleware(http.HandlerFunc(betshandlers.MarketBetsDisplayHandler))).Methods("GET")
	router.Handle("/v0/markets/positions/{marketId}", securityMiddleware(http.HandlerFunc(positions.MarketDBPMPositionsHandler))).Methods("GET")
	router.Handle("/v0/markets/positions/{marketId}/{username}", securityMiddleware(http.HandlerFunc(positions.MarketDBPMUserPositionsHandler))).Methods("GET")
	router.Handle("/v0/markets/leaderboard/{marketId}", securityMiddleware(http.HandlerFunc(marketshandlers.MarketLeaderboardHandler))).Methods("GET")

	// handle public user stuff
	router.Handle("/v0/userinfo/{username}", securityMiddleware(http.HandlerFunc(publicuser.GetPublicUserResponse))).Methods("GET")
	router.Handle("/v0/usercredit/{username}", securityMiddleware(http.HandlerFunc(usercredit.GetUserCreditHandler))).Methods("GET")
	router.Handle("/v0/portfolio/{username}", securityMiddleware(http.HandlerFunc(publicuser.GetPortfolio))).Methods("GET")
	router.Handle("/v0/users/{username}/financial", securityMiddleware(http.HandlerFunc(usershandlers.GetUserFinancialHandler))).Methods("GET")

	// handle private user stuff, display sensitive profile information to customize
	router.Handle("/v0/privateprofile", securityMiddleware(http.HandlerFunc(privateuser.GetPrivateProfileUserResponse))).Methods("GET")

	// changing profile stuff - apply security middleware
	router.Handle("/v0/changepassword", securityMiddleware(http.HandlerFunc(usershandlers.ChangePassword))).Methods("POST")
	router.Handle("/v0/profilechange/displayname", securityMiddleware(http.HandlerFunc(usershandlers.ChangeDisplayName))).Methods("POST")
	router.Handle("/v0/profilechange/emoji", securityMiddleware(http.HandlerFunc(usershandlers.ChangeEmoji))).Methods("POST")
	router.Handle("/v0/profilechange/description", securityMiddleware(http.HandlerFunc(usershandlers.ChangeDescription))).Methods("POST")
	router.Handle("/v0/profilechange/links", securityMiddleware(http.HandlerFunc(usershandlers.ChangePersonalLinks))).Methods("POST")

	// handle private user actions such as resolve a market, make a bet, create a market, change profile
	router.Handle("/v0/resolve/{marketId}", securityMiddleware(http.HandlerFunc(marketshandlers.ResolveMarketHandler))).Methods("POST")
	router.Handle("/v0/bet", securityMiddleware(http.HandlerFunc(buybetshandlers.PlaceBetHandler(setup.EconomicsConfig)))).Methods("POST")
	router.Handle("/v0/userposition/{marketId}", securityMiddleware(http.HandlerFunc(usershandlers.UserMarketPositionHandler))).Methods("GET")
	router.Handle("/v0/sell", securityMiddleware(http.HandlerFunc(sellbetshandlers.SellPositionHandler(setup.EconomicsConfig)))).Methods("POST")
	router.Handle("/v0/create", securityMiddleware(http.HandlerFunc(marketshandlers.CreateMarketHandler(setup.EconomicsConfig)))).Methods("POST")

	// admin stuff - apply security middleware
	router.Handle("/v0/admin/createuser", securityMiddleware(http.HandlerFunc(adminhandlers.AddUserHandler(setup.EconomicsConfig)))).Methods("POST")

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
