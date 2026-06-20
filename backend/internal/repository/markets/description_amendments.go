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

func (r *GormRepository) CreateMarketDescriptionAmendment(ctx context.Context, amendment dmarkets.MarketDescriptionAmendment) (*dmarkets.MarketDescriptionAmendment, error) {
	if amendment.MarketID <= 0 {
		return nil, dmarkets.ErrInvalidInput
	}
	var created models.MarketDescriptionAmendment
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var maxVersion int
		if err := tx.Model(&models.MarketDescriptionAmendment{}).
			Where("market_id = ?", amendment.MarketID).
			Select("COALESCE(MAX(version), 1)").
			Scan(&maxVersion).Error; err != nil {
			return err
		}
		created = domainDescriptionAmendmentToModel(amendment)
		created.Version = maxVersion + 1
		if created.Version < 2 {
			created.Version = 2
		}
		return tx.Create(&created).Error
	})
	if err != nil {
		return nil, err
	}
	out := modelDescriptionAmendmentToDomain(created)
	return &out, nil
}

func (r *GormRepository) ListMarketDescriptionAmendments(ctx context.Context, filters dmarkets.MarketDescriptionAmendmentFilters) ([]dmarkets.MarketDescriptionAmendment, error) {
	query := r.db.WithContext(ctx).Model(&models.MarketDescriptionAmendment{})
	if filters.MarketID > 0 {
		query = query.Where("market_id = ?", filters.MarketID)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", dmarkets.NormalizeDescriptionAmendmentStatus(filters.Status))
	}
	if filters.CreatedBy != "" {
		query = query.Where("created_by = ?", filters.CreatedBy)
	}
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	query = query.Order("market_id ASC, version ASC").Limit(filters.Limit)
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var rows []models.MarketDescriptionAmendment
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dmarkets.MarketDescriptionAmendment, 0, len(rows))
	for _, row := range rows {
		out = append(out, modelDescriptionAmendmentToDomain(row))
	}
	return out, nil
}

func (r *GormRepository) ReviewMarketDescriptionAmendment(ctx context.Context, id int64, status string, actorUsername string, reason string, reviewedAt time.Time) (*dmarkets.MarketDescriptionAmendment, error) {
	if id <= 0 {
		return nil, dmarkets.ErrInvalidInput
	}
	status = dmarkets.NormalizeDescriptionAmendmentStatus(status)
	updates := map[string]any{
		"status":     status,
		"updated_at": reviewedAt,
	}
	if status == dmarkets.DescriptionAmendmentStatusApproved {
		updates["approved_by"] = actorUsername
		updates["approved_at"] = reviewedAt
		updates["rejected_by"] = ""
		updates["rejected_at"] = nil
		updates["rejection_reason"] = ""
	} else if status == dmarkets.DescriptionAmendmentStatusRejected {
		updates["rejected_by"] = actorUsername
		updates["rejected_at"] = reviewedAt
		updates["rejection_reason"] = reason
		updates["approved_by"] = ""
		updates["approved_at"] = nil
	} else {
		return nil, dmarkets.ErrInvalidInput
	}

	result := r.db.WithContext(ctx).Model(&models.MarketDescriptionAmendment{}).
		Where("id = ? AND status = ?", id, dmarkets.DescriptionAmendmentStatusPending).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		var existing models.MarketDescriptionAmendment
		if err := r.db.WithContext(ctx).First(&existing, id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dmarkets.ErrMarketNotFound
		} else if err != nil {
			return nil, err
		}
		return nil, dmarkets.ErrInvalidState
	}

	var row models.MarketDescriptionAmendment
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	out := modelDescriptionAmendmentToDomain(row)
	return &out, nil
}

func (r *GormRepository) GetMarketGovernanceSettings(ctx context.Context) (*dmarkets.MarketGovernanceSettings, error) {
	var row models.MarketGovernanceSettings
	if err := r.db.WithContext(ctx).First(&row, 1).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dmarkets.MarketGovernanceSettings{
				MarketGroupAnswerAdditionApprovalPolicy: dmarkets.MarketGroupAnswerAdditionApprovalPolicyModerator,
				Version:                                 1,
			}, nil
		}
		return nil, err
	}
	out := modelMarketGovernanceSettingsToDomain(row)
	return &out, nil
}

func (r *GormRepository) UpdateMarketGovernanceSettings(ctx context.Context, update dmarkets.MarketGovernanceSettingsUpdate) (*dmarkets.MarketGovernanceSettings, error) {
	if update.AutoApproveDescriptionAmendments == nil &&
		update.AutoApproveMarketProposals == nil &&
		update.AutoApproveMarketGroupAnswers == nil &&
		update.MarketGroupAnswerAdditionApprovalPolicy == nil {
		return nil, dmarkets.ErrInvalidInput
	}
	var saved models.MarketGovernanceSettings
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row models.MarketGovernanceSettings
		err := tx.First(&row, 1).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row = models.MarketGovernanceSettings{
				ID:                                      1,
				Version:                                 1,
				MarketGroupAnswerAdditionApprovalPolicy: dmarkets.MarketGroupAnswerAdditionApprovalPolicyModerator,
			}
		} else if err != nil {
			return err
		} else if update.Version != 0 && update.Version != row.Version {
			return dmarkets.ErrInvalidState
		}
		if update.AutoApproveDescriptionAmendments != nil {
			row.AutoApproveDescriptionAmendments = *update.AutoApproveDescriptionAmendments
		}
		if update.AutoApproveMarketProposals != nil {
			row.AutoApproveMarketProposals = *update.AutoApproveMarketProposals
		}
		if update.AutoApproveMarketGroupAnswers != nil {
			row.AutoApproveMarketGroupAnswers = *update.AutoApproveMarketGroupAnswers
			if update.MarketGroupAnswerAdditionApprovalPolicy == nil {
				row.MarketGroupAnswerAdditionApprovalPolicy = marketGroupAnswerAdditionApprovalPolicyFromLegacy(*update.AutoApproveMarketGroupAnswers)
			}
		}
		if update.MarketGroupAnswerAdditionApprovalPolicy != nil {
			policy := dmarkets.NormalizeMarketGroupAnswerAdditionApprovalPolicy(*update.MarketGroupAnswerAdditionApprovalPolicy)
			if !dmarkets.IsValidMarketGroupAnswerAdditionApprovalPolicy(policy) {
				return dmarkets.ErrInvalidInput
			}
			row.MarketGroupAnswerAdditionApprovalPolicy = policy
			row.AutoApproveMarketGroupAnswers = policy == dmarkets.MarketGroupAnswerAdditionApprovalPolicyAuto
		}
		if strings.TrimSpace(row.MarketGroupAnswerAdditionApprovalPolicy) == "" {
			row.MarketGroupAnswerAdditionApprovalPolicy = marketGroupAnswerAdditionApprovalPolicyFromLegacy(row.AutoApproveMarketGroupAnswers)
		}
		row.UpdatedBy = update.UpdatedBy
		if row.ID == 0 {
			row.ID = 1
			row.Version = 1
		} else if !row.CreatedAt.IsZero() {
			row.Version++
		}
		if row.Version == 0 {
			row.Version = 1
		}
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		saved = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := modelMarketGovernanceSettingsToDomain(saved)
	return &out, nil
}

func domainDescriptionAmendmentToModel(amendment dmarkets.MarketDescriptionAmendment) models.MarketDescriptionAmendment {
	return models.MarketDescriptionAmendment{
		ID:              amendment.ID,
		MarketID:        amendment.MarketID,
		Version:         amendment.Version,
		Body:            amendment.Body,
		BodyFormat:      amendment.BodyFormat,
		Status:          dmarkets.NormalizeDescriptionAmendmentStatus(amendment.Status),
		CreatedBy:       amendment.CreatedBy,
		ApprovedBy:      amendment.ApprovedBy,
		ApprovedAt:      amendment.ApprovedAt,
		RejectedBy:      amendment.RejectedBy,
		RejectedAt:      amendment.RejectedAt,
		RejectionReason: amendment.RejectionReason,
		SubmitReason:    amendment.SubmitReason,
	}
}

func modelMarketGovernanceSettingsToDomain(row models.MarketGovernanceSettings) dmarkets.MarketGovernanceSettings {
	policy := dmarkets.NormalizeMarketGroupAnswerAdditionApprovalPolicy(row.MarketGroupAnswerAdditionApprovalPolicy)
	if strings.TrimSpace(row.MarketGroupAnswerAdditionApprovalPolicy) == "" {
		policy = marketGroupAnswerAdditionApprovalPolicyFromLegacy(row.AutoApproveMarketGroupAnswers)
	}
	return dmarkets.MarketGovernanceSettings{
		AutoApproveDescriptionAmendments:        row.AutoApproveDescriptionAmendments,
		AutoApproveMarketProposals:              row.AutoApproveMarketProposals,
		AutoApproveMarketGroupAnswers:           policy == dmarkets.MarketGroupAnswerAdditionApprovalPolicyAuto,
		MarketGroupAnswerAdditionApprovalPolicy: policy,
		Version:                                 row.Version,
		UpdatedBy:                               row.UpdatedBy,
		UpdatedAt:                               row.UpdatedAt,
	}
}

func marketGroupAnswerAdditionApprovalPolicyFromLegacy(enabled bool) string {
	if enabled {
		return dmarkets.MarketGroupAnswerAdditionApprovalPolicyAuto
	}
	return dmarkets.MarketGroupAnswerAdditionApprovalPolicyModerator
}

func modelDescriptionAmendmentToDomain(row models.MarketDescriptionAmendment) dmarkets.MarketDescriptionAmendment {
	return dmarkets.MarketDescriptionAmendment{
		ID:              row.ID,
		MarketID:        row.MarketID,
		Version:         row.Version,
		Body:            row.Body,
		BodyFormat:      row.BodyFormat,
		Status:          row.Status,
		CreatedBy:       row.CreatedBy,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		ApprovedBy:      row.ApprovedBy,
		ApprovedAt:      cloneTimePtr(row.ApprovedAt),
		RejectedBy:      row.RejectedBy,
		RejectedAt:      cloneTimePtr(row.RejectedAt),
		RejectionReason: row.RejectionReason,
		SubmitReason:    row.SubmitReason,
	}
}
