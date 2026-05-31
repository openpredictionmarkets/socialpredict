package bets

import (
	"context"
	"errors"
	"time"

	dbets "socialpredict/internal/domain/bets"
	"socialpredict/internal/domain/boundary"
	dmarkets "socialpredict/internal/domain/markets"
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ dbets.SellUnitOfWork = (*GormRepository)(nil)

// SellBetTransaction commits the sale credit and sale bet as one unit of work.
func (r *GormRepository) SellBetTransaction(ctx context.Context, fn dbets.SellTransactionFunc) error {
	// The transaction-scoped market reader locks the market row before deriving
	// the user's position so overlapping sell settlements serialize per market.
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, NewGormRepository(tx), sellMarketRepository{db: tx}, newPlaceUserService(tx))
	})
}

type sellMarketRepository struct {
	db *gorm.DB
}

func (r sellMarketRepository) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	var market models.Market

	query := r.db.WithContext(ctx)
	if r.db.Dialector.Name() == "postgres" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.First(&market, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dmarkets.ErrMarketNotFound
		}
		return nil, err
	}

	return sellMarketModelToDomain(&market), nil
}

func (r sellMarketRepository) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	snapshot, bets, err := r.loadMarketData(ctx, marketID)
	if err != nil {
		return nil, err
	}

	position, err := positionsmath.CalculateMarketPositionForUser_WPAM_DBPM(snapshot, bets, username)
	if err != nil {
		return nil, err
	}

	return &dmarkets.UserPosition{
		Username:         username,
		MarketID:         marketID,
		YesSharesOwned:   position.YesSharesOwned,
		NoSharesOwned:    position.NoSharesOwned,
		Value:            position.Value,
		TotalSpent:       position.TotalSpent,
		TotalSpentInPlay: position.TotalSpentInPlay,
		IsResolved:       position.IsResolved,
		ResolutionResult: position.ResolutionResult,
	}, nil
}

func (r sellMarketRepository) loadMarketData(ctx context.Context, marketID int64) (positionsmath.MarketSnapshot, []boundary.Bet, error) {
	var market models.Market
	marketQuery := r.db.WithContext(ctx)
	if r.db.Dialector.Name() == "postgres" {
		marketQuery = marketQuery.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := marketQuery.First(&market, marketID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return positionsmath.MarketSnapshot{}, nil, dmarkets.ErrMarketNotFound
		}
		return positionsmath.MarketSnapshot{}, nil, err
	}

	var dbBets []models.Bet
	betsQuery := r.db.WithContext(ctx).
		Where("market_id = ?", marketID).
		Order("placed_at ASC")
	if r.db.Dialector.Name() == "postgres" {
		betsQuery = betsQuery.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := betsQuery.Find(&dbBets).Error; err != nil {
		return positionsmath.MarketSnapshot{}, nil, err
	}

	snapshot := positionsmath.MarketSnapshot{
		ID:               int64(market.ID),
		CreatedAt:        market.CreatedAt,
		IsResolved:       market.IsResolved,
		ResolutionResult: market.ResolutionResult,
	}

	return snapshot, sellModelBetsToBoundary(dbBets), nil
}

func sellModelBetsToBoundary(dbBets []models.Bet) []boundary.Bet {
	bets := make([]boundary.Bet, len(dbBets))
	for i, bet := range dbBets {
		bets[i] = boundary.Bet{
			ID:        uint(bet.ID),
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

func sellMarketModelToDomain(dbMarket *models.Market) *dmarkets.Market {
	status := "active"
	switch {
	case dbMarket.IsResolved:
		status = "resolved"
	case !dbMarket.ResolutionDateTime.After(time.Now()):
		status = "closed"
	}

	return &dmarkets.Market{
		ID:                      dbMarket.ID,
		QuestionTitle:           dbMarket.QuestionTitle,
		Description:             dbMarket.Description,
		OutcomeType:             dbMarket.OutcomeType,
		ResolutionDateTime:      dbMarket.ResolutionDateTime,
		FinalResolutionDateTime: dbMarket.FinalResolutionDateTime,
		ResolutionResult:        dbMarket.ResolutionResult,
		CreatorUsername:         dbMarket.CreatorUsername,
		YesLabel:                dbMarket.YesLabel,
		NoLabel:                 dbMarket.NoLabel,
		Status:                  status,
		CreatedAt:               dbMarket.CreatedAt,
		UpdatedAt:               dbMarket.UpdatedAt,
		InitialProbability:      dbMarket.InitialProbability,
		UTCOffset:               dbMarket.UTCOffset,
	}
}
