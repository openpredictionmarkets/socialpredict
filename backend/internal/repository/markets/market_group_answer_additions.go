package markets

import (
	"context"
	"errors"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"

	"gorm.io/gorm"
)

func (r *GormRepository) CreateMarketGroupAnswerAddition(ctx context.Context, addition dmarkets.MarketGroupAnswerAddition) (*dmarkets.MarketGroupAnswerAddition, error) {
	if addition.GroupID <= 0 || addition.AnswerLabel == "" || addition.ProposedBy == "" {
		return nil, dmarkets.ErrInvalidInput
	}
	row := domainAnswerAdditionToModel(addition)
	if row.Status == "" {
		row.Status = dmarkets.MarketGroupAnswerAdditionStatusPending
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	out := modelAnswerAdditionToDomain(row)
	return &out, nil
}

func (r *GormRepository) GetMarketGroupAnswerAddition(ctx context.Context, id int64) (*dmarkets.MarketGroupAnswerAddition, error) {
	if id <= 0 {
		return nil, dmarkets.ErrInvalidInput
	}
	var row models.MarketGroupAnswerAddition
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dmarkets.ErrMarketGroupNotFound
		}
		return nil, err
	}
	out := modelAnswerAdditionToDomain(row)
	if group, err := r.GetMarketGroup(ctx, out.GroupID); err == nil {
		out.MarketGroup = group
		out.GroupTitle = group.QuestionTitle
	}
	return &out, nil
}

func (r *GormRepository) ListMarketGroupAnswerAdditions(ctx context.Context, filters dmarkets.MarketGroupAnswerAdditionFilters) ([]dmarkets.MarketGroupAnswerAddition, error) {
	query := r.db.WithContext(ctx).Model(&models.MarketGroupAnswerAddition{})
	if filters.GroupID > 0 {
		query = query.Where("group_id = ?", filters.GroupID)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", dmarkets.NormalizeMarketGroupAnswerAdditionStatus(filters.Status))
	}
	if filters.ProposedBy != "" {
		query = query.Where("proposed_by = ?", filters.ProposedBy)
	}
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	query = query.Order("created_at DESC, id DESC").Limit(filters.Limit)
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var rows []models.MarketGroupAnswerAddition
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dmarkets.MarketGroupAnswerAddition, 0, len(rows))
	for _, row := range rows {
		out = append(out, modelAnswerAdditionToDomain(row))
	}
	return r.hydrateAnswerAdditionGroups(ctx, out), nil
}

func (r *GormRepository) ReviewMarketGroupAnswerAddition(ctx context.Context, id int64, status string, marketID int64, actorUsername string, reason string, reviewedAt time.Time) (*dmarkets.MarketGroupAnswerAddition, error) {
	if id <= 0 {
		return nil, dmarkets.ErrInvalidInput
	}
	status = dmarkets.NormalizeMarketGroupAnswerAdditionStatus(status)
	updates := map[string]any{
		"status":      status,
		"reviewed_by": actorUsername,
		"reviewed_at": reviewedAt,
		"updated_at":  reviewedAt,
	}
	switch status {
	case dmarkets.MarketGroupAnswerAdditionStatusApproved:
		if marketID <= 0 {
			return nil, dmarkets.ErrInvalidInput
		}
		updates["market_id"] = marketID
		updates["rejection_reason"] = ""
	case dmarkets.MarketGroupAnswerAdditionStatusRejected:
		updates["market_id"] = 0
		updates["rejection_reason"] = reason
	default:
		return nil, dmarkets.ErrInvalidInput
	}

	result := r.db.WithContext(ctx).Model(&models.MarketGroupAnswerAddition{}).
		Where("id = ? AND status = ?", id, dmarkets.MarketGroupAnswerAdditionStatusPending).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		var existing models.MarketGroupAnswerAddition
		if err := r.db.WithContext(ctx).First(&existing, id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dmarkets.ErrMarketGroupNotFound
		} else if err != nil {
			return nil, err
		}
		return nil, dmarkets.ErrInvalidState
	}
	return r.GetMarketGroupAnswerAddition(ctx, id)
}

func (r *GormRepository) AddMarketGroupMember(ctx context.Context, groupID int64, member dmarkets.MarketGroupMember) (*dmarkets.MarketGroupMember, error) {
	if groupID <= 0 || member.MarketID <= 0 {
		return nil, dmarkets.ErrInvalidInput
	}
	member.GroupID = groupID
	row := domainMarketGroupMemberToModel(member)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	out := modelMarketGroupMemberToDomain(row)
	return &out, nil
}

func (r *GormRepository) hydrateAnswerAdditionGroups(ctx context.Context, additions []dmarkets.MarketGroupAnswerAddition) []dmarkets.MarketGroupAnswerAddition {
	if len(additions) == 0 {
		return additions
	}
	groupIDs := make([]int64, 0, len(additions))
	seen := map[int64]bool{}
	for _, addition := range additions {
		if addition.GroupID > 0 && !seen[addition.GroupID] {
			groupIDs = append(groupIDs, addition.GroupID)
			seen[addition.GroupID] = true
		}
	}
	if len(groupIDs) == 0 {
		return additions
	}
	var rows []models.MarketGroup
	if err := r.db.WithContext(ctx).Where("id IN ?", groupIDs).Find(&rows).Error; err != nil {
		return additions
	}
	byID := make(map[int64]dmarkets.MarketGroup, len(rows))
	for _, row := range rows {
		group := modelMarketGroupToDomain(row)
		members, err := r.ListMarketGroupMembers(ctx, group.ID)
		if err == nil {
			group.Members = members
		}
		byID[group.ID] = group
	}
	for index := range additions {
		group, ok := byID[additions[index].GroupID]
		if !ok {
			continue
		}
		additions[index].GroupTitle = group.QuestionTitle
		groupCopy := group
		additions[index].MarketGroup = &groupCopy
	}
	return additions
}

func domainAnswerAdditionToModel(addition dmarkets.MarketGroupAnswerAddition) models.MarketGroupAnswerAddition {
	return models.MarketGroupAnswerAddition{
		ID:              addition.ID,
		GroupID:         addition.GroupID,
		MarketID:        addition.MarketID,
		AnswerLabel:     addition.AnswerLabel,
		Status:          dmarkets.NormalizeMarketGroupAnswerAdditionStatus(addition.Status),
		ProposedBy:      addition.ProposedBy,
		ReviewedBy:      addition.ReviewedBy,
		ReviewedAt:      copyTimePtr(addition.ReviewedAt),
		RejectionReason: addition.RejectionReason,
		AdditionCost:    addition.AdditionCost,
	}
}

func modelAnswerAdditionToDomain(row models.MarketGroupAnswerAddition) dmarkets.MarketGroupAnswerAddition {
	return dmarkets.MarketGroupAnswerAddition{
		ID:              row.ID,
		GroupID:         row.GroupID,
		MarketID:        row.MarketID,
		AnswerLabel:     row.AnswerLabel,
		Status:          dmarkets.NormalizeMarketGroupAnswerAdditionStatus(row.Status),
		ProposedBy:      row.ProposedBy,
		ReviewedBy:      row.ReviewedBy,
		ReviewedAt:      copyTimePtr(row.ReviewedAt),
		RejectionReason: row.RejectionReason,
		AdditionCost:    row.AdditionCost,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}
