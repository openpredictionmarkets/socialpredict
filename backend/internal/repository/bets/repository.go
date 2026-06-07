package bets

import (
	"context"

	"socialpredict/internal/domain/boundary"
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
func (r *GormRepository) Create(ctx context.Context, bet *boundary.Bet) error {
	dbBet := models.Bet{
		ID:       bet.ID,
		Username: bet.Username,
		MarketID: bet.MarketID,
		Amount:   bet.Amount,
		Outcome:  bet.Outcome,
		PlacedAt: bet.PlacedAt,
	}
	if err := r.db.WithContext(ctx).Create(&dbBet).Error; err != nil {
		return err
	}
	bet.ID = uint(dbBet.ID)
	bet.CreatedAt = dbBet.CreatedAt
	return nil
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
