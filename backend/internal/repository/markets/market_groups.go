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

func domainMarketGroupToModel(group *dmarkets.MarketGroup) models.MarketGroup {
	return models.MarketGroup{
		ID:                 group.ID,
		QuestionTitle:      group.QuestionTitle,
		Description:        group.Description,
		GroupType:          group.GroupType,
		ProbabilityPolicy:  group.ProbabilityPolicy,
		ResolutionPolicy:   group.ResolutionPolicy,
		LifecycleStatus:    group.LifecycleStatus,
		ProposalCost:       group.ProposalCost,
		CreatorUsername:    group.CreatorUsername,
		StewardUsername:    group.StewardUsername,
		ApprovedBy:         group.ApprovedBy,
		ApprovedAt:         copyTimePtr(group.ApprovedAt),
		RejectedBy:         group.RejectedBy,
		RejectedAt:         copyTimePtr(group.RejectedAt),
		RejectionReason:    group.RejectionReason,
		ResolutionDateTime: group.ResolutionDateTime,
	}
}

func modelMarketGroupToDomain(group models.MarketGroup) dmarkets.MarketGroup {
	return dmarkets.MarketGroup{
		ID:                 group.ID,
		QuestionTitle:      group.QuestionTitle,
		Description:        group.Description,
		GroupType:          group.GroupType,
		ProbabilityPolicy:  group.ProbabilityPolicy,
		ResolutionPolicy:   group.ResolutionPolicy,
		LifecycleStatus:    dmarkets.NormalizeLifecycleStatus(group.LifecycleStatus),
		ProposalCost:       group.ProposalCost,
		CreatorUsername:    group.CreatorUsername,
		StewardUsername:    group.StewardUsername,
		ApprovedBy:         group.ApprovedBy,
		ApprovedAt:         copyTimePtr(group.ApprovedAt),
		RejectedBy:         group.RejectedBy,
		RejectedAt:         copyTimePtr(group.RejectedAt),
		RejectionReason:    group.RejectionReason,
		ResolutionDateTime: group.ResolutionDateTime,
		CreatedAt:          group.CreatedAt,
		UpdatedAt:          group.UpdatedAt,
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
