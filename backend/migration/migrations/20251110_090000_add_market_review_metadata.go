package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

func MigrateAddMarketReviewMetadata(db *gorm.DB) error {
	m := db.Migrator()
	for _, column := range []string{"ApprovedBy", "ApprovedAt", "RejectedBy", "RejectedAt", "RejectionReason"} {
		if !m.HasColumn(&models.Market{}, column) {
			if err := m.AddColumn(&models.Market{}, column); err != nil {
				return err
			}
		}
	}
	return nil
}

func init() {
	migration.Register("20251110090000", func(db *gorm.DB) error {
		return MigrateAddMarketReviewMetadata(db)
	})
}
