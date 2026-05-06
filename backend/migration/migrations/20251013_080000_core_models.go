package migrations

import (
	"os"

	"socialpredict/logger"
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

func init() {
	err := migration.Register("20251013080000", func(db *gorm.DB) error {
		// Migrate the User models first
		if err := db.AutoMigrate(&models.User{}); err != nil {
			return err
		}

		// Then, migrate the Market model
		if err := db.AutoMigrate(&models.Market{}); err != nil {
			return err
		}

		// Then, migrate the Bet model
		if err := db.AutoMigrate(&models.Bet{}); err != nil {
			return err
		}

		// Then, migrate the HomepageContent model
		if err := db.AutoMigrate(&models.HomepageContent{}); err != nil {
			return err
		}

		return nil
	})

	// In init() functions, registration failure is a critical startup error
	if err != nil {
		logger.LogError("MigrationRegistration", "init", err)
		os.Exit(1)
	}
}
