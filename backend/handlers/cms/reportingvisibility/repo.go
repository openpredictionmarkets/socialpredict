package reportingvisibility

import (
	"socialpredict/models"

	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) GetSettings() (*models.ReportingVisibilitySettings, error) {
	var item models.ReportingVisibilitySettings
	if err := r.db.Where("slug = ?", SettingsSlugDefault).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) SaveSettings(item *models.ReportingVisibilitySettings) error {
	return r.db.Save(item).Error
}
