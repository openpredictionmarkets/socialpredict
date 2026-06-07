package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddReportingVisibilitySettings adds a singleton CMS setting that
// controls public visibility for aggregate reporting endpoints.
func MigrateAddReportingVisibilitySettings(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.ReportingVisibilitySettings{}); err != nil {
		return err
	}
	return db.Where("slug = ?", "default").FirstOrCreate(&models.ReportingVisibilitySettings{}, models.ReportingVisibilitySettings{
		Slug:                    "default",
		SystemMetricsPublic:     true,
		GlobalLeaderboardPublic: true,
		Version:                 1,
	}).Error
}

func init() {
	migration.Register("20260607092000", func(db *gorm.DB) error {
		return MigrateAddReportingVisibilitySettings(db)
	})
}
