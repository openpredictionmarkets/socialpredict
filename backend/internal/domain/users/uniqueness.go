package users

import (
	"errors"
	"fmt"

	"socialpredict/models"

	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
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

func CountByField(db *gorm.DB, field, value string) int64 {
	var count int64
	db.Model(&models.User{}).Where(fmt.Sprintf("%s = ?", field), value).Count(&count)
	return count
}

func UniqueDisplayName(db *gorm.DB) string {
	for {
		name := gofakeit.Name()
		if count := CountByField(db, "display_name", name); count == 0 {
			return name
		}
	}
}

func UniqueEmail(db *gorm.DB) string {
	for {
		email := gofakeit.Email()
		if count := CountByField(db, "email", email); count == 0 {
			return email
		}
	}
}

func GenerateUniqueAPIKey(db *gorm.DB) string {
	for {
		apiKey := uuid.NewString()
		if count := CountByField(db, "api_key", apiKey); count == 0 {
			return apiKey
		}
	}
}
