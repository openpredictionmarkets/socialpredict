package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddMarketGroupAnswerAdditionPolicy adds per-group answer-addition
// auto-approval policy for multiple-choice binary market groups.
func MigrateAddMarketGroupAnswerAdditionPolicy(db *gorm.DB) error {
	return db.AutoMigrate(&models.MarketGroup{})
}

func init() {
	migration.Register("20260614100000", func(db *gorm.DB) error {
		return MigrateAddMarketGroupAnswerAdditionPolicy(db)
	})
}
