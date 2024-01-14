package server

import (
	"log"
	"net/http"
	"socialpredict/handlers"
	betsHandlers "socialpredict/handlers/bets"
	marketsHandlers "socialpredict/handlers/markets"
	usersHandlers "socialpredict/handlers/users"
	"socialpredict/middleware"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func Start() {
	// CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "https://brierfoxforecast.ngrok.app", "http://localhost:8089"},
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
	router.HandleFunc("/v0/markets", marketsHandlers.ListMarketsHandler)
	router.HandleFunc("/v0/markets/{marketId}", marketsHandlers.MarketDetailsHandler).Methods("GET")
	// handle market positions, get trades
	router.HandleFunc("/v0/markets/bets/{marketId}", betsHandlers.MarketBetsDisplayHandler).Methods("GET")
	// router.HandleFunc("/v0/markets/positions/{marketId}", marketPositionsHandler).Methods("GET")
	// show comments on markets

	// handle public user stuff
	router.HandleFunc("/v0/userinfo/{username}", usersHandlers.GetPublicUserResponse).Methods("GET")
	// user portfolio, (which is public)
	// router.HandleFunc("v0/portfolio/{username}", userHandlers.GetPortfolio).Methods("GET")

	// handle private user stuff, display sensitive profile information to customize
	router.HandleFunc("/v0/user/privateprofile", usersHandlers.GetPrivateProfileUserResponse)
	// router.HandleFunc("/v0/profilechange", updateUserProfile).Methods("POST")

	// handle private user actions such as resolve a market, make a bet, create a market, change profile
	router.HandleFunc("/v0/resolve/{marketId}", marketsHandlers.ResolveMarketHandler).Methods("POST")
	router.HandleFunc("/v0/bet", betsHandlers.PlaceBetHandler).Methods("POST")
	router.HandleFunc("/v0/create", marketsHandlers.CreateMarketHandler)

	// Apply the CORS middleware to the Gorilla Mux router
	handler := c.Handler(router) // Use the Gorilla Mux router here

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
