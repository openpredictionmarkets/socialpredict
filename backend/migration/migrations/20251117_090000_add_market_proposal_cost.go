package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

func MigrateAddMarketProposalCost(db *gorm.DB) error {
	m := db.Migrator()
	if !m.HasColumn(&models.Market{}, "ProposalCost") {
		if err := m.AddColumn(&models.Market{}, "ProposalCost"); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	migration.Register("20251117090000", func(db *gorm.DB) error {
		return MigrateAddMarketProposalCost(db)
	})
}
