package usershandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/util"
)

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	db := util.GetDB()
	user, httperr := middleware.ValidateTokenAndGetUser(r, db)
	if httperr != nil {
		http.Error(w, httperr.Error(), httperr.StatusCode)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	// Check if the current password is correct
	if !user.CheckPasswordHash(req.CurrentPassword) {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	// Hash the new password
	if err := user.HashPassword(req.NewPassword); err != nil {
		http.Error(w, "Failed to hash new password", http.StatusInternalServerError)
		return
	}

	// Set MustChangePassword to false
	user.MustChangePassword = false

	// Update the password and MustChangePassword in the database
	if result := db.Save(&user); result.Error != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	// Send a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Password changed successfully"))
}
