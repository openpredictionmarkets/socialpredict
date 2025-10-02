package usershandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/logger"
	"socialpredict/middleware"
	"socialpredict/security"
	"socialpredict/util"

	"fmt"
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

	logger.LogInfo("ChangePassword", "ChangePassword", "ChangePassword handler called")

	// Initialize security service
	securityService := security.NewSecurityService()

	db := util.GetDB()
	user, httperr := middleware.ValidateTokenAndGetUser(r, db)
	if httperr != nil {
		http.Error(w, "Invalid token: "+httperr.Error(), http.StatusUnauthorized)
		logger.LogError("ChangePassword", "ValidateTokenAndGetUser", httperr)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		logger.LogError("ChangePassword", "DecodeRequestBody", err)
		return
	}

	// Validate input fields
	if req.CurrentPassword == "" {
		http.Error(w, "Current password is required", http.StatusBadRequest)
		logger.LogError("ChangePassword", "ValidateInputFields", fmt.Errorf("Current password is required"))
		return
	}

	if req.NewPassword == "" {
		http.Error(w, "New password is required", http.StatusBadRequest)
		logger.LogError("ChangePassword", "ValidateInputFields", fmt.Errorf("New password is required"))
		return
	}

	// Check if the current password is correct
	if !user.CheckPasswordHash(req.CurrentPassword) {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		logger.LogError("ChangePassword", "CheckPasswordHash", fmt.Errorf("Current password is incorrect"))
		return
	}

	// Validate new password strength
	if _, err := securityService.Sanitizer.SanitizePassword(req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		// http.Error(w, "New password does not meet security requirements: "+err.Error(), http.StatusBadRequest)
		logger.LogError("ChangePassword", "ValidateNewPasswordStrength", err)
		return
	}

	// Hash the new password
	if err := user.HashPassword(req.NewPassword); err != nil {
		http.Error(w, "Failed to hash new password", http.StatusInternalServerError)
		logger.LogError("ChangePassword", "HashNewPassword", err)
		return
	}

	// Set MustChangePassword to false
	user.MustChangePassword = false

	// Update the password and MustChangePassword in the database
	if result := db.Save(&user); result.Error != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		logger.LogError("ChangePassword", "UpdatePasswordInDB", result.Error)
		return
	}

	// Send a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Password changed successfully"))
	logger.LogInfo("ChangePassword", "ChangePassword", "Password changed successfully for user "+user.Username)
}
