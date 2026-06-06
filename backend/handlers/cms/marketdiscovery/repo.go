package marketdiscovery

import (
	"socialpredict/models"

	"gorm.io/gorm"
)

type Repository interface {
	GetPageBySlug(slug string) (*models.MarketDiscoveryPage, error)
	SavePage(page *models.MarketDiscoveryPage) error
	ListPins(pageID uint) ([]models.MarketDiscoveryPin, error)
	ReplacePins(pageID uint, pins []models.MarketDiscoveryPin) error
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

func (r *GormRepository) ListPins(pageID uint) ([]models.MarketDiscoveryPin, error) {
	var pins []models.MarketDiscoveryPin
	if err := r.db.Where("scope_type = ? AND scope_id = ?", "page", pageID).Order("sort_order ASC, id ASC").Find(&pins).Error; err != nil {
		return nil, err
	}
	return pins, nil
}

func (r *GormRepository) ReplacePins(pageID uint, pins []models.MarketDiscoveryPin) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("scope_type = ? AND scope_id = ?", "page", pageID).Delete(&models.MarketDiscoveryPin{}).Error; err != nil {
			return err
		}
		for index := range pins {
			pins[index].ScopeType = "page"
			pins[index].ScopeID = pageID
			if err := tx.Create(&pins[index]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
