package migrations

import (
	"socialpredict/migration"

	"gorm.io/gorm"
)

type socialShareSettings20251124 struct {
	gorm.Model
	Slug               string
	SiteName           string
	DefaultDescription string
	DefaultImageURL    string
	ImageAlt           string
	Version            uint
	UpdatedBy          string
}

func (socialShareSettings20251124) TableName() string {
	return "social_share_settings"
}

func MigrateAddSocialShareSettings(db *gorm.DB) error {
	return db.AutoMigrate(&socialShareSettings20251124{})
}

func init() {
	migration.Register("20251124090000", func(db *gorm.DB) error {
		return MigrateAddSocialShareSettings(db)
	})
}
