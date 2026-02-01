package migrations

import (
	"log"

	"socialpredict/logger"
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

func init() {
	err := migration.Register("20260201120000", func(db *gorm.DB) error {
		// Create the ledger_entries table for wallet audit trail
		if err := db.AutoMigrate(&models.LedgerEntry{}); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logger.LogError("migrations", "init", err)
		log.Fatalf("Failed to register migration 20260201120000: %v", err)
	}
}
