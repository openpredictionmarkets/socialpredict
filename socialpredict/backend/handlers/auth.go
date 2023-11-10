package handlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/models"
)

func Register(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = user.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Error while hashing password", http.StatusInternalServerError)
		return
	}
	// Here you would insert the user into the database
	// and then return a confirmation or the created user object
}

func Login(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Here you would authenticate the user (check if exists and password is correct)
	// then return a JWT token or session cookie
}
