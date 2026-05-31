package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

func MigrateAddSocialShareImageEnabled(db *gorm.DB) error {
	m := db.Migrator()
	added := false
	if !m.HasColumn(&models.SocialShareSettings{}, "ImageEnabled") {
		if err := m.AddColumn(&models.SocialShareSettings{}, "ImageEnabled"); err != nil {
			return err
		}
		added = true
	}
	if added {
		return db.Model(&models.SocialShareSettings{}).Where("1 = 1").Update("image_enabled", true).Error
	}
	return nil
}

func init() {
	migration.Register("20251201090000", func(db *gorm.DB) error {
		return MigrateAddSocialShareImageEnabled(db)
	})
}
