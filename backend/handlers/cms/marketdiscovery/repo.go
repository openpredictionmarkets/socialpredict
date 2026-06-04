package marketdiscovery

import (
	"socialpredict/models"

	"gorm.io/gorm"
)

type Repository interface {
	GetPageBySlug(slug string) (*models.MarketDiscoveryPage, error)
	SavePage(page *models.MarketDiscoveryPage) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) GetPageBySlug(slug string) (*models.MarketDiscoveryPage, error) {
	var page models.MarketDiscoveryPage
	if err := r.db.Where("slug = ?", slug).First(&page).Error; err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *GormRepository) SavePage(page *models.MarketDiscoveryPage) error {
	return r.db.Save(page).Error
}
