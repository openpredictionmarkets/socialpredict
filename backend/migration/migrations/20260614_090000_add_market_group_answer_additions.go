package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddMarketGroupAnswerAdditions adds governed answer-addition storage
// for multiple-choice binary market groups.
func MigrateAddMarketGroupAnswerAdditions(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.MarketGroupAnswerAddition{},
		&models.MarketGovernanceSettings{},
	)
}

func init() {
	migration.Register("20260614090000", func(db *gorm.DB) error {
		return MigrateAddMarketGroupAnswerAdditions(db)
	})
}
