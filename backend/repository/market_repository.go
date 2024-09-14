package repository

import (
	"socialpredict/models"
)

type MarketRepository struct {
	db Database
}

func NewMarketRepository(db Database) *MarketRepository {
	return &MarketRepository{db: db}
}

func (repo *MarketRepository) GetAllMarkets() ([]models.Market, error) {
	var markets []models.Market
	result := repo.db.Find(&markets)
	if err := result.Error(); err != nil {
		return nil, err
	}
	return markets, nil
}

func (repo *MarketRepository) GetMarketByID(id int64) (*models.Market, error) {
	var market models.Market
	result := repo.db.First(&market, id)
	if err := result.Error(); err != nil {
		return nil, err
	}
	return &market, nil
}
