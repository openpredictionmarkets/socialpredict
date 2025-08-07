package server

import (
	"log"
	"net/http"
	"socialpredict/handlers"
	adminhandlers "socialpredict/handlers/admin"
	betshandlers "socialpredict/handlers/bets"
	buybetshandlers "socialpredict/handlers/bets/buying"
	sellbetshandlers "socialpredict/handlers/bets/selling"
	marketshandlers "socialpredict/handlers/markets"
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

func Start() {
	// Initialize security service
	securityService := security.NewSecurityService()

	// CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:8089",
			"http://localhost",
		},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

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

	// Apply the CORS middleware to the Gorilla Mux router
	handler := c.Handler(router) // Use the Gorilla Mux router here

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
