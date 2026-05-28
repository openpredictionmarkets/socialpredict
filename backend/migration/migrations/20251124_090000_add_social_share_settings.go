package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

func MigrateAddSocialShareSettings(db *gorm.DB) error {
	return db.AutoMigrate(&models.SocialShareSettings{})
}

func init() {
	migration.Register("20251124090000", func(db *gorm.DB) error {
		return MigrateAddSocialShareSettings(db)
	})
}
