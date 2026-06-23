package markets

import (
	"context"
	"errors"
	"strings"
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
	query := r.marketGroupAnswerAdditionBaseQuery(ctx, dmarkets.AdminAnswerAdditionReviewFilters{
		GroupID: filters.GroupID,
		Status:  filters.Status,
	})
	if filters.ProposedBy != "" {
		query = query.Where("proposed_by = ?", filters.ProposedBy)
	}
	if filters.ReviewerUsername != "" {
		query = query.Joins("JOIN market_groups ON market_groups.id = market_group_answer_additions.group_id").
			Where("COALESCE(NULLIF(market_groups.steward_username, ''), market_groups.creator_username) = ?", filters.ReviewerUsername)
	}
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	query = query.Limit(filters.Limit).Order("created_at DESC, id DESC")
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}
	return r.findMarketGroupAnswerAdditions(ctx, query)
}

func (r *GormRepository) ListMarketGroupAnswerAdditionsForAdminReview(ctx context.Context, filters dmarkets.AdminAnswerAdditionReviewFilters) ([]dmarkets.MarketGroupAnswerAddition, int, error) {
	query := r.marketGroupAnswerAdditionBaseQuery(ctx, filters)
	query = applyMarketGroupAnswerAdditionReviewSearch(query, filters.Query)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	query = query.Order("market_group_answer_additions.created_at DESC, market_group_answer_additions.id DESC").Limit(filters.Limit)
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}
	items, err := r.findMarketGroupAnswerAdditions(ctx, query)
	return items, int(total), err
}

func (r *GormRepository) marketGroupAnswerAdditionBaseQuery(ctx context.Context, filters dmarkets.AdminAnswerAdditionReviewFilters) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&models.MarketGroupAnswerAddition{})
	if filters.GroupID > 0 {
		query = query.Where("group_id = ?", filters.GroupID)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", dmarkets.NormalizeMarketGroupAnswerAdditionStatus(filters.Status))
	}
	return query
}

func applyMarketGroupAnswerAdditionReviewSearch(query *gorm.DB, value string) *gorm.DB {
	term := strings.TrimSpace(value)
	if term == "" {
		return query
	}
	pattern := "%" + strings.ToLower(term) + "%"
	return query.
		Joins("LEFT JOIN market_groups answer_addition_search_groups ON answer_addition_search_groups.id = market_group_answer_additions.group_id").
		Where(`(
			LOWER(market_group_answer_additions.answer_label) LIKE ?
			OR LOWER(market_group_answer_additions.proposed_by) LIKE ?
			OR LOWER(market_group_answer_additions.reviewed_by) LIKE ?
			OR LOWER(market_group_answer_additions.rejection_reason) LIKE ?
			OR LOWER(answer_addition_search_groups.question_title) LIKE ?
			OR LOWER(answer_addition_search_groups.description) LIKE ?
			OR LOWER(answer_addition_search_groups.creator_username) LIKE ?
			OR LOWER(answer_addition_search_groups.steward_username) LIKE ?
		)`, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern)
}

func (r *GormRepository) findMarketGroupAnswerAdditions(ctx context.Context, query *gorm.DB) ([]dmarkets.MarketGroupAnswerAddition, error) {
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
		byID[group.ID] = group
	}
	var memberRows []models.MarketGroupMember
	if err := r.db.WithContext(ctx).
		Where("group_id IN ?", groupIDs).
		Order("group_id ASC, display_order ASC, id ASC").
		Find(&memberRows).Error; err != nil {
		return additions
	}
	membersByGroup := map[int64][]dmarkets.MarketGroupMember{}
	for _, row := range memberRows {
		member := modelMarketGroupMemberToDomain(row)
		membersByGroup[member.GroupID] = append(membersByGroup[member.GroupID], member)
	}
	for groupID, members := range membersByGroup {
		group, ok := byID[groupID]
		if !ok {
			continue
		}
		group.Members = dmarkets.OrderedMarketGroupMembers(members)
		byID[groupID] = group
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
