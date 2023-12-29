package util

import (
	"github.com/google/uuid"
)

func GenerateUniqueApiKey() string {
	return uuid.New().String()
}
