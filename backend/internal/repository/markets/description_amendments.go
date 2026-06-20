package markets

import (
	"context"
	"errors"
	"strings"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	return r.listMarketDescriptionAmendments(ctx, filters, true)
}

func (r *GormRepository) ListMarketDescriptionAmendmentReviewCandidates(ctx context.Context, filters dmarkets.AdminDescriptionAmendmentReviewFilters) ([]dmarkets.MarketDescriptionAmendment, int, error) {
	whereSQL, whereArgs := descriptionAmendmentReviewWhere(filters)
	groupSQL := descriptionAmendmentReviewGroupSQL(whereSQL)

	var total int64
	if err := r.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM ("+groupSQL+") review_groups", whereArgs...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []dmarkets.MarketDescriptionAmendment{}, 0, nil
	}
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	pageArgs := append([]any{}, whereArgs...)
	pageArgs = append(pageArgs, filters.Limit, filters.Offset)
	var keys []descriptionAmendmentReviewKeyRow
	if err := r.db.WithContext(ctx).Raw(groupSQL+" ORDER BY sort_created_at DESC, representative_id DESC LIMIT ? OFFSET ?", pageArgs...).Scan(&keys).Error; err != nil {
		return nil, 0, err
	}
	if len(keys) == 0 {
		return []dmarkets.MarketDescriptionAmendment{}, int(total), nil
	}

	rows, err := r.descriptionAmendmentsForReviewKeys(ctx, keys)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dmarkets.MarketDescriptionAmendment, 0, len(rows))
	for _, row := range rows {
		out = append(out, modelDescriptionAmendmentToDomain(row))
	}
	return out, int(total), nil
}

type descriptionAmendmentReviewKeyRow struct {
	RowKind          string `gorm:"column:row_kind"`
	GroupID          int64  `gorm:"column:group_id"`
	SoloAmendmentID  int64  `gorm:"column:solo_amendment_id"`
	RepresentativeID int64  `gorm:"column:representative_id"`
	Version          int    `gorm:"column:version"`
	Status           string `gorm:"column:status"`
	Body             string `gorm:"column:body"`
	CreatedBy        string `gorm:"column:created_by"`
	SubmitReason     string `gorm:"column:submit_reason"`
}

func descriptionAmendmentReviewWhere(filters dmarkets.AdminDescriptionAmendmentReviewFilters) (string, []any) {
	clauses := []string{"a.deleted_at IS NULL", "m.deleted_at IS NULL"}
	args := []any{}
	if filters.MarketID > 0 {
		clauses = append(clauses, "a.market_id = ?")
		args = append(args, filters.MarketID)
	}
	if filters.Status != "" {
		clauses = append(clauses, "a.status = ?")
		args = append(args, dmarkets.NormalizeDescriptionAmendmentStatus(filters.Status))
	}
	if query := strings.TrimSpace(filters.Query); query != "" {
		pattern := "%" + strings.ToLower(query) + "%"
		clauses = append(clauses, `(
			LOWER(COALESCE(m.question_title, '')) LIKE ?
			OR LOWER(COALESCE(m.description, '')) LIKE ?
			OR LOWER(COALESCE(a.body, '')) LIKE ?
			OR LOWER(COALESCE(a.created_by, '')) LIKE ?
			OR LOWER(COALESCE(a.submit_reason, '')) LIKE ?
			OR LOWER(COALESCE(a.rejection_reason, '')) LIKE ?
			OR LOWER(COALESCE(mg.question_title, '')) LIKE ?
			OR LOWER(COALESCE(mg.description, '')) LIKE ?
			OR LOWER(COALESCE(mgm.answer_label, '')) LIKE ?
		)`)
		for i := 0; i < 9; i++ {
			args = append(args, pattern)
		}
	}
	return strings.Join(clauses, " AND "), args
}

func descriptionAmendmentReviewGroupSQL(whereSQL string) string {
	rowKind := "CASE WHEN mgm.group_id IS NULL THEN 'amendment' ELSE 'group' END"
	groupID := "COALESCE(mgm.group_id, 0)"
	soloID := "CASE WHEN mgm.group_id IS NULL THEN a.id ELSE 0 END"
	return `
		SELECT
			` + rowKind + ` AS row_kind,
			` + groupID + ` AS group_id,
			` + soloID + ` AS solo_amendment_id,
			MIN(a.id) AS representative_id,
			a.version AS version,
			a.status AS status,
			a.body AS body,
			a.created_by AS created_by,
			a.submit_reason AS submit_reason,
			MAX(a.created_at) AS sort_created_at
		FROM market_description_amendments a
		JOIN markets m ON m.id = a.market_id
		LEFT JOIN market_group_members mgm ON mgm.market_id = a.market_id AND mgm.deleted_at IS NULL
		LEFT JOIN market_groups mg ON mg.id = mgm.group_id AND mg.deleted_at IS NULL
		WHERE ` + whereSQL + `
		GROUP BY ` + rowKind + `, ` + groupID + `, ` + soloID + `, a.version, a.status, a.body, a.created_by, a.submit_reason
	`
}

func (r *GormRepository) descriptionAmendmentsForReviewKeys(ctx context.Context, keys []descriptionAmendmentReviewKeyRow) ([]models.MarketDescriptionAmendment, error) {
	clauses := make([]string, 0, len(keys))
	args := []any{}
	for _, key := range keys {
		if key.RowKind == "group" && key.GroupID > 0 {
			clauses = append(clauses, "(mgm.group_id = ? AND a.version = ? AND a.status = ? AND a.body = ? AND a.created_by = ? AND a.submit_reason = ?)")
			args = append(args, key.GroupID, key.Version, key.Status, key.Body, key.CreatedBy, key.SubmitReason)
			continue
		}
		clauses = append(clauses, "a.id = ?")
		args = append(args, key.SoloAmendmentID)
	}
	if len(clauses) == 0 {
		return []models.MarketDescriptionAmendment{}, nil
	}
	selectSQL := `
		SELECT a.*
		FROM market_description_amendments a
		LEFT JOIN market_group_members mgm ON mgm.market_id = a.market_id AND mgm.deleted_at IS NULL
		WHERE a.deleted_at IS NULL AND (` + strings.Join(clauses, " OR ") + `)
		ORDER BY a.market_id ASC, a.version ASC, a.id ASC
	`
	var rows []models.MarketDescriptionAmendment
	if err := r.db.WithContext(ctx).Raw(selectSQL, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *GormRepository) listMarketDescriptionAmendments(ctx context.Context, filters dmarkets.MarketDescriptionAmendmentFilters, applyPagination bool) ([]dmarkets.MarketDescriptionAmendment, error) {
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
	if applyPagination && filters.Limit <= 0 {
		filters.Limit = 50
		query = query.Limit(filters.Limit)
	} else if applyPagination {
		query = query.Limit(filters.Limit)
	}
	query = query.Order("market_id ASC, version ASC")
	if applyPagination && filters.Offset > 0 {
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
	updates, ok := descriptionAmendmentReviewUpdates(status, actorUsername, reason, reviewedAt)
	if !ok {
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

func (r *GormRepository) ReviewGroupedMarketDescriptionAmendments(ctx context.Context, ids []int64, status string, actorUsername string, reason string, reviewedAt time.Time) ([]dmarkets.MarketDescriptionAmendment, error) {
	if len(ids) == 0 {
		return nil, dmarkets.ErrInvalidInput
	}
	status = dmarkets.NormalizeDescriptionAmendmentStatus(status)
	updates, ok := descriptionAmendmentReviewUpdates(status, actorUsername, reason, reviewedAt)
	if !ok {
		return nil, dmarkets.ErrInvalidInput
	}

	var updated []models.MarketDescriptionAmendment
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if hasDuplicateInt64(ids) {
			return dmarkets.ErrInvalidInput
		}
		var rows []models.MarketDescriptionAmendment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ?", ids).
			Order("id ASC").
			Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) != len(ids) {
			return dmarkets.ErrMarketNotFound
		}
		groupID, err := groupedDescriptionAmendmentGroupID(tx, rows)
		if err != nil {
			return err
		}
		first := rows[0]
		var fullGroupRows []models.MarketDescriptionAmendment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Table("market_description_amendments AS a").
			Select("a.*").
			Joins("JOIN market_group_members AS mgm ON mgm.market_id = a.market_id AND mgm.deleted_at IS NULL").
			Where("mgm.group_id = ?", groupID).
			Where("a.version = ? AND a.status = ? AND a.body = ? AND a.created_by = ? AND a.submit_reason = ?",
				first.Version,
				dmarkets.DescriptionAmendmentStatusPending,
				first.Body,
				first.CreatedBy,
				first.SubmitReason,
			).
			Where("a.deleted_at IS NULL").
			Order("a.id ASC").
			Find(&fullGroupRows).Error; err != nil {
			return err
		}
		if len(fullGroupRows) == 0 {
			return dmarkets.ErrInvalidState
		}
		if !sameInt64Set(ids, descriptionAmendmentIDs(fullGroupRows)) {
			return dmarkets.ErrInvalidState
		}
		updateIDs := descriptionAmendmentIDs(fullGroupRows)

		result := tx.Model(&models.MarketDescriptionAmendment{}).
			Where("id IN ? AND status = ?", updateIDs, dmarkets.DescriptionAmendmentStatusPending).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != int64(len(updateIDs)) {
			return dmarkets.ErrInvalidState
		}
		return tx.Where("id IN ?", updateIDs).Order("id ASC").Find(&updated).Error
	})
	if err != nil {
		return nil, err
	}
	out := make([]dmarkets.MarketDescriptionAmendment, 0, len(updated))
	for _, row := range updated {
		out = append(out, modelDescriptionAmendmentToDomain(row))
	}
	return out, nil
}

func descriptionAmendmentReviewUpdates(status string, actorUsername string, reason string, reviewedAt time.Time) (map[string]any, bool) {
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
		return updates, true
	}
	if status == dmarkets.DescriptionAmendmentStatusRejected {
		updates["rejected_by"] = actorUsername
		updates["rejected_at"] = reviewedAt
		updates["rejection_reason"] = reason
		updates["approved_by"] = ""
		updates["approved_at"] = nil
		return updates, true
	}
	return nil, false
}

func groupedDescriptionAmendmentGroupID(tx *gorm.DB, rows []models.MarketDescriptionAmendment) (int64, error) {
	if len(rows) == 0 {
		return 0, dmarkets.ErrInvalidInput
	}
	marketIDs := make([]int64, 0, len(rows))
	marketSeen := map[int64]bool{}
	first := rows[0]
	for _, row := range rows {
		if row.Status != dmarkets.DescriptionAmendmentStatusPending {
			return 0, dmarkets.ErrInvalidState
		}
		if row.Version != first.Version || row.Body != first.Body || row.CreatedBy != first.CreatedBy || row.SubmitReason != first.SubmitReason {
			return 0, dmarkets.ErrInvalidInput
		}
		if !marketSeen[row.MarketID] {
			marketSeen[row.MarketID] = true
			marketIDs = append(marketIDs, row.MarketID)
		}
	}
	if len(marketIDs) != len(rows) {
		return 0, dmarkets.ErrInvalidInput
	}

	var members []models.MarketGroupMember
	if err := tx.Where("market_id IN ?", marketIDs).Find(&members).Error; err != nil {
		return 0, err
	}
	if len(members) != len(marketIDs) {
		return 0, dmarkets.ErrInvalidInput
	}
	groupID := int64(0)
	memberMarkets := map[int64]bool{}
	for _, member := range members {
		if groupID == 0 {
			groupID = member.GroupID
		}
		if member.GroupID <= 0 || member.GroupID != groupID {
			return 0, dmarkets.ErrInvalidInput
		}
		memberMarkets[member.MarketID] = true
	}
	for _, marketID := range marketIDs {
		if !memberMarkets[marketID] {
			return 0, dmarkets.ErrInvalidInput
		}
	}
	return groupID, nil
}

func hasDuplicateInt64(values []int64) bool {
	seen := map[int64]bool{}
	for _, value := range values {
		if value <= 0 || seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}

func descriptionAmendmentIDs(rows []models.MarketDescriptionAmendment) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}

func sameInt64Set(a []int64, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[int64]int{}
	for _, value := range a {
		seen[value]++
	}
	for _, value := range b {
		if seen[value] == 0 {
			return false
		}
		seen[value]--
	}
	return true
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
