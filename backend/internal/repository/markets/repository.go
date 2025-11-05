package markets

import (
	"context"
	"errors"
	"strings"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"

	"gorm.io/gorm"
)

// GormRepository implements the markets domain repository interface using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM-based markets repository
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// Create creates a new market in the database
func (r *GormRepository) Create(ctx context.Context, market *dmarkets.Market) error {
	dbMarket := r.domainToModel(market)

	result := r.db.WithContext(ctx).Create(&dbMarket)
	if result.Error != nil {
		return result.Error
	}

	// Update the domain model with the generated ID
	market.ID = dbMarket.ID
	return nil
}

// GetByID retrieves a market by its ID
func (r *GormRepository) GetByID(ctx context.Context, id int64) (*dmarkets.Market, error) {
	var dbMarket models.Market

	err := r.db.WithContext(ctx).First(&dbMarket, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dmarkets.ErrMarketNotFound
		}
		return nil, err
	}

	return r.modelToDomain(&dbMarket), nil
}

// GetPublicMarket retrieves a market with public-facing attributes.
func (r *GormRepository) GetPublicMarket(ctx context.Context, marketID int64) (*dmarkets.PublicMarket, error) {
	var market models.Market
	if err := r.db.WithContext(ctx).First(&market, marketID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dmarkets.ErrMarketNotFound
		}
		return nil, err
	}

	return &dmarkets.PublicMarket{
		ID:                      market.ID,
		QuestionTitle:           market.QuestionTitle,
		Description:             market.Description,
		OutcomeType:             market.OutcomeType,
		ResolutionDateTime:      market.ResolutionDateTime,
		FinalResolutionDateTime: market.FinalResolutionDateTime,
		UTCOffset:               market.UTCOffset,
		IsResolved:              market.IsResolved,
		ResolutionResult:        market.ResolutionResult,
		InitialProbability:      market.InitialProbability,
		CreatorUsername:         market.CreatorUsername,
		CreatedAt:               market.CreatedAt,
		YesLabel:                market.YesLabel,
		NoLabel:                 market.NoLabel,
	}, nil
}

// UpdateLabels updates the yes and no labels for a market
func (r *GormRepository) UpdateLabels(ctx context.Context, id int64, yesLabel, noLabel string) error {
	result := r.db.WithContext(ctx).Model(&models.Market{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"yes_label":  yesLabel,
			"no_label":   noLabel,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return dmarkets.ErrMarketNotFound
	}

	return nil
}

// List retrieves markets with the given filters
func (r *GormRepository) List(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	query := r.db.WithContext(ctx).Model(&models.Market{})

	if filters.Status != "" {
		if filters.Status == "active" {
			query = query.Where("is_resolved = ?", false)
		} else if filters.Status == "resolved" {
			query = query.Where("is_resolved = ?", true)
		}
	}

	if filters.CreatedBy != "" {
		query = query.Where("creator_username = ?", filters.CreatedBy)
	}

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	query = query.Order("created_at DESC")

	var dbMarkets []models.Market
	if err := query.Find(&dbMarkets).Error; err != nil {
		return nil, err
	}

	markets := make([]*dmarkets.Market, len(dbMarkets))
	for i, dbMarket := range dbMarkets {
		markets[i] = r.modelToDomain(&dbMarket)
	}

	return markets, nil
}

// ListByStatus retrieves markets filtered by status with pagination
func (r *GormRepository) ListByStatus(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error) {
	query := r.db.WithContext(ctx).Model(&models.Market{})

	// Apply status filter
	now := time.Now()
	switch status {
	case "active":
		query = query.Where("is_resolved = ? AND resolution_date_time > ?", false, now)
	case "closed":
		query = query.Where("is_resolved = ? AND resolution_date_time <= ?", false, now)
	case "resolved":
		query = query.Where("is_resolved = ?", true)
	case "all":
		// No status filter
	}

	// Apply pagination
	if p.Limit > 0 {
		query = query.Limit(p.Limit)
	}
	if p.Offset > 0 {
		query = query.Offset(p.Offset)
	}

	query = query.Order("created_at DESC")

	var dbMarkets []models.Market
	if err := query.Find(&dbMarkets).Error; err != nil {
		return nil, err
	}

	markets := make([]*dmarkets.Market, len(dbMarkets))
	for i, dbMarket := range dbMarkets {
		markets[i] = r.modelToDomain(&dbMarket)
	}

	return markets, nil
}

// Search searches for markets by query string
func (r *GormRepository) Search(ctx context.Context, query string, filters dmarkets.SearchFilters) ([]*dmarkets.Market, error) {
	dbQuery := r.db.WithContext(ctx).Model(&models.Market{})

	// Search in question title and description (case insensitive, SQLite compatible)
	searchTerm := strings.ToLower(query)
	searchPattern := "%" + searchTerm + "%"
	dbQuery = dbQuery.Where("(LOWER(question_title) LIKE ? OR LOWER(description) LIKE ?)", searchPattern, searchPattern)

	if filters.Status != "" && filters.Status != "all" {
		now := time.Now()
		switch filters.Status {
		case "active":
			dbQuery = dbQuery.Where("is_resolved = ? AND resolution_date_time > ?", false, now)
		case "closed":
			dbQuery = dbQuery.Where("is_resolved = ? AND resolution_date_time <= ?", false, now)
		case "resolved":
			dbQuery = dbQuery.Where("is_resolved = ?", true)
		}
	}

	if filters.Limit > 0 {
		dbQuery = dbQuery.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		dbQuery = dbQuery.Offset(filters.Offset)
	}

	dbQuery = dbQuery.Order("created_at DESC")

	var dbMarkets []models.Market
	if err := dbQuery.Find(&dbMarkets).Error; err != nil {
		return nil, err
	}

	markets := make([]*dmarkets.Market, len(dbMarkets))
	for i, dbMarket := range dbMarkets {
		markets[i] = r.modelToDomain(&dbMarket)
	}

	return markets, nil
}

// GetUserPosition retrieves the aggregated position for a specific user in a market.
func (r *GormRepository) GetUserPosition(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
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

// ListMarketPositions retrieves aggregated positions for all users in a market.
func (r *GormRepository) ListMarketPositions(ctx context.Context, marketID int64) (dmarkets.MarketPositions, error) {
	snapshot, bets, err := r.loadMarketData(ctx, marketID)
	if err != nil {
		return nil, err
	}

	positions, err := positionsmath.CalculateMarketPositions_WPAM_DBPM(snapshot, bets)
	if err != nil {
		return nil, err
	}

	out := make(dmarkets.MarketPositions, 0, len(positions))
	for _, pos := range positions {
		out = append(out, &dmarkets.UserPosition{
			Username:         pos.Username,
			MarketID:         int64(pos.MarketID),
			YesSharesOwned:   pos.YesSharesOwned,
			NoSharesOwned:    pos.NoSharesOwned,
			Value:            pos.Value,
			TotalSpent:       pos.TotalSpent,
			TotalSpentInPlay: pos.TotalSpentInPlay,
			IsResolved:       pos.IsResolved,
			ResolutionResult: pos.ResolutionResult,
		})
	}
	return out, nil
}

// Delete removes a market from the database
func (r *GormRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Market{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return dmarkets.ErrMarketNotFound
	}

	return nil
}

// ResolveMarket marks a market as resolved with the given resolution
func (r *GormRepository) ResolveMarket(ctx context.Context, id int64, resolution string) error {
	result := r.db.WithContext(ctx).Model(&models.Market{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_resolved":                true,
			"resolution_result":          resolution,
			"final_resolution_date_time": time.Now(),
			"updated_at":                 time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return dmarkets.ErrMarketNotFound
	}

	return nil
}

// ListBetsForMarket returns all bets for the specified market ordered by placement time.
func (r *GormRepository) ListBetsForMarket(ctx context.Context, marketID int64) ([]*dmarkets.Bet, error) {
	var bets []models.Bet
	if err := r.db.WithContext(ctx).
		Where("market_id = ?", marketID).
		Order("placed_at ASC").
		Find(&bets).Error; err != nil {
		return nil, err
	}

	result := make([]*dmarkets.Bet, len(bets))
	for i := range bets {
		result[i] = &dmarkets.Bet{
			ID:        bets[i].ID,
			Username:  bets[i].Username,
			MarketID:  bets[i].MarketID,
			Amount:    bets[i].Amount,
			Outcome:   bets[i].Outcome,
			PlacedAt:  bets[i].PlacedAt,
			CreatedAt: bets[i].CreatedAt,
		}
	}
	return result, nil
}

// CalculatePayoutPositions computes the resolved valuations for a market's participants.
func (r *GormRepository) CalculatePayoutPositions(ctx context.Context, marketID int64) ([]*dmarkets.PayoutPosition, error) {
	snapshot, bets, err := r.loadMarketData(ctx, marketID)
	if err != nil {
		return nil, err
	}

	positions, err := positionsmath.CalculateMarketPositions_WPAM_DBPM(snapshot, bets)
	if err != nil {
		return nil, err
	}

	result := make([]*dmarkets.PayoutPosition, 0, len(positions))
	for _, pos := range positions {
		result = append(result, &dmarkets.PayoutPosition{
			Username: pos.Username,
			Value:    pos.Value,
		})
	}
	return result, nil
}

func (r *GormRepository) loadMarketData(ctx context.Context, marketID int64) (positionsmath.MarketSnapshot, []models.Bet, error) {
	var market models.Market
	if err := r.db.WithContext(ctx).First(&market, marketID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return positionsmath.MarketSnapshot{}, nil, dmarkets.ErrMarketNotFound
		}
		return positionsmath.MarketSnapshot{}, nil, err
	}

	var bets []models.Bet
	if err := r.db.WithContext(ctx).
		Where("market_id = ?", marketID).
		Order("placed_at ASC").
		Find(&bets).Error; err != nil {
		return positionsmath.MarketSnapshot{}, nil, err
	}

	snapshot := positionsmath.MarketSnapshot{
		ID:               int64(market.ID),
		CreatedAt:        market.CreatedAt,
		IsResolved:       market.IsResolved,
		ResolutionResult: market.ResolutionResult,
	}

	return snapshot, bets, nil
}

// domainToModel converts a domain market to a GORM model
func (r *GormRepository) domainToModel(market *dmarkets.Market) models.Market {
	return models.Market{
		ID:                      market.ID,
		QuestionTitle:           market.QuestionTitle,
		Description:             market.Description,
		OutcomeType:             market.OutcomeType,
		ResolutionDateTime:      market.ResolutionDateTime,
		FinalResolutionDateTime: market.FinalResolutionDateTime,
		ResolutionResult:        market.ResolutionResult,
		CreatorUsername:         market.CreatorUsername,
		YesLabel:                market.YesLabel,
		NoLabel:                 market.NoLabel,
		UTCOffset:               market.UTCOffset,
		IsResolved:              market.Status == "resolved",
		InitialProbability:      market.InitialProbability,
	}
}

// modelToDomain converts a GORM model to a domain market
func (r *GormRepository) modelToDomain(dbMarket *models.Market) *dmarkets.Market {
	status := "active"
	if dbMarket.IsResolved {
		status = "resolved"
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
