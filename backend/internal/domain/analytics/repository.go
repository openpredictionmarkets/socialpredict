package analytics

import (
	"context"
	"errors"
	"sort"
	"time"

	"socialpredict/internal/domain/boundary"
	positionsmath "socialpredict/internal/domain/math/positions"

	"gorm.io/gorm"
)

// GormRepository implements the analytics repository interface using GORM.
type GormRepository struct {
	db                 *gorm.DB
	positionCalculator MarketPositionCalculator
}

// RepositoryOption configures the GormRepository strategies.
type RepositoryOption func(*GormRepository)

func defaultRepositoryPositionCalculator() MarketPositionCalculator {
	return defaultMarketPositionCalculator{}
}

func positionCalculatorOrDefault(calculator MarketPositionCalculator) MarketPositionCalculator {
	if calculator == nil {
		return defaultRepositoryPositionCalculator()
	}
	return calculator
}

// WithRepositoryPositionCalculator overrides the default position calculator for the repository.
func WithRepositoryPositionCalculator(c MarketPositionCalculator) RepositoryOption {
	return func(r *GormRepository) {
		if r != nil {
			r.positionCalculator = positionCalculatorOrDefault(c)
		}
	}
}

// NewGormRepository constructs a GORM-backed analytics repository.
func NewGormRepository(db *gorm.DB, opts ...RepositoryOption) *GormRepository {
	repo := &GormRepository{
		db:                 db,
		positionCalculator: defaultRepositoryPositionCalculator(),
	}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

func (r *GormRepository) WithContext(ctx context.Context) *gorm.DB {
	if r == nil || r.db == nil {
		return nil
	}
	if ctx == nil {
		return r.db
	}
	return r.db.WithContext(ctx)
}

func (r *GormRepository) dbWithContext(ctx context.Context) (*gorm.DB, error) {
	db := r.WithContext(ctx)
	if db == nil {
		return nil, errors.New("gorm repository not initialized")
	}
	return db, nil
}

func (r *GormRepository) ListUsers(ctx context.Context) ([]UserAccount, error) {
	db, err := r.dbWithContext(ctx)
	if err != nil {
		return nil, err
	}
	var users []analyticsUserRow
	if err := db.Table("users").
		Select("username", "account_balance").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return mapUsers(users), nil
}

func (r *GormRepository) CountUsersByType(ctx context.Context, userType string) (int64, error) {
	db, err := r.dbWithContext(ctx)
	if err != nil {
		return 0, err
	}
	var count int64
	if err := db.Table("users").Where("user_type = ?", userType).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *GormRepository) ListMarkets(ctx context.Context) ([]MarketRecord, error) {
	db, err := r.dbWithContext(ctx)
	if err != nil {
		return nil, err
	}
	var markets []analyticsMarketRow
	if err := db.Table("markets").
		Select("id", "created_at", "is_resolved", "resolution_result").
		Find(&markets).Error; err != nil {
		return nil, err
	}
	return mapMarkets(markets), nil
}

func (r *GormRepository) ListBetsForMarket(ctx context.Context, marketID uint) ([]boundary.Bet, error) {
	db, err := r.dbWithContext(ctx)
	if err != nil {
		return nil, err
	}
	var bets []analyticsBetRow
	if err := db.Table("bets").
		Select("id", "username", "market_id", "amount", "outcome", "placed_at", "created_at").
		Where("market_id = ?", marketID).
		Order("placed_at ASC").
		Find(&bets).Error; err != nil {
		return nil, err
	}
	return mapBets(bets), nil
}

func (r *GormRepository) ListBetsOrdered(ctx context.Context) ([]boundary.Bet, error) {
	db, err := r.dbWithContext(ctx)
	if err != nil {
		return nil, err
	}
	var bets []analyticsBetRow
	if err := db.Table("bets").
		Select("id", "username", "market_id", "amount", "outcome", "placed_at", "created_at").
		Order("market_id ASC, placed_at ASC, id ASC").
		Find(&bets).Error; err != nil {
		return nil, err
	}
	return mapBets(bets), nil
}

func (r *GormRepository) UserMarketPositions(ctx context.Context, username string) ([]positionsmath.MarketPosition, error) {
	db, err := r.dbWithContext(ctx)
	if err != nil {
		return nil, err
	}

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
	calculator := r.ensurePositionCalculator()

	positions, err := r.calculateUserPositions(calculator, username, marketIDs, snapshots, betsByMarket)
	if err != nil {
		return nil, err
	}

	return positions, nil
}

func (r *GormRepository) listUserBets(db *gorm.DB, username string) ([]boundary.Bet, error) {
	var userBets []analyticsBetRow
	if err := db.Table("bets").
		Select("id", "username", "market_id", "amount", "outcome", "placed_at", "created_at").
		Where("username = ?", username).
		Order("market_id ASC, placed_at ASC, id ASC").
		Find(&userBets).Error; err != nil {
		return nil, err
	}
	return mapBets(userBets), nil
}

func collectMarketIDs(bets []boundary.Bet) []uint {
	marketIDSet := make(map[uint]struct{}, len(bets))
	for _, bet := range bets {
		marketIDSet[bet.MarketID] = struct{}{}
	}

	marketIDs := make([]uint, 0, len(marketIDSet))
	for id := range marketIDSet {
		marketIDs = append(marketIDs, uint(id))
	}
	sort.Slice(marketIDs, func(i, j int) bool { return marketIDs[i] < marketIDs[j] })
	return marketIDs
}

func (r *GormRepository) listMarketsByIDs(db *gorm.DB, marketIDs []uint) ([]MarketRecord, error) {
	if len(marketIDs) == 0 {
		return []MarketRecord{}, nil
	}
	var markets []analyticsMarketRow
	if err := db.Table("markets").
		Select("id", "created_at", "is_resolved", "resolution_result").
		Where("id IN ?", marketIDs).
		Find(&markets).Error; err != nil {
		return nil, err
	}
	return mapMarkets(markets), nil
}

func buildMarketSnapshots(markets []MarketRecord) map[int64]positionsmath.MarketSnapshot {
	marketSnapshots := make(map[int64]positionsmath.MarketSnapshot, len(markets))
	for _, market := range markets {
		marketSnapshots[int64(market.ID)] = market.Snapshot()
	}
	return marketSnapshots
}

func (r *GormRepository) listBetsByMarketIDs(db *gorm.DB, marketIDs []uint) ([]boundary.Bet, error) {
	if len(marketIDs) == 0 {
		return []boundary.Bet{}, nil
	}
	var allBets []analyticsBetRow
	if err := db.Table("bets").
		Select("id", "username", "market_id", "amount", "outcome", "placed_at", "created_at").
		Where("market_id IN ?", marketIDs).
		Order("market_id ASC, placed_at ASC, id ASC").
		Find(&allBets).Error; err != nil {
		return nil, err
	}
	return mapBets(allBets), nil
}

func groupBetsByMarket(bets []boundary.Bet) map[int64][]boundary.Bet {
	betsByMarket := make(map[int64][]boundary.Bet, len(bets))
	for _, bet := range bets {
		betsByMarket[int64(bet.MarketID)] = append(betsByMarket[int64(bet.MarketID)], bet)
	}
	return betsByMarket
}

func (r *GormRepository) ensurePositionCalculator() MarketPositionCalculator {
	if r == nil {
		return defaultRepositoryPositionCalculator()
	}
	r.positionCalculator = positionCalculatorOrDefault(r.positionCalculator)
	return r.positionCalculator
}

func (r *GormRepository) calculateUserPositions(calculator MarketPositionCalculator, username string, marketIDs []uint, snapshots map[int64]positionsmath.MarketSnapshot, betsByMarket map[int64][]boundary.Bet) ([]positionsmath.MarketPosition, error) {
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

		calculated, err := calculator.Calculate(snapshot, bets)
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

type analyticsUserRow struct {
	Username       string
	AccountBalance int64
}

type analyticsMarketRow struct {
	ID               uint
	CreatedAt        time.Time
	IsResolved       bool
	ResolutionResult string
}

type analyticsBetRow struct {
	ID        uint
	Username  string
	MarketID  uint
	Amount    int64
	Outcome   string
	PlacedAt  time.Time
	CreatedAt time.Time
}

func mapUsers(dbUsers []analyticsUserRow) []UserAccount {
	users := make([]UserAccount, len(dbUsers))
	for i, user := range dbUsers {
		users[i] = UserAccount{
			Username:       user.Username,
			AccountBalance: user.AccountBalance,
		}
	}
	return users
}

func mapMarkets(dbMarkets []analyticsMarketRow) []MarketRecord {
	markets := make([]MarketRecord, len(dbMarkets))
	for i, market := range dbMarkets {
		markets[i] = MarketRecord{
			ID:               market.ID,
			CreatedAt:        market.CreatedAt,
			IsResolved:       market.IsResolved,
			ResolutionResult: market.ResolutionResult,
		}
	}
	return markets
}

func mapBets(dbBets []analyticsBetRow) []boundary.Bet {
	bets := make([]boundary.Bet, len(dbBets))
	for i, bet := range dbBets {
		bets[i] = boundary.Bet{
			ID:        bet.ID,
			Username:  bet.Username,
			MarketID:  bet.MarketID,
			Amount:    bet.Amount,
			Outcome:   bet.Outcome,
			PlacedAt:  bet.PlacedAt,
			CreatedAt: bet.CreatedAt,
		}
	}
	return bets
}

var (
	_ Repository            = (*GormRepository)(nil)
	_ LeaderboardRepository = (*GormRepository)(nil)
	_ FinancialsRepository  = (*GormRepository)(nil)
	_ DebtRepository        = (*GormRepository)(nil)
	_ VolumeRepository      = (*GormRepository)(nil)
	_ FeeRepository         = (*GormRepository)(nil)
)
