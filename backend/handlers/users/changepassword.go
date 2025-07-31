package usershandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/security"
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

	// Initialize security service
	securityService := security.NewSecurityService()

	db := util.GetDB()
	user, httperr := middleware.ValidateTokenAndGetUser(r, db)
	if httperr != nil {
		http.Error(w, "Invalid token: "+httperr.Error(), http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	// Validate input fields
	if req.CurrentPassword == "" {
		http.Error(w, "Current password is required", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		http.Error(w, "New password is required", http.StatusBadRequest)
		return
	}

	// Check if the current password is correct
	if !user.CheckPasswordHash(req.CurrentPassword) {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	// Validate new password strength
	if err := securityService.Sanitizer.SanitizePassword(req.NewPassword); err != nil {
		http.Error(w, "New password does not meet security requirements: "+err.Error(), http.StatusBadRequest)
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
