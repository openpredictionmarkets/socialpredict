package migration

import (
	"log"
	"socialpredict/models"

	"gorm.io/gorm"
)

func MigrateDB(db *gorm.DB) {
	// Migrate the User modelsfirst
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("Error migrating User model: %v", err)
	}

	// Then, migrate the Market model
	err = db.AutoMigrate(&models.Market{})
	if err != nil {
		log.Fatalf("Error migrating Market model: %v", err)
	}

	// Then, migrate the Bet model
	err = db.AutoMigrate(&models.Bet{})
	if err != nil {
		log.Fatalf("Error migrating Bet model: %v", err)
	}
}
