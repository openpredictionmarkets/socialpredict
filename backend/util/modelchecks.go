package util

import (
	"errors"
	"socialpredict/models"

	"gorm.io/gorm"
)

func CheckUserIsReal(db *gorm.DB, username string) error {
	var user models.User
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("creator user not found")
		}
		return result.Error
	}
	return nil
}
