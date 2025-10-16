package homepage

import (
	"socialpredict/models"

	"gorm.io/gorm"
)

type Repository interface {
	GetBySlug(slug string) (*models.HomepageContent, error)
	Save(item *models.HomepageContent) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) GetBySlug(slug string) (*models.HomepageContent, error) {
	var item models.HomepageContent
	if err := r.db.Where("slug = ?", slug).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Save(item *models.HomepageContent) error {
	return r.db.Save(item).Error
}
