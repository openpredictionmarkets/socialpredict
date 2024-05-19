package usershandlers

import (
	"encoding/json"
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

	user.PersonalLink1 = request.PersonalLink1
	user.PersonalLink2 = request.PersonalLink2
	user.PersonalLink3 = request.PersonalLink3
	user.PersonalLink4 = request.PersonalLink4
	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Failed to update personal links: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
