package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"socialpredict/models"
	"socialpredict/setup"
	"socialpredict/util"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gorm.io/gorm"
)

// login and validation stuff
var jwtKey = []byte(os.Getenv("JWT_SIGNING_KEY"))

// UserClaims represents the expected structure of the JWT claims
type UserClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// PrivateUserResponse is a struct for user data that is safe to send to the client for login
type PrivateUserResponse struct {
	ID                    uint    `json:"id"`
	Username              string  `json:"username"`
	DisplayName           string  `json:"displayname" gorm:"unique;not null"`
	Email                 string  `json:"email"`
	UserType              string  `json:"usertype"`
	InitialAccountBalance float64 `json:"initialAccountBalance"`
	AccountBalance        float64 `json:"accountBalance"`
	ApiKey                string  `json:"apiKey,omitempty" gorm:"unique"`
	PersonalEmoji         string  `json:"personalEmoji,omitempty"`
	Description           string  `json:"description,omitempty"`
	PersonalLink1         string  `json:"personalink1,omitempty"`
	PersonalLink2         string  `json:"personalink2,omitempty"`
	PersonalLink3         string  `json:"personalink3,omitempty"`
	PersonalLink4         string  `json:"personalink4,omitempty"`
}

// PublicUserResponse is a struct for user data that is safe to send to the client for Profiles
type PublicUserResponse struct {
	Username              string  `json:"username"`
	DisplayName           string  `json:"displayname" gorm:"unique;not null"`
	UserType              string  `json:"usertype"`
	InitialAccountBalance float64 `json:"initialAccountBalance"`
	AccountBalance        float64 `json:"accountBalance"`
	PersonalEmoji         string  `json:"personalEmoji,omitempty"`
	Description           string  `json:"description,omitempty"`
	PersonalLink1         string  `json:"personalink1,omitempty"`
	PersonalLink2         string  `json:"personalink2,omitempty"`
	PersonalLink3         string  `json:"personalink3,omitempty"`
	PersonalLink4         string  `json:"personalink4,omitempty"`
}

type ProbabilityChange struct {
	Probability float64   `json:"probability"`
	Timestamp   time.Time `json:"timestamp"`
}

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
	router.HandleFunc("/v0/home", homeHandler)
	router.HandleFunc("/v0/login", loginHandler)
	router.HandleFunc("/v0/markets", marketsHandler)
	router.HandleFunc("/v0/create", createHandler)
	router.HandleFunc("/v0/markets/{marketId}", marketDetailsHandler).Methods("GET")
	router.HandleFunc("/v0/bet", betHandler).Methods("POST")
	// handle private user stuff, get private user info, update profile
	router.HandleFunc("/v0/user", userHandler)
	router.HandleFunc("/v0/resolve/{marketId}", resolveMarketHandler).Methods("POST")

	// router.HandleFunc("/v0/profilechange", updateUserProfile).Methods("POST")
	// handle public user stuff
	router.HandleFunc("/v0/userinfo/{username}", getPublicUserInfo).Methods("GET")

	// Apply the CORS middleware to the Gorilla Mux router
	handler := c.Handler(router) // Use the Gorilla Mux router here

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	// Set the Content-Type header to indicate a JSON response
	w.Header().Set("Content-Type", "application/json")

	// Send a JSON-formatted response
	fmt.Fprint(w, `{"message": "Data From the Backend!"}`)

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	// Parse the request body
	type loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Use database connection
	db := util.GetDB()

	// Find user by username
	var user models.User
	result := db.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Error accessing database", http.StatusInternalServerError)
		return
	}

	// Check password
	if !user.CheckPasswordHash(req.Password) {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}

	// Create UserClaim
	claims := &UserClaims{
		Username: user.Username, // Set this to the actual username
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
		},
	}

	// Create a new token object
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	log.Printf("Token issued for user: %s", user.Username)
	log.Printf("Tokenstring: %s", tokenString)

	// Send token and user ID in response
	responseData := map[string]interface{}{
		"token":    tokenString,
		"username": user.Username,
	}
	json.NewEncoder(w).Encode(responseData)
}

func marketsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	db := util.GetDB()

	var countMarketTotal int64
	countResult := db.Raw("SELECT COUNT(*) FROM markets;").Scan(&countMarketTotal)
	if countResult.Error != nil {
		// Handle error
		log.Printf("Error executing raw SQL query: %v", countResult.Error)
	} else {
		log.Printf("Total number of markets: %d", countMarketTotal)
	}

	var markets []models.Market
	result := db.Order("RANDOM()").Limit(int(countMarketTotal)).Find(&markets) // Adjust this line based on your ORM
	if result.Error != nil {
		http.Error(w, "Error fetching markets", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "application/json")

	// Encode the markets slice to JSON and send it as a response
	if err := json.NewEncoder(w).Encode(markets); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Use database connection
	db := util.GetDB()
	user, err := validateTokenAndGetUser(r, db)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var newMarket models.Market

	// Decode the request body into newMarket
	err = json.NewDecoder(r.Body).Decode(&newMarket)
	if err != nil {
		// Log the error and the request body for debugging
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		log.Printf("Error reading request body: %v, Body: %s", err, string(bodyBytes))

		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Validate the newMarket data as needed

	// Find the User by username
	result := db.Where("username = ?", newMarket.CreatorUsername).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Creator user not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error finding creator user", http.StatusInternalServerError)
		}
		return
	}

	// Create the market in the database
	result = db.Create(&newMarket)
	if result.Error != nil {
		log.Printf("Error creating new market: %v", result.Error)
		http.Error(w, "Error creating new market", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "application/json")

	// Send a success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newMarket)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not supported", http.StatusNotFound)
		return
	}

	// Extract userID from query parameters
	queryValues := r.URL.Query()
	userID, err := strconv.Atoi(queryValues.Get("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	db := util.GetDB()
	var user models.User

	// Fetch the user from the database
	result := db.First(&user, userID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error fetching user: %v", result.Error)
		http.Error(w, "Error fetching user", http.StatusInternalServerError)
		return
	}

	// Create a response object excluding sensitive data
	userResponse := PrivateUserResponse{
		ID:                    user.ID,
		Username:              user.Username,
		Email:                 user.Email,
		UserType:              user.UserType,
		InitialAccountBalance: user.InitialAccountBalance,
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "application/json")

	// Encode the userResponse to JSON and send it as a response
	if err := json.NewEncoder(w).Encode(userResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func marketDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketId := vars["marketId"]

	// Use database connection
	db := util.GetDB()

	var market models.Market
	// Use Preload to fetch the Creator along with the Market
	result := db.Preload("Creator").Where("ID = ?", marketId).First(&market)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Market not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching market", http.StatusInternalServerError)
		}
		return
	}

	// Parsing a String to an Unsigned Integer, base10, 32bits
	marketIDUint, err := strconv.ParseUint(marketId, 10, 32)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Fetch all bets for the market once
	bets, err := getBetsForMarket(marketIDUint)
	if err != nil {
		http.Error(w, "Error retrieving bets.", http.StatusInternalServerError)
		return
	}

	// Calculate probabilities using the fetched bets
	probabilityChanges, err := calculateMarketProbabilities(market, bets)
	if err != nil {
		http.Error(w, "Error calculating market probabilities", http.StatusInternalServerError)
		return
	}

	numUsers := getMarketUsers(bets)
	if err != nil {
		http.Error(w, "Error retrieving number of users.", http.StatusInternalServerError)
		return
	}

	// Inside your handler
	marketVolume := getMarketVolume(bets)
	if err != nil {
		// Handle error
	}

	// get market creator

	// Update your response struct accordingly
	response := struct {
		Market             models.Market       `json:"market"`
		CreatorUsername    string              `json:"creatorUsername"`
		ProbabilityChanges []ProbabilityChange `json:"probabilityChanges"`
		NumUsers           int                 `json:"numUsers"`
		TotalVolume        float64             `json:"totalVolume"`
	}{
		Market:             market,
		CreatorUsername:    market.Creator.Username, // Include the creator's username
		ProbabilityChanges: probabilityChanges,
		NumUsers:           numUsers,
		TotalVolume:        marketVolume,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func betHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	// Validate JWT token and extract user information
	db := util.GetDB() // Get the database connection
	user, err := validateTokenAndGetUser(r, db)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var betRequest models.Bet
	// Decode the request body into betRequest
	err = json.NewDecoder(r.Body).Decode(&betRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the request (check if user and market exist, if amount is positive, etc.)
	// ...

	// Fetch the market to check if it is resolved
	var market models.Market
	if result := db.First(&market, betRequest.MarketID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Market not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching market", http.StatusInternalServerError)
		}
		return
	}

	// Check if the market is already resolved
	if market.IsResolved {
		http.Error(w, "Cannot place a bet on a resolved market", http.StatusBadRequest)
		return
	}

	// user-specific validation, sufficient balance,
	// Fetch the user's current balance
	if err := db.First(&user, user.ID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the user has enough balance to place the bet

	// load the config constants
	config := setup.LoadEconomicsConfig()
	// Use the config as needed
	maximumDebtAllowed := config.Economics.User.MaximumDebtAllowed

	// Check if the user's balance after the bet would be lower than the allowed maximum debt
	if user.AccountBalance-betRequest.Amount < -maximumDebtAllowed {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	// Deduct the bet amount from the user's balance
	user.AccountBalance -= betRequest.Amount

	// Update the user's balance in the database
	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Error updating user balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new Bet object
	bet := models.Bet{
		Username: user.Username,
		MarketID: betRequest.MarketID,
		Amount:   betRequest.Amount,
		PlacedAt: time.Now(), // Set the current time as the placement time
		Outcome:  betRequest.Outcome,
	}

	// Save the Bet to the database
	result := db.Create(&bet)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Return a success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bet)
}

func getBetsForMarket(marketID uint64) ([]models.Bet, error) {
	var bets []models.Bet

	// Retrieve all bets for the market
	db := util.GetDB()
	if err := db.Where("market_id = ?", marketID).Find(&bets).Error; err != nil {
		return nil, err
	}

	return bets, nil
}

// Modify calculateMarketProbabilities to accept bets directly
func calculateMarketProbabilities(market models.Market, bets []models.Bet) ([]ProbabilityChange, error) {

	var probabilityChanges []ProbabilityChange

	// Initial state
	P_initial := market.InitialProbability // Assuming this is the initial probability
	I_initial := 10.0                      // You might want to make this a constant or part of the market struct
	totalYes := 0.0
	totalNo := 0.0

	// Add initial state
	probabilityChanges = append(probabilityChanges, ProbabilityChange{Probability: P_initial, Timestamp: market.CreatedAt})

	// Calculate probabilities after each bet
	for _, bet := range bets {
		if bet.Outcome == "YES" {
			totalYes += bet.Amount
		} else if bet.Outcome == "NO" {
			totalNo += bet.Amount
		}

		newProbability := (P_initial*I_initial + totalYes) / (I_initial + totalYes + totalNo)
		probabilityChanges = append(probabilityChanges, ProbabilityChange{Probability: newProbability, Timestamp: bet.PlacedAt})
	}

	return probabilityChanges, nil
}

// getMarketUsers returns the number of unique users for a given market
func getMarketUsers(bets []models.Bet) int {
	userMap := make(map[string]bool)
	for _, bet := range bets {
		userMap[bet.Username] = true
	}

	return len(userMap)
}

// getMarketVolume returns the total volume of trades for a given market
func getMarketVolume(bets []models.Bet) float64 {
	var totalVolume float64
	for _, bet := range bets {
		totalVolume += bet.Amount
	}

	return totalVolume
}

func getPublicUserInfo(w http.ResponseWriter, r *http.Request) {
	// Extract the username from the URL
	vars := mux.Vars(r)
	username := vars["username"]

	db := util.GetDB()
	var user models.User

	// Fetch user data from the database
	result := db.Where("Username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching user data", http.StatusInternalServerError)
		}
		return
	}

	// Convert to PublicUserResponse
	response := PublicUserResponse{
		Username:              user.Username,
		DisplayName:           user.DisplayName,
		UserType:              user.UserType,
		InitialAccountBalance: user.InitialAccountBalance,
		AccountBalance:        user.AccountBalance, // Added AccountBalance
		PersonalEmoji:         user.PersonalEmoji,
		Description:           user.Description,
		PersonalLink1:         user.PersonalLink1,
		PersonalLink2:         user.PersonalLink2,
		PersonalLink3:         user.PersonalLink3,
		PersonalLink4:         user.PersonalLink4,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func updateUserProfile(w http.ResponseWriter, r *http.Request) {
	// Authentication check (to be implemented based on your auth system)
	// ...

	// Assuming userID is obtained from the session after authentication
	// userID := getSessionUserID(r) // getSessionUserID is a placeholder function

	var updateData models.User
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//db := util.GetDB()

	// Update the user data
	//result := db.Model(&models.User{}).Where("ID = ?", userID).Updates(models.User{
	//	DisplayName:   updateData.DisplayName,
	//	Password:      updateData.Password, // Ensure this is hashed
	//	ApiKey:        updateData.ApiKey,   // Generate new API key if necessary
	//	PersonalEmoji: updateData.PersonalEmoji,
	//	Description:   updateData.Description,
	//	SocialMedia:   updateData.SocialMedia,
	//})

	//if result.Error != nil {
	//	http.Error(w, "Error updating user profile", http.StatusInternalServerError)
	//	return
	//}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Profile updated successfully"))
}

// validateTokenAndGetUser validates the JWT token and returns the user info
func validateTokenAndGetUser(r *http.Request, db *gorm.DB) (*models.User, error) {
	// Extract the token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("authorization header is required")
	}

	// Typically, the Authorization header is in the format "Bearer {token}"
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Define a function to handle token parsing and claims extraction
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Here, you would return your JWT signing key, used to validate the token
		return jwtKey, nil
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, keyFunc)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		log.Printf("claims.Username is %s", claims.Username)
		if claims.Username == "" {
			return nil, errors.New("username claim is empty")
		}
		log.Printf("Extracted username: %s", claims.Username)
		var user models.User
		result := db.Where("username = ?", claims.Username).First(&user)
		if result.Error != nil {
			// Format the error message as JSON
			return nil, fmt.Errorf(`{"error": "user not found"}`)
		}
		return &user, nil
	}

	return nil, errors.New("invalid token")
}

func generateSigningKey() string {
	key := make([]byte, 32) // 256 bits
	_, err := rand.Read(key)
	if err != nil {
		// handle error
	}
	return base64.StdEncoding.EncodeToString(key)
}

func resolveMarketHandler(w http.ResponseWriter, r *http.Request) {
	// Use database connection
	db := util.GetDB()

	// Retrieve marketId from URL parameters
	vars := mux.Vars(r)
	marketIdStr := vars["marketId"]
	marketId, err := strconv.ParseUint(marketIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Validate token and get user
	user, err := validateTokenAndGetUser(r, db)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Parse request body for resolution outcome
	var resolutionData struct {
		Outcome string `json:"outcome"`
	}
	if err := json.NewDecoder(r.Body).Decode(&resolutionData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Find the market by ID
	var market models.Market
	if result := db.First(&market, marketId); result.Error != nil {
		http.Error(w, "Market not found", http.StatusNotFound)
		return
	}

	// Check if the logged-in user is the creator of the market
	if market.CreatorUsername != user.Username {
		http.Error(w, "User is not the creator of the market", http.StatusUnauthorized)
		return
	}

	// Check if the market is already resolved
	if market.IsResolved {
		http.Error(w, "Market is already resolved", http.StatusBadRequest)
		return
	}

	// Validate the resolution outcome
	if resolutionData.Outcome != "YES" && resolutionData.Outcome != "NO" && resolutionData.Outcome != "N/A" {
		http.Error(w, "Invalid resolution outcome", http.StatusBadRequest)
		return
	}

	// Update the market with the resolution result
	market.IsResolved = true
	market.ResolutionResult = resolutionData.Outcome
	market.FinalResolutionDateTime = time.Now()

	// Handle payouts (if applicable)
	err = distributePayouts(&market, db)
	if err != nil {
		http.Error(w, "Error distributing payouts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save the market changes
	if err := db.Save(&market).Error; err != nil {
		http.Error(w, "Error saving market resolution: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send a response back
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Market resolved successfully"})
}

// distributePayouts handles the logic for calculating and distributing payouts
func distributePayouts(market *models.Market, db *gorm.DB) error {
	// Retrieve all bets associated with the market
	var bets []models.Bet
	if err := db.Where("market_id = ?", market.ID).Find(&bets).Error; err != nil {
		return err
	}

	// Initialize variables to calculate total amounts for each outcome
	var totalYes, totalNo float64

	// Determine the pool sizes for each outcome
	for _, bet := range bets {
		if bet.Outcome == "YES" {
			totalYes += bet.Amount
		} else if bet.Outcome == "NO" {
			totalNo += bet.Amount
		}
	}

	// Handle the N/A outcome by refunding all bets
	if market.ResolutionResult == "N/A" {
		for _, bet := range bets {
			if err := updateUserBalance(bet.Username, bet.Amount, db, "refund"); err != nil {
				return err
			}
		}
		return nil
	}

	// Calculate payouts based on CPMM for YES and NO outcomes
	for _, bet := range bets {
		if bet.Outcome == market.ResolutionResult {
			var payout, totalPoolForOutcome float64
			if market.ResolutionResult == "YES" {
				totalPoolForOutcome = totalYes
			} else {
				totalPoolForOutcome = totalNo
			}

			totalPool := totalYes + totalNo
			payout = (bet.Amount / totalPoolForOutcome) * totalPool

			if err := updateUserBalance(bet.Username, payout, db, "win"); err != nil {
				return err
			}
		}
	}

	return nil
}

// updateUserBalance updates the user's account balance for winnings or refunds
func updateUserBalance(username string, amount float64, db *gorm.DB, transactionType string) error {
	var user models.User

	// Retrieve the user from the database using the username
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return err
	}

	// Adjust the user's account balance
	switch transactionType {
	case "win":
		// Add the winning amount to the user's balance
		user.AccountBalance += amount
	case "refund":
		// Refund the bet amount to the user's balance
		user.AccountBalance += amount
	default:
		// Handle unknown transaction types if necessary
		return fmt.Errorf("unknown transaction type")
	}

	// Save the updated user record
	if err := db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}
