package bets

import (
	"context"

	"socialpredict/models"

	"gorm.io/gorm"
)

// GormRepository implements the bets repository using GORM.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new bets repository backed by GORM.
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// Create persists a bet record.
func (r *GormRepository) Create(ctx context.Context, bet *models.Bet) error {
	return r.db.WithContext(ctx).Create(bet).Error
}

// UserHasBet checks whether the user has previously placed a bet in the market.
func (r *GormRepository) UserHasBet(ctx context.Context, marketID uint, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Bet{}).
		Where("market_id = ? AND username = ?", marketID, username).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
