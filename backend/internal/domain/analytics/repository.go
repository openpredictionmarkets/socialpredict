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
	db := r.WithContext(ctx)

	var userBets []models.Bet
	if err := db.Where("username = ?", username).
		Order("market_id ASC, placed_at ASC, id ASC").
		Find(&userBets).Error; err != nil {
		return nil, err
	}

	if len(userBets) == 0 {
		return []positionsmath.MarketPosition{}, nil
	}

	marketIDSet := make(map[int64]struct{})
	for _, bet := range userBets {
		marketIDSet[int64(bet.MarketID)] = struct{}{}
	}

	marketIDs := make([]uint, 0, len(marketIDSet))
	for id := range marketIDSet {
		marketIDs = append(marketIDs, uint(id))
	}

	var markets []models.Market
	if err := db.Where("id IN ?", marketIDs).Find(&markets).Error; err != nil {
		return nil, err
	}

	marketSnapshots := make(map[int64]positionsmath.MarketSnapshot, len(markets))
	for _, market := range markets {
		marketSnapshots[int64(market.ID)] = positionsmath.MarketSnapshot{
			ID:               int64(market.ID),
			CreatedAt:        market.CreatedAt,
			IsResolved:       market.IsResolved,
			ResolutionResult: market.ResolutionResult,
		}
	}

	var allBets []models.Bet
	if err := db.Where("market_id IN ?", marketIDs).
		Order("market_id ASC, placed_at ASC, id ASC").
		Find(&allBets).Error; err != nil {
		return nil, err
	}

	betsByMarket := make(map[int64][]models.Bet)
	for _, bet := range allBets {
		betsByMarket[int64(bet.MarketID)] = append(betsByMarket[int64(bet.MarketID)], bet)
	}

	var positions []positionsmath.MarketPosition
	for _, marketID := range marketIDs {
		snapshot, ok := marketSnapshots[int64(marketID)]
		if !ok {
			continue
		}

		bets := betsByMarket[int64(marketID)]
		if len(bets) == 0 {
			continue
		}

		calculated, err := positionsmath.CalculateMarketPositions_WPAM_DBPM(snapshot, bets)
		if err != nil {
			return nil, err
		}

		for _, pos := range calculated {
			if pos.Username == username {
				positions = append(positions, pos)
				break
			}
		}
	}

	return positions, nil
}
