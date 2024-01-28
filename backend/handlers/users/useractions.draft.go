package usershandlers

import (
	"socialpredict/models"

	"gorm.io/gorm"
)

func UpdateUserDisplayName(db *gorm.DB, username, newDisplayName string) error {
	var user models.User
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return result.Error
	}

	user.DisplayName = newDisplayName
	return db.Save(&user).Error
}

func UpdateUserPersonalEmoji(db *gorm.DB, username, newEmoji string) error {
	var user models.User
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return result.Error
	}

	user.PersonalEmoji = newEmoji
	return db.Save(&user).Error
}

func UpdateUserDescription(db *gorm.DB, username, newDescription string) error {
	var user models.User
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return result.Error
	}

	user.Description = newDescription
	return db.Save(&user).Error
}

func UpdateUserPersonalLinks(db *gorm.DB, username string, newLinks [4]string) error {
	var user models.User
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return result.Error
	}

	user.PersonalLink1 = newLinks[0]
	user.PersonalLink2 = newLinks[1]
	user.PersonalLink3 = newLinks[2]
	user.PersonalLink4 = newLinks[3]
	return db.Save(&user).Error
}

func UpdateUserEmail(db *gorm.DB, username, newEmail string) error {
	var user models.User
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return result.Error
	}

	// Assume some encoding mechanism
	// encodedEmail := EncodeEmail(newEmail)
	// user.Email = encodedEmail
	return db.Save(&user).Error
}

//func UpdateUserAPIKey(db *gorm.DB, username string) (string, error) {
//	var user models.User
//	result := db.Where("username = ?", username).First(&user)
//	if result.Error != nil {
//		return "", result.Error
//	}

// newAPIKey := GenerateRandomAPIKey()      // Implement this function
//  encodedAPIKey := EncodeAPIKey(newAPIKey) // Implement this function
// user.ApiKey = encodedAPIKey
// err := db.Save(&user).Error
// return newAPIKey, err // Return the raw API key for one-time display
//}
