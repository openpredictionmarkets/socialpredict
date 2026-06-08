package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddMarketProposalAutoApproval adds the market proposal auto-approval
// flag to the singleton market governance settings table.
func MigrateAddMarketProposalAutoApproval(db *gorm.DB) error {
	return db.AutoMigrate(&models.MarketGovernanceSettings{})
}

func init() {
	migration.Register("20260607094000", func(db *gorm.DB) error {
		return MigrateAddMarketProposalAutoApproval(db)
	})
}
