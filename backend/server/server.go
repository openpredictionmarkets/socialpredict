package server

import (
	"log"
	"net/http"
	"socialpredict/handlers"
	betshandlers "socialpredict/handlers/bets"
	marketshandlers "socialpredict/handlers/markets"
	"socialpredict/handlers/positions"
	usershandlers "socialpredict/handlers/users"
	"socialpredict/middleware"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func Start() {
	// CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://172.29.0.10:5173/", "https://brierfoxforecast.ngrok.app", "http://localhost:8089"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	// Initialize mux router
	router := mux.NewRouter()

	// Define endpoint handlers using Gorilla Mux router
	// This defines all functions starting with /api/
	router.HandleFunc("/v0/home", handlers.HomeHandler)
	router.HandleFunc("/v0/login", middleware.LoginHandler)

	// markets display, market information
	router.HandleFunc("/v0/markets", marketshandlers.ListMarketsHandler)
	router.HandleFunc("/v0/markets/{marketId}", marketshandlers.MarketDetailsHandler).Methods("GET")
	// handle market positions, get trades
	router.HandleFunc("/v0/markets/bets/{marketId}", betshandlers.MarketBetsDisplayHandler).Methods("GET")
	router.HandleFunc("/v0/markets/positions/{marketId}", positions.MarketDBPMPositionsHandler).Methods("GET")
	router.HandleFunc("/v0/markets/positions/{marketId}/{username}", positions.MarketDBPMUserPositionsHandler).Methods("GET")
	// show comments on markets

	// handle public user stuff
	router.HandleFunc("/v0/userinfo/{username}", usershandlers.GetPublicUserResponse).Methods("GET")
	// user portfolio, (which is public)
	router.HandleFunc("/v0/portfolio/{username}", usershandlers.GetPublicUserPortfolio).Methods("GET")

	// handle private user stuff, display sensitive profile information to customize
	router.HandleFunc("/v0/user/privateprofile", usershandlers.GetPrivateProfileUserResponse)
	// router.HandleFunc("/v0/profilechange", updateUserProfile).Methods("POST")

	// handle private user actions such as resolve a market, make a bet, create a market, change profile
	router.HandleFunc("/v0/resolve/{marketId}", marketshandlers.ResolveMarketHandler).Methods("POST")
	router.HandleFunc("/v0/bet", betshandlers.PlaceBetHandler).Methods("POST")
	router.HandleFunc("/v0/userposition", usershandlers.UserMarketPositionHandler)
	router.HandleFunc("/v0/sell", betshandlers.SellPositionHandler).Methods("POST")
	router.HandleFunc("/v0/create", marketshandlers.CreateMarketHandler)

	// Apply the CORS middleware to the Gorilla Mux router
	handler := c.Handler(router) // Use the Gorilla Mux router here

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
