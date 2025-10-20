package markets

import (
	"context"
	"errors"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
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

// Search searches for markets by query string
func (r *GormRepository) Search(ctx context.Context, query string, filters dmarkets.SearchFilters) ([]*dmarkets.Market, error) {
	dbQuery := r.db.WithContext(ctx).Model(&models.Market{})

	// Search in question title and description
	searchPattern := "%" + query + "%"
	dbQuery = dbQuery.Where("question_title ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)

	if filters.Status != "" {
		if filters.Status == "active" {
			dbQuery = dbQuery.Where("is_resolved = ?", false)
		} else if filters.Status == "resolved" {
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

// domainToModel converts a domain market to a GORM model
func (r *GormRepository) domainToModel(market *dmarkets.Market) models.Market {
	return models.Market{
		ID:                 market.ID,
		QuestionTitle:      market.QuestionTitle,
		Description:        market.Description,
		OutcomeType:        market.OutcomeType,
		ResolutionDateTime: market.ResolutionDateTime,
		CreatorUsername:    market.CreatorUsername,
		YesLabel:           market.YesLabel,
		NoLabel:            market.NoLabel,
		IsResolved:         market.Status == "resolved",
		InitialProbability: 0.5, // Default initial probability
	}
}

// modelToDomain converts a GORM model to a domain market
func (r *GormRepository) modelToDomain(dbMarket *models.Market) *dmarkets.Market {
	status := "active"
	if dbMarket.IsResolved {
		status = "resolved"
	}

	return &dmarkets.Market{
		ID:                 dbMarket.ID,
		QuestionTitle:      dbMarket.QuestionTitle,
		Description:        dbMarket.Description,
		OutcomeType:        dbMarket.OutcomeType,
		ResolutionDateTime: dbMarket.ResolutionDateTime,
		CreatorUsername:    dbMarket.CreatorUsername,
		YesLabel:           dbMarket.YesLabel,
		NoLabel:            dbMarket.NoLabel,
		Status:             status,
		CreatedAt:          dbMarket.CreatedAt,
		UpdatedAt:          dbMarket.UpdatedAt,
	}
}
