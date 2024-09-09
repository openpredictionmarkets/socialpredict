package server

import (
	"log"
	"net/http"
	"socialpredict/handlers"
	adminhandlers "socialpredict/handlers/admin"
	betshandlers "socialpredict/handlers/bets"
	marketshandlers "socialpredict/handlers/markets"
	positions "socialpredict/handlers/positions"
	setuphandlers "socialpredict/handlers/setup"
	statshandlers "socialpredict/handlers/stats"
	usershandlers "socialpredict/handlers/users"
	"socialpredict/middleware"
	"socialpredict/setup"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func Start() {
	// CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:8089",
		},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Initialize mux router
	router := mux.NewRouter()

	// Define endpoint handlers using Gorilla Mux router
	// This defines all functions starting with /api/
	router.HandleFunc("/v0/home", handlers.HomeHandler)
	router.HandleFunc("/v0/login", middleware.LoginHandler)

	// application setup and stats information
	router.HandleFunc("/v0/setup", setuphandlers.GetSetupHandler(setup.LoadEconomicsConfig)).Methods("GET")
	router.HandleFunc("/v0/stats", statshandlers.StatsHandler()).Methods("GET")
	// markets display, market information
	router.HandleFunc("/v0/markets", marketshandlers.ListMarketsHandler).Methods("GET")
	router.HandleFunc("/v0/markets/{marketId}", marketshandlers.MarketDetailsHandler).Methods("GET")
	router.HandleFunc("/v0/marketprojection/{marketId}/{amount}/{outcome}/", marketshandlers.ProjectNewProbabilityHandler).Methods("GET")

	// handle market positions, get trades
	router.HandleFunc("/v0/markets/bets/{marketId}", betshandlers.MarketBetsDisplayHandler).Methods("GET")
	router.HandleFunc("/v0/markets/positions/{marketId}", positions.MarketDBPMPositionsHandler).Methods("GET")
	router.HandleFunc("/v0/markets/positions/{marketId}/{username}", positions.MarketDBPMUserPositionsHandler).Methods("GET")
	// show comments on markets

	// handle public user stuff
	router.HandleFunc("/v0/userinfo/{username}", usershandlers.GetPublicUserResponse).Methods("GET")
	router.HandleFunc("/v0/usercredit/{username}", usershandlers.GetUserCreditHandler).Methods("GET")
	// user portfolio, (which is public)
	router.HandleFunc("/v0/portfolio/{username}", usershandlers.GetPublicUserPortfolio).Methods("GET")

	// handle private user stuff, display sensitive profile information to customize
	router.HandleFunc("/v0/privateprofile", usershandlers.GetPrivateProfileUserResponse)
	// changing profile stuff
	router.HandleFunc("/v0/changepassword", usershandlers.ChangePassword).Methods("POST")
	router.HandleFunc("/v0/profilechange/displayname", usershandlers.ChangeDisplayName).Methods("POST")
	router.HandleFunc("/v0/profilechange/emoji", usershandlers.ChangeEmoji).Methods("POST")
	router.HandleFunc("/v0/profilechange/description", usershandlers.ChangeDescription).Methods("POST")
	router.HandleFunc("/v0/profilechange/links", usershandlers.ChangePersonalLinks).Methods("POST")

	// handle private user actions such as resolve a market, make a bet, create a market, change profile
	router.HandleFunc("/v0/resolve/{marketId}", marketshandlers.ResolveMarketHandler).Methods("POST")
	router.HandleFunc("/v0/bet", betshandlers.PlaceBetHandler).Methods("POST")
	router.HandleFunc("/v0/userposition/{marketId}", usershandlers.UserMarketPositionHandler)
	router.HandleFunc("/v0/sell", betshandlers.SellPositionHandler).Methods("POST")
	router.HandleFunc("/v0/create", marketshandlers.CreateMarketHandler).Methods("POST")

	// admin stuff
	router.HandleFunc("/v0/admin/createuser", adminhandlers.AddUserHandler(setup.EconomicsConfig)).Methods("POST")

	// Apply the CORS middleware to the Gorilla Mux router
	handler := c.Handler(router) // Use the Gorilla Mux router here

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
