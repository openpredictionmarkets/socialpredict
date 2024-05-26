package util

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GenerateUniqueApiKey(db *gorm.DB) string {
	for {
		apiKey := uuid.NewString()
		if count := CountByField(db, "api_key", apiKey); count == 0 {
			return apiKey
		}
	}
}
