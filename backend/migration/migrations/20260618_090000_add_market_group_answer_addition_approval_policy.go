package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddMarketGroupAnswerAdditionApprovalPolicy adds the global admin
// approval policy for later answer options on grouped binary markets.
func MigrateAddMarketGroupAnswerAdditionApprovalPolicy(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.MarketGovernanceSettings{}); err != nil {
		return err
	}
	return db.Model(&models.MarketGovernanceSettings{}).
		Where("market_group_answer_addition_approval_policy = '' OR market_group_answer_addition_approval_policy IS NULL").
		Update("market_group_answer_addition_approval_policy", gorm.Expr(
			"CASE WHEN auto_approve_market_group_answers THEN ? ELSE ? END",
			"auto",
			"moderator",
		)).Error
}

func init() {
	migration.Register("20260618090000", func(db *gorm.DB) error {
		return MigrateAddMarketGroupAnswerAdditionApprovalPolicy(db)
	})
}
