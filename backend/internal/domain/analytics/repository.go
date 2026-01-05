package analytics

import (
	"context"
	"sort"

	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"

	"gorm.io/gorm"
)

// GormRepository implements the analytics repository interface using GORM.
type GormRepository struct {
	db                 *gorm.DB
	positionCalculator MarketPositionCalculator
}

// RepositoryOption configures the GormRepository strategies.
type RepositoryOption func(*GormRepository)

// WithRepositoryPositionCalculator overrides the default position calculator for the repository.
func WithRepositoryPositionCalculator(c MarketPositionCalculator) RepositoryOption {
	return func(r *GormRepository) {
		if c != nil {
			r.positionCalculator = c
		}
	}
}

// NewGormRepository constructs a GORM-backed analytics repository.
func NewGormRepository(db *gorm.DB, opts ...RepositoryOption) *GormRepository {
	repo := &GormRepository{
		db:                 db,
		positionCalculator: defaultMarketPositionCalculator{},
	}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
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

	userBets, err := r.listUserBets(db, username)
	if err != nil {
		return nil, err
	}
	if len(userBets) == 0 {
		return []positionsmath.MarketPosition{}, nil
	}

	marketIDs := collectMarketIDs(userBets)

	markets, err := r.listMarketsByIDs(db, marketIDs)
	if err != nil {
		return nil, err
	}

	snapshots := buildMarketSnapshots(markets)

	allBets, err := r.listBetsByMarketIDs(db, marketIDs)
	if err != nil {
		return nil, err
	}

	betsByMarket := groupBetsByMarket(allBets)

	positions, err := r.calculateUserPositions(username, marketIDs, snapshots, betsByMarket)
	if err != nil {
		return nil, err
	}

	return positions, nil
}

func (r *GormRepository) listUserBets(db *gorm.DB, username string) ([]models.Bet, error) {
	var userBets []models.Bet
	if err := db.Where("username = ?", username).
		Order("market_id ASC, placed_at ASC, id ASC").
		Find(&userBets).Error; err != nil {
		return nil, err
	}
	return userBets, nil
}

func collectMarketIDs(bets []models.Bet) []uint {
	marketIDSet := make(map[int64]struct{})
	for _, bet := range bets {
		marketIDSet[int64(bet.MarketID)] = struct{}{}
	}

	marketIDs := make([]uint, 0, len(marketIDSet))
	for id := range marketIDSet {
		marketIDs = append(marketIDs, uint(id))
	}
	sort.Slice(marketIDs, func(i, j int) bool { return marketIDs[i] < marketIDs[j] })
	return marketIDs
}

func (r *GormRepository) listMarketsByIDs(db *gorm.DB, marketIDs []uint) ([]models.Market, error) {
	var markets []models.Market
	if err := db.Where("id IN ?", marketIDs).Find(&markets).Error; err != nil {
		return nil, err
	}
	return markets, nil
}

func buildMarketSnapshots(markets []models.Market) map[int64]positionsmath.MarketSnapshot {
	marketSnapshots := make(map[int64]positionsmath.MarketSnapshot, len(markets))
	for _, market := range markets {
		marketSnapshots[int64(market.ID)] = positionsmath.MarketSnapshot{
			ID:               int64(market.ID),
			CreatedAt:        market.CreatedAt,
			IsResolved:       market.IsResolved,
			ResolutionResult: market.ResolutionResult,
		}
	}
	return marketSnapshots
}

func (r *GormRepository) listBetsByMarketIDs(db *gorm.DB, marketIDs []uint) ([]models.Bet, error) {
	var allBets []models.Bet
	if err := db.Where("market_id IN ?", marketIDs).
		Order("market_id ASC, placed_at ASC, id ASC").
		Find(&allBets).Error; err != nil {
		return nil, err
	}
	return allBets, nil
}

func groupBetsByMarket(bets []models.Bet) map[int64][]models.Bet {
	betsByMarket := make(map[int64][]models.Bet)
	for _, bet := range bets {
		betsByMarket[int64(bet.MarketID)] = append(betsByMarket[int64(bet.MarketID)], bet)
	}
	return betsByMarket
}

func (r *GormRepository) ensurePositionCalculator() {
	if r.positionCalculator == nil {
		r.positionCalculator = defaultMarketPositionCalculator{}
	}
}

func (r *GormRepository) calculateUserPositions(username string, marketIDs []uint, snapshots map[int64]positionsmath.MarketSnapshot, betsByMarket map[int64][]models.Bet) ([]positionsmath.MarketPosition, error) {
	r.ensurePositionCalculator()
	var positions []positionsmath.MarketPosition
	for _, marketID := range marketIDs {
		snapshot, ok := snapshots[int64(marketID)]
		if !ok {
			continue
		}

		bets := betsByMarket[int64(marketID)]
		if len(bets) == 0 {
			continue
		}

		calculated, err := r.positionCalculator.Calculate(snapshot, bets)
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
