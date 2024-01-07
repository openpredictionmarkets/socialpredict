package handlers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"

	"gorm.io/gorm"
)

func CreateMarketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Use database connection
	db := util.GetDB()
	user, err := middleware.ValidateTokenAndGetUser(r, db)
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
