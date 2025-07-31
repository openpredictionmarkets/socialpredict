package usershandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/security"
	"socialpredict/util"
)

type ChangeDisplayNameRequest struct {
	DisplayName string `json:"displayName"`
}

func ChangeDisplayName(w http.ResponseWriter, r *http.Request) {
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

	var request ChangeDisplayNameRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate display name length and content
	if len(request.DisplayName) > 50 || len(request.DisplayName) < 1 {
		http.Error(w, "Display name must be between 1 and 50 characters", http.StatusBadRequest)
		return
	}

	// Sanitize the display name to prevent XSS
	sanitizedDisplayName, err := securityService.Sanitizer.SanitizeDisplayName(request.DisplayName)
	if err != nil {
		http.Error(w, "Invalid display name: "+err.Error(), http.StatusBadRequest)
		return
	}

	user.DisplayName = sanitizedDisplayName
	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Failed to update display name: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
