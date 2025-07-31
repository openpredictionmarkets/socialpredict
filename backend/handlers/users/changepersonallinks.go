package usershandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/security"
	"socialpredict/util"
)

type ChangePersonalLinksRequest struct {
	PersonalLink1 string `json:"personalLink1"`
	PersonalLink2 string `json:"personalLink2"`
	PersonalLink3 string `json:"personalLink3"`
	PersonalLink4 string `json:"personalLink4"`
}

func ChangePersonalLinks(w http.ResponseWriter, r *http.Request) {
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

	var request ChangePersonalLinksRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received links update: %+v", request)

	// Validate and sanitize each personal link individually
	links := [4]string{request.PersonalLink1, request.PersonalLink2, request.PersonalLink3, request.PersonalLink4}
	var sanitizedLinks [4]string

	for i, link := range links {
		// Allow empty links
		if link == "" {
			sanitizedLinks[i] = ""
			continue
		}

		// Validate link length
		if len(link) > 200 {
			http.Error(w, "Personal link exceeds maximum length of 200 characters", http.StatusBadRequest)
			return
		}

		// Sanitize the link
		sanitizedLink, err := securityService.Sanitizer.SanitizePersonalLink(link)
		if err != nil {
			http.Error(w, "Invalid personal link: "+err.Error(), http.StatusBadRequest)
			return
		}
		sanitizedLinks[i] = sanitizedLink
	}

	// Update user with sanitized links
	user.PersonalLink1 = sanitizedLinks[0]
	user.PersonalLink2 = sanitizedLinks[1]
	user.PersonalLink3 = sanitizedLinks[2]
	user.PersonalLink4 = sanitizedLinks[3]

	// Use direct update with GORM to specify which fields to update
	if err := db.Model(&user).Select("PersonalLink1", "PersonalLink2", "PersonalLink3", "PersonalLink4").Updates(map[string]interface{}{
		"PersonalLink1": user.PersonalLink1,
		"PersonalLink2": user.PersonalLink2,
		"PersonalLink3": user.PersonalLink3,
		"PersonalLink4": user.PersonalLink4,
	}).Error; err != nil {
		http.Error(w, "Failed to update personal links: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
