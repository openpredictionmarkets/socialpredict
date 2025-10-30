package analytics

import (
	"context"

	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"

	"gorm.io/gorm"
)

// GormRepository implements the analytics repository interface using GORM.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository constructs a GORM-backed analytics repository.
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) WithContext(ctx context.Context) *gorm.DB {
	if ctx != nil {
		return r.db.WithContext(ctx)
	}
	return r.db
}

func (r *GormRepository) ListUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := r.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *GormRepository) ListMarkets(ctx context.Context) ([]models.Market, error) {
	var markets []models.Market
	if err := r.WithContext(ctx).Find(&markets).Error; err != nil {
		return nil, err
	}
	return markets, nil
}

func (r *GormRepository) ListBetsForMarket(ctx context.Context, marketID uint) ([]models.Bet, error) {
	var bets []models.Bet
	if err := r.WithContext(ctx).
		Where("market_id = ?", marketID).
		Order("placed_at ASC").
		Find(&bets).Error; err != nil {
		return nil, err
	}
	return bets, nil
}

func (r *GormRepository) ListBetsOrdered(ctx context.Context) ([]models.Bet, error) {
	var bets []models.Bet
	if err := r.WithContext(ctx).
		Order("market_id ASC, placed_at ASC, id ASC").
		Find(&bets).Error; err != nil {
		return nil, err
	}
	return bets, nil
}

func (r *GormRepository) UserMarketPositions(ctx context.Context, username string) ([]positionsmath.MarketPosition, error) {
	return positionsmath.CalculateAllUserMarketPositions_WPAM_DBPM(r.WithContext(ctx), username)
}
