package markets

import (
	"context"
	"errors"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"

	"gorm.io/gorm"
)

// CreateMarketGroup creates a parent market group and links ordered binary
// child markets. It does not create child markets; callers must create normal
// binary markets through the existing market transaction boundary first.
func (r *GormRepository) CreateMarketGroup(ctx context.Context, group *dmarkets.MarketGroup, members []dmarkets.MarketGroupMember) error {
	if group == nil {
		return dmarkets.ErrInvalidInput
	}
	dmarkets.NormalizeMarketGroupDefaults(group)
	if err := dmarkets.ValidateMarketGroupMembers(members); err != nil {
		return err
	}

	dbGroup := domainMarketGroupToModel(group)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&dbGroup).Error; err != nil {
			return err
		}

		dbMembers := make([]models.MarketGroupMember, 0, len(members))
		for _, member := range members {
			member.GroupID = dbGroup.ID
			dbMembers = append(dbMembers, domainMarketGroupMemberToModel(member))
		}
		if err := tx.Create(&dbMembers).Error; err != nil {
			return err
		}

		group.ID = dbGroup.ID
		group.CreatedAt = dbGroup.CreatedAt
		group.UpdatedAt = dbGroup.UpdatedAt
		group.Members = modelMarketGroupMembersToDomain(dbMembers)
		return nil
	})
}

// GetMarketGroup returns a parent market group with ordered child members.
func (r *GormRepository) GetMarketGroup(ctx context.Context, groupID int64) (*dmarkets.MarketGroup, error) {
	var dbGroup models.MarketGroup
	if err := r.db.WithContext(ctx).First(&dbGroup, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dmarkets.ErrMarketGroupNotFound
		}
		return nil, err
	}

	members, err := r.ListMarketGroupMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	group := modelMarketGroupToDomain(dbGroup)
	group.Members = members
	return &group, nil
}

// ListMarketGroupMembers returns the child market links in display order.
func (r *GormRepository) ListMarketGroupMembers(ctx context.Context, groupID int64) ([]dmarkets.MarketGroupMember, error) {
	var dbMembers []models.MarketGroupMember
	if err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("display_order ASC, id ASC").
		Find(&dbMembers).Error; err != nil {
		return nil, err
	}
	return modelMarketGroupMembersToDomain(dbMembers), nil
}

// MarkMarketGroupResolved terminally marks the display/governance parent after
// all child binary markets have been resolved through normal resolution paths.
func (r *GormRepository) MarkMarketGroupResolved(ctx context.Context, groupID int64, resolvedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&models.MarketGroup{}).
		Where("id = ? AND lifecycle_status = ?", groupID, dmarkets.MarketLifecyclePublished).
		Updates(map[string]any{
			"lifecycle_status": dmarkets.MarketLifecycleResolved,
			"updated_at":       resolvedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return dmarkets.ErrInvalidState
	}
	return nil
}

// UpdateMarketGroupAnswerAdditionAutoApproval changes per-group answer
// addition review policy. It is governance state on the parent group only.
func (r *GormRepository) UpdateMarketGroupAnswerAdditionAutoApproval(ctx context.Context, groupID int64, enabled bool, updatedAt time.Time) (*dmarkets.MarketGroup, error) {
	if groupID <= 0 {
		return nil, dmarkets.ErrInvalidInput
	}
	result := r.db.WithContext(ctx).Model(&models.MarketGroup{}).
		Where("id = ?", groupID).
		Updates(map[string]any{
			"auto_approve_answer_additions": enabled,
			"updated_at":                    updatedAt,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, dmarkets.ErrMarketGroupNotFound
	}
	return r.GetMarketGroup(ctx, groupID)
}

// GetMarketGroupForMarket resolves the parent group for a child market.
func (r *GormRepository) GetMarketGroupForMarket(ctx context.Context, marketID int64) (*dmarkets.MarketGroup, error) {
	var member models.MarketGroupMember
	if err := r.db.WithContext(ctx).
		Where("market_id = ?", marketID).
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dmarkets.ErrMarketGroupNotFound
		}
		return nil, err
	}
	return r.GetMarketGroup(ctx, member.GroupID)
}

// ApproveMarketGroup publishes a proposed parent group and all proposed child
// markets together. Group lifecycle is a governance concern; child markets
// remain the transaction boundary after publication.
func (r *GormRepository) ApproveMarketGroup(ctx context.Context, groupID int64, actorUsername string, approvedAt time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		childIDs, err := marketGroupChildIDs(ctx, tx, groupID)
		if err != nil {
			return err
		}
		if len(childIDs) == 0 {
			return dmarkets.ErrInvalidState
		}

		groupResult := tx.Model(&models.MarketGroup{}).
			Where("id = ? AND lifecycle_status = ?", groupID, dmarkets.MarketLifecycleProposed).
			Updates(map[string]any{
				"lifecycle_status": dmarkets.MarketLifecyclePublished,
				"approved_by":      actorUsername,
				"approved_at":      approvedAt,
				"rejected_by":      "",
				"rejected_at":      nil,
				"rejection_reason": "",
				"updated_at":       approvedAt,
			})
		if groupResult.Error != nil {
			return groupResult.Error
		}
		if groupResult.RowsAffected == 0 {
			return dmarkets.ErrInvalidState
		}

		childResult := tx.Model(&models.Market{}).
			Where("id IN ? AND lifecycle_status = ?", childIDs, dmarkets.MarketLifecycleProposed).
			Updates(map[string]any{
				"lifecycle_status": dmarkets.MarketLifecyclePublished,
				"approved_by":      actorUsername,
				"approved_at":      approvedAt,
				"rejected_by":      "",
				"rejected_at":      nil,
				"rejection_reason": "",
				"updated_at":       approvedAt,
			})
		if childResult.Error != nil {
			return childResult.Error
		}
		if childResult.RowsAffected != int64(len(childIDs)) {
			return dmarkets.ErrInvalidState
		}
		return nil
	})
}

// RejectMarketGroup rejects a proposed parent group and all proposed child
// markets together. The domain service owns the one-time parent cost refund.
func (r *GormRepository) RejectMarketGroup(ctx context.Context, groupID int64, actorUsername string, rejectedAt time.Time, reason string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		childIDs, err := marketGroupChildIDs(ctx, tx, groupID)
		if err != nil {
			return err
		}
		if len(childIDs) == 0 {
			return dmarkets.ErrInvalidState
		}

		groupResult := tx.Model(&models.MarketGroup{}).
			Where("id = ? AND lifecycle_status = ?", groupID, dmarkets.MarketLifecycleProposed).
			Updates(map[string]any{
				"lifecycle_status": dmarkets.MarketLifecycleRejected,
				"rejected_by":      actorUsername,
				"rejected_at":      rejectedAt,
				"rejection_reason": reason,
				"updated_at":       rejectedAt,
			})
		if groupResult.Error != nil {
			return groupResult.Error
		}
		if groupResult.RowsAffected == 0 {
			return dmarkets.ErrInvalidState
		}

		childResult := tx.Model(&models.Market{}).
			Where("id IN ? AND lifecycle_status = ?", childIDs, dmarkets.MarketLifecycleProposed).
			Updates(map[string]any{
				"lifecycle_status": dmarkets.MarketLifecycleRejected,
				"rejected_by":      actorUsername,
				"rejected_at":      rejectedAt,
				"rejection_reason": reason,
				"updated_at":       rejectedAt,
			})
		if childResult.Error != nil {
			return childResult.Error
		}
		if childResult.RowsAffected != int64(len(childIDs)) {
			return dmarkets.ErrInvalidState
		}
		return nil
	})
}

// ReassignMarketGroupSteward changes the parent steward and every child market
// steward in one transaction. Existing child-level audit rows remain the
// traceable source for stewardship changes on operational market records.
func (r *GormRepository) ReassignMarketGroupSteward(ctx context.Context, groupID int64, fromStewardUsername string, toStewardUsername string, actorUsername string, reason string, changedAt time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		childIDs, err := marketGroupChildIDs(ctx, tx, groupID)
		if err != nil {
			return err
		}
		if len(childIDs) == 0 {
			return dmarkets.ErrInvalidState
		}

		groupResult := tx.Model(&models.MarketGroup{}).
			Where("id = ?", groupID).
			Updates(map[string]any{
				"steward_username": toStewardUsername,
				"updated_at":       changedAt,
			})
		if groupResult.Error != nil {
			return groupResult.Error
		}
		if groupResult.RowsAffected == 0 {
			return dmarkets.ErrMarketGroupNotFound
		}

		childResult := tx.Model(&models.Market{}).
			Where("id IN ?", childIDs).
			Updates(map[string]any{
				"steward_username": toStewardUsername,
				"updated_at":       changedAt,
			})
		if childResult.Error != nil {
			return childResult.Error
		}
		if childResult.RowsAffected != int64(len(childIDs)) {
			return dmarkets.ErrInvalidState
		}

		audits := make([]models.MarketStewardshipAudit, 0, len(childIDs))
		for _, childID := range childIDs {
			audit := models.MarketStewardshipAudit{
				MarketID:            childID,
				FromStewardUsername: fromStewardUsername,
				ToStewardUsername:   toStewardUsername,
				ActorUsername:       actorUsername,
				Reason:              reason,
			}
			audit.CreatedAt = changedAt
			audit.UpdatedAt = changedAt
			audits = append(audits, audit)
		}
		return tx.Create(&audits).Error
	})
}

func marketGroupChildIDs(ctx context.Context, tx *gorm.DB, groupID int64) ([]int64, error) {
	var members []models.MarketGroupMember
	if err := tx.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("display_order ASC, id ASC").
		Find(&members).Error; err != nil {
		return nil, err
	}
	childIDs := make([]int64, 0, len(members))
	for _, member := range members {
		childIDs = append(childIDs, member.MarketID)
	}
	return childIDs, nil
}

func domainMarketGroupToModel(group *dmarkets.MarketGroup) models.MarketGroup {
	return models.MarketGroup{
		ID:                         group.ID,
		QuestionTitle:              group.QuestionTitle,
		Description:                group.Description,
		GroupType:                  group.GroupType,
		ProbabilityPolicy:          group.ProbabilityPolicy,
		ResolutionPolicy:           group.ResolutionPolicy,
		LifecycleStatus:            group.LifecycleStatus,
		ProposalCost:               group.ProposalCost,
		CreatorUsername:            group.CreatorUsername,
		StewardUsername:            group.StewardUsername,
		ApprovedBy:                 group.ApprovedBy,
		ApprovedAt:                 copyTimePtr(group.ApprovedAt),
		RejectedBy:                 group.RejectedBy,
		RejectedAt:                 copyTimePtr(group.RejectedAt),
		RejectionReason:            group.RejectionReason,
		ResolutionDateTime:         group.ResolutionDateTime,
		AutoApproveAnswerAdditions: group.AutoApproveAnswerAdditions,
	}
}

func modelMarketGroupToDomain(group models.MarketGroup) dmarkets.MarketGroup {
	return dmarkets.MarketGroup{
		ID:                         group.ID,
		QuestionTitle:              group.QuestionTitle,
		Description:                group.Description,
		GroupType:                  group.GroupType,
		ProbabilityPolicy:          group.ProbabilityPolicy,
		ResolutionPolicy:           group.ResolutionPolicy,
		LifecycleStatus:            dmarkets.NormalizeLifecycleStatus(group.LifecycleStatus),
		ProposalCost:               group.ProposalCost,
		CreatorUsername:            group.CreatorUsername,
		StewardUsername:            group.StewardUsername,
		ApprovedBy:                 group.ApprovedBy,
		ApprovedAt:                 copyTimePtr(group.ApprovedAt),
		RejectedBy:                 group.RejectedBy,
		RejectedAt:                 copyTimePtr(group.RejectedAt),
		RejectionReason:            group.RejectionReason,
		ResolutionDateTime:         group.ResolutionDateTime,
		AutoApproveAnswerAdditions: group.AutoApproveAnswerAdditions,
		CreatedAt:                  group.CreatedAt,
		UpdatedAt:                  group.UpdatedAt,
	}
}

func domainMarketGroupMemberToModel(member dmarkets.MarketGroupMember) models.MarketGroupMember {
	return models.MarketGroupMember{
		ID:           member.ID,
		GroupID:      member.GroupID,
		MarketID:     member.MarketID,
		AnswerLabel:  member.AnswerLabel,
		DisplayOrder: member.DisplayOrder,
	}
}

func modelMarketGroupMemberToDomain(member models.MarketGroupMember) dmarkets.MarketGroupMember {
	return dmarkets.MarketGroupMember{
		ID:           member.ID,
		GroupID:      member.GroupID,
		MarketID:     member.MarketID,
		AnswerLabel:  member.AnswerLabel,
		DisplayOrder: member.DisplayOrder,
		CreatedAt:    member.CreatedAt,
		UpdatedAt:    member.UpdatedAt,
	}
}

func modelMarketGroupMembersToDomain(members []models.MarketGroupMember) []dmarkets.MarketGroupMember {
	out := make([]dmarkets.MarketGroupMember, 0, len(members))
	for _, member := range members {
		out = append(out, modelMarketGroupMemberToDomain(member))
	}
	return dmarkets.OrderedMarketGroupMembers(out)
}

func copyTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
