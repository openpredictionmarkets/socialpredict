package usershandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/security"
	"socialpredict/util"
)

type ChangeDescriptionRequest struct {
	Description string `json:"description"`
}

func ChangeDescription(w http.ResponseWriter, r *http.Request) {
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

	var request ChangeDescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate description length and content
	if len(request.Description) > 2000 {
		http.Error(w, "Description exceeds maximum length of 2000 characters", http.StatusBadRequest)
		return
	}

	// Sanitize the description to prevent XSS
	sanitizedDescription, err := securityService.Sanitizer.SanitizeDescription(request.Description)
	if err != nil {
		http.Error(w, "Invalid description: "+err.Error(), http.StatusBadRequest)
		return
	}

	user.Description = sanitizedDescription
	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Failed to update description: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
