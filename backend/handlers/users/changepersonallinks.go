package usershandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/middleware"
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

	db := util.GetDB()
	user, err := middleware.ValidateTokenAndGetUser(r, db)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var request ChangePersonalLinksRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received links update: %+v", request)

	// Directly map request fields to user model fields
	user.PersonalLink1 = request.PersonalLink1
	user.PersonalLink2 = request.PersonalLink2
	user.PersonalLink3 = request.PersonalLink3
	user.PersonalLink4 = request.PersonalLink4

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
