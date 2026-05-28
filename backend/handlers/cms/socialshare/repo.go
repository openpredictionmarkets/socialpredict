package socialshare

import (
	"socialpredict/models"

	"gorm.io/gorm"
)

type Repository interface {
	GetBySlug(slug string) (*models.SocialShareSettings, error)
	Save(item *models.SocialShareSettings) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) GetBySlug(slug string) (*models.SocialShareSettings, error) {
	var item models.SocialShareSettings
	if err := r.db.Where("slug = ?", slug).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Save(item *models.SocialShareSettings) error {
	return r.db.Save(item).Error
}
