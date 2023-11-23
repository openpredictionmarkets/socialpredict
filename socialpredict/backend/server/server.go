package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/cors"
)

var jwtKey = []byte("your_secret_key") // Use a secret key for JWT signing

func Start() {
	// CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"}, // Adjust the origin here
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})

	// Define endpoint handlers
	// Define the handler for the /api/v0/home endpoint
	http.HandleFunc("/api/v0/home", homeHandler)
	// Define the handler for the /api/v0/login endpoint
	http.HandleFunc("/api/v0/login", loginHandler)

	// Apply the CORS middleware to our top-level router
	handler := c.Handler(http.DefaultServeMux)

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

	if req.Username == "admin" && req.Password == "password" {
		// Create a new token object
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
			Issuer:    req.Username,
		})

		// Sign and get the complete encoded token as a string
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, "Error creating token", http.StatusInternalServerError)
			return
		}

		log.Printf("Token issued for user: %s", req.Username)

		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	} else {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
	}
}
