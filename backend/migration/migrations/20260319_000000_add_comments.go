package migrations

import (
	"log"

	"socialpredict/logger"
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

func init() {
	err := migration.Register("20260319000000", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Comment{})
	})
	if err != nil {
		logger.LogError("migrations", "init", err)
		log.Fatalf("Failed to register migration 20260319000000: %v", err)
	}
}
