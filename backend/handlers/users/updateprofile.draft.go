package usershandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/models"
)

func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	// Authentication check (to be implemented based on your auth system)
	// ...

	// Assuming userID is obtained from the session after authentication
	// userID := getSessionUserID(r) // getSessionUserID is a placeholder function

	var updateData models.User
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//db := util.GetDB()

	// Update the user data
	//result := db.Model(&models.User{}).Where("ID = ?", userID).Updates(models.User{
	//	DisplayName:   updateData.DisplayName,
	//	Password:      updateData.Password, // Ensure this is hashed
	//	ApiKey:        updateData.ApiKey,   // Generate new API key if necessary
	//	PersonalEmoji: updateData.PersonalEmoji,
	//	Description:   updateData.Description,
	//	SocialMedia:   updateData.SocialMedia,
	//})

	//if result.Error != nil {
	//	http.Error(w, "Error updating user profile", http.StatusInternalServerError)
	//	return
	//}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Profile updated successfully"))
}
