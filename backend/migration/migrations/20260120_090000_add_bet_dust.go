package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddBetDust adds exact sale dust metadata to bets.
func MigrateAddBetDust(db *gorm.DB) error {
	return db.AutoMigrate(&models.Bet{})
}

func init() {
	migration.Register("20260120090000", func(db *gorm.DB) error {
		return MigrateAddBetDust(db)
	})
}
