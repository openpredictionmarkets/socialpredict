package markets

import (
	"context"
	"errors"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"

	"gorm.io/gorm"
)

func (r *GormRepository) ListMarketTags(ctx context.Context, includeInactive bool) ([]dmarkets.MarketTag, error) {
	query := r.db.WithContext(ctx).Model(&models.MarketTag{})
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	query = query.Order("sort_order ASC, display_name ASC, slug ASC")

	var tags []models.MarketTag
	if err := query.Find(&tags).Error; err != nil {
		return nil, err
	}
	return modelTagsToDomain(tags), nil
}

func (r *GormRepository) CreateMarketTag(ctx context.Context, tag dmarkets.MarketTag) (*dmarkets.MarketTag, error) {
	dbTag := domainTagToModel(tag)
	if err := r.db.WithContext(ctx).
		Select("Slug", "DisplayName", "Description", "ColorKey", "SortOrder", "IsActive", "CreatedBy").
		Create(&dbTag).Error; err != nil {
		return nil, err
	}
	if !tag.IsActive {
		if err := r.db.WithContext(ctx).Model(&models.MarketTag{}).
			Where("id = ?", dbTag.ID).
			Update("is_active", false).Error; err != nil {
			return nil, err
		}
		if err := r.db.WithContext(ctx).First(&dbTag, dbTag.ID).Error; err != nil {
			return nil, err
		}
	}
	out := modelTagToDomain(dbTag)
	return &out, nil
}

func (r *GormRepository) UpdateMarketTag(ctx context.Context, slug string, update dmarkets.MarketTagRequest) (*dmarkets.MarketTag, error) {
	updates := map[string]any{
		"updated_at": time.Now(),
	}
	if update.DisplayName != "" {
		updates["display_name"] = update.DisplayName
	}
	updates["description"] = update.Description
	updates["color_key"] = update.ColorKey
	updates["sort_order"] = update.SortOrder
	if update.IsActive != nil {
		updates["is_active"] = *update.IsActive
	}

	result := r.db.WithContext(ctx).Model(&models.MarketTag{}).
		Where("slug = ?", slug).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, dmarkets.ErrInvalidInput
	}

	var dbTag models.MarketTag
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&dbTag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dmarkets.ErrInvalidInput
		}
		return nil, err
	}
	out := modelTagToDomain(dbTag)
	return &out, nil
}

func (r *GormRepository) SetMarketTags(ctx context.Context, marketID int64, tagSlugs []string, assignedBy string, source string, assignedAt time.Time) ([]dmarkets.MarketTag, error) {
	if marketID <= 0 {
		return nil, dmarkets.ErrInvalidInput
	}
	normalized, err := dmarkets.NormalizeMarketTagSlugs(tagSlugs)
	if err != nil {
		return nil, err
	}

	var tags []models.MarketTag
	if len(normalized) > 0 {
		if err := r.db.WithContext(ctx).
			Where("slug IN ? AND is_active = ?", normalized, true).
			Order("sort_order ASC, display_name ASC, slug ASC").
			Find(&tags).Error; err != nil {
			return nil, err
		}
		if len(tags) != len(normalized) {
			return nil, dmarkets.ErrInvalidInput
		}
	}

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("market_id = ?", marketID).Delete(&models.MarketTagAssignment{}).Error; err != nil {
			return err
		}
		for _, tag := range tags {
			assignment := models.MarketTagAssignment{
				MarketID:   marketID,
				TagID:      tag.ID,
				AssignedBy: assignedBy,
				Source:     source,
			}
			assignment.CreatedAt = assignedAt
			assignment.UpdatedAt = assignedAt
			if err := tx.Create(&assignment).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return modelTagsToDomain(tags), nil
}

func (r *GormRepository) SetMarketGroupTags(ctx context.Context, groupID int64, tagSlugs []string, assignedBy string, source string, assignedAt time.Time) ([]dmarkets.MarketTag, error) {
	if groupID <= 0 {
		return nil, dmarkets.ErrInvalidInput
	}
	normalized, err := dmarkets.NormalizeMarketTagSlugs(tagSlugs)
	if err != nil {
		return nil, err
	}

	var tags []models.MarketTag
	if len(normalized) > 0 {
		if err := r.db.WithContext(ctx).
			Where("slug IN ? AND is_active = ?", normalized, true).
			Order("sort_order ASC, display_name ASC, slug ASC").
			Find(&tags).Error; err != nil {
			return nil, err
		}
		if len(tags) != len(normalized) {
			return nil, dmarkets.ErrInvalidInput
		}
	}

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var members []models.MarketGroupMember
		if err := tx.Where("group_id = ?", groupID).Order("display_order ASC, id ASC").Find(&members).Error; err != nil {
			return err
		}
		if len(members) == 0 {
			return dmarkets.ErrMarketGroupNotFound
		}
		marketIDs := make([]int64, 0, len(members))
		for _, member := range members {
			marketIDs = append(marketIDs, member.MarketID)
		}
		if err := tx.Unscoped().Where("market_id IN ?", marketIDs).Delete(&models.MarketTagAssignment{}).Error; err != nil {
			return err
		}
		for _, marketID := range marketIDs {
			for _, tag := range tags {
				assignment := models.MarketTagAssignment{
					MarketID:   marketID,
					TagID:      tag.ID,
					AssignedBy: assignedBy,
					Source:     source,
				}
				assignment.CreatedAt = assignedAt
				assignment.UpdatedAt = assignedAt
				if err := tx.Create(&assignment).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return modelTagsToDomain(tags), nil
}

func (r *GormRepository) hydrateTagsForMarkets(ctx context.Context, markets []*dmarkets.Market) error {
	if len(markets) == 0 {
		return nil
	}

	marketIDs := make([]int64, 0, len(markets))
	marketByID := make(map[int64]*dmarkets.Market, len(markets))
	for _, market := range markets {
		if market == nil {
			continue
		}
		marketIDs = append(marketIDs, market.ID)
		marketByID[market.ID] = market
	}
	if len(marketIDs) == 0 {
		return nil
	}

	rows := []struct {
		MarketID int64
		models.MarketTag
	}{}
	if err := r.db.WithContext(ctx).
		Table("market_tag_assignments").
		Select("market_tag_assignments.market_id, market_tags.*").
		Joins("JOIN market_tags ON market_tags.id = market_tag_assignments.tag_id").
		Where("market_tag_assignments.market_id IN ?", marketIDs).
		Order("market_tags.sort_order ASC, market_tags.display_name ASC, market_tags.slug ASC").
		Scan(&rows).Error; err != nil {
		return err
	}

	for _, row := range rows {
		market := marketByID[row.MarketID]
		if market == nil {
			continue
		}
		tag := modelTagToDomain(row.MarketTag)
		market.Tags = append(market.Tags, tag)
	}
	for _, market := range markets {
		if market != nil && market.Tags == nil {
			market.Tags = []dmarkets.MarketTag{}
		}
	}
	return nil
}

func domainTagToModel(tag dmarkets.MarketTag) models.MarketTag {
	return models.MarketTag{
		ID:          tag.ID,
		Slug:        tag.Slug,
		DisplayName: tag.DisplayName,
		Description: tag.Description,
		ColorKey:    tag.ColorKey,
		SortOrder:   tag.SortOrder,
		IsActive:    tag.IsActive,
		CreatedBy:   tag.CreatedBy,
	}
}

func modelTagToDomain(tag models.MarketTag) dmarkets.MarketTag {
	return dmarkets.MarketTag{
		ID:          tag.ID,
		Slug:        tag.Slug,
		DisplayName: tag.DisplayName,
		Description: tag.Description,
		ColorKey:    tag.ColorKey,
		SortOrder:   tag.SortOrder,
		IsActive:    tag.IsActive,
		CreatedBy:   tag.CreatedBy,
		CreatedAt:   tag.CreatedAt,
		UpdatedAt:   tag.UpdatedAt,
	}
}

func modelTagsToDomain(tags []models.MarketTag) []dmarkets.MarketTag {
	out := make([]dmarkets.MarketTag, 0, len(tags))
	for _, tag := range tags {
		out = append(out, modelTagToDomain(tag))
	}
	return out
}
