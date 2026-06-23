package markets

import (
	"context"
	"strconv"
	"strings"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"
)

func (r *GormRepository) ListAdminMarketReviewRows(ctx context.Context, filters dmarkets.AdminMarketReviewFilters) (*dmarkets.AdminMarketReviewPage, error) {
	whereSQL, whereArgs := adminMarketReviewWhere(filters, time.Now())
	groupSQL := adminMarketReviewGroupSQL(whereSQL)

	var total int64
	if err := r.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM ("+groupSQL+") admin_review_rows", whereArgs...).Scan(&total).Error; err != nil {
		return nil, err
	}
	if total == 0 {
		return &dmarkets.AdminMarketReviewPage{Rows: []dmarkets.AdminMarketReviewRow{}, Total: 0, Limit: filters.Limit, Offset: filters.Offset}, nil
	}
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	pageArgs := append([]any{}, whereArgs...)
	pageArgs = append(pageArgs, filters.Limit, filters.Offset)
	var keys []adminMarketReviewKeyRow
	if err := r.db.WithContext(ctx).Raw(groupSQL+" ORDER BY sort_created_at DESC, representative_market_id DESC LIMIT ? OFFSET ?", pageArgs...).Scan(&keys).Error; err != nil {
		return nil, err
	}
	rows, err := r.adminMarketReviewRowsForKeys(ctx, keys)
	if err != nil {
		return nil, err
	}
	return &dmarkets.AdminMarketReviewPage{Rows: rows, Total: int(total), Limit: filters.Limit, Offset: filters.Offset}, nil
}

type adminMarketReviewKeyRow struct {
	RowKind                string `gorm:"column:row_kind"`
	GroupID                int64  `gorm:"column:group_id"`
	MarketID               int64  `gorm:"column:market_id"`
	RepresentativeMarketID int64  `gorm:"column:representative_market_id"`
}

func adminMarketReviewGroupSQL(whereSQL string) string {
	rowKind := "CASE WHEN mg.id IS NULL THEN 'market' ELSE 'group' END"
	groupID := "COALESCE(mg.id, 0)"
	marketID := "CASE WHEN mg.id IS NULL THEN markets.id ELSE 0 END"
	return `
		SELECT
			` + rowKind + ` AS row_kind,
			` + groupID + ` AS group_id,
			` + marketID + ` AS market_id,
			MIN(markets.id) AS representative_market_id,
			MAX(COALESCE(mg.created_at, markets.created_at)) AS sort_created_at
		FROM markets
		LEFT JOIN market_group_members mgm ON mgm.market_id = markets.id AND mgm.deleted_at IS NULL
		LEFT JOIN market_groups mg ON mg.id = mgm.group_id AND mg.deleted_at IS NULL
		WHERE ` + whereSQL + `
		GROUP BY ` + rowKind + `, ` + groupID + `, ` + marketID + `
	`
}

func adminMarketReviewWhere(filters dmarkets.AdminMarketReviewFilters, now time.Time) (string, []any) {
	clauses := []string{"markets.deleted_at IS NULL"}
	args := []any{}
	lifecycleExpr := "COALESCE(NULLIF(mg.lifecycle_status, ''), NULLIF(markets.lifecycle_status, ''), 'published')"
	resolutionTimeExpr := "COALESCE(mg.resolution_date_time, markets.resolution_date_time)"

	switch dmarkets.NormalizeLifecycleStatus(filters.Status) {
	case dmarkets.MarketStatusAll:
		clauses = append(clauses, lifecycleExpr+" IN ?")
		args = append(args, []string{
			dmarkets.MarketLifecycleProposed,
			dmarkets.MarketLifecyclePublished,
			dmarkets.MarketLifecycleClosed,
			dmarkets.MarketLifecycleResolved,
		})
	case dmarkets.MarketLifecycleClosed:
		clauses = append(clauses, "("+lifecycleExpr+" = ? OR ("+lifecycleExpr+" = ? AND markets.is_resolved = ? AND "+resolutionTimeExpr+" <= ?))")
		args = append(args, dmarkets.MarketLifecycleClosed, dmarkets.MarketLifecyclePublished, false, now)
	case dmarkets.MarketLifecycleResolved:
		clauses = append(clauses, "(markets.is_resolved = ? OR "+lifecycleExpr+" = ?)")
		args = append(args, true, dmarkets.MarketLifecycleResolved)
	default:
		clauses = append(clauses, lifecycleExpr+" = ?")
		args = append(args, dmarkets.NormalizeLifecycleStatus(filters.Status))
	}

	if query := strings.TrimSpace(filters.Query); query != "" {
		pattern := "%" + strings.ToLower(query) + "%"
		clauses = append(clauses, `(
			LOWER(COALESCE(markets.question_title, '')) LIKE ?
			OR LOWER(COALESCE(markets.description, '')) LIKE ?
			OR LOWER(COALESCE(markets.creator_username, '')) LIKE ?
			OR LOWER(COALESCE(markets.steward_username, '')) LIKE ?
			OR LOWER(COALESCE(mg.question_title, '')) LIKE ?
			OR LOWER(COALESCE(mg.description, '')) LIKE ?
			OR LOWER(COALESCE(mg.creator_username, '')) LIKE ?
			OR LOWER(COALESCE(mg.steward_username, '')) LIKE ?
			OR LOWER(COALESCE(mgm.answer_label, '')) LIKE ?
		)`)
		for i := 0; i < 9; i++ {
			args = append(args, pattern)
		}
	}
	return strings.Join(clauses, " AND "), args
}

func (r *GormRepository) adminMarketReviewRowsForKeys(ctx context.Context, keys []adminMarketReviewKeyRow) ([]dmarkets.AdminMarketReviewRow, error) {
	if len(keys) == 0 {
		return []dmarkets.AdminMarketReviewRow{}, nil
	}
	groupIDs := []int64{}
	groupSeen := map[int64]bool{}
	marketIDs := []int64{}
	for _, key := range keys {
		if key.RowKind == "group" && key.GroupID > 0 {
			if !groupSeen[key.GroupID] {
				groupSeen[key.GroupID] = true
				groupIDs = append(groupIDs, key.GroupID)
			}
			continue
		}
		if key.MarketID > 0 {
			marketIDs = append(marketIDs, key.MarketID)
		}
	}

	soloByID, err := r.adminReviewMarketsByID(ctx, marketIDs)
	if err != nil {
		return nil, err
	}
	groupByID, childrenByGroup, err := r.adminReviewGroupsByID(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	rows := make([]dmarkets.AdminMarketReviewRow, 0, len(keys))
	for _, key := range keys {
		if key.RowKind == "group" && key.GroupID > 0 {
			group := groupByID[key.GroupID]
			children := childrenByGroup[key.GroupID]
			if group == nil || len(children) == 0 {
				continue
			}
			rows = append(rows, dmarkets.AdminMarketReviewRow{
				RowKey:        "group:" + strconv.FormatInt(key.GroupID, 10),
				IsMarketGroup: true,
				Market:        firstDiscoveryChild(children),
				Group:         group,
				Children:      children,
				Tags:          adminReviewUniqueTags(children),
			})
			continue
		}
		market := soloByID[key.MarketID]
		if market == nil {
			continue
		}
		rows = append(rows, dmarkets.AdminMarketReviewRow{
			RowKey: "market:" + strconv.FormatInt(market.ID, 10),
			Market: market,
			Tags:   market.Tags,
		})
	}
	return rows, nil
}

func (r *GormRepository) adminReviewMarketsByID(ctx context.Context, marketIDs []int64) (map[int64]*dmarkets.Market, error) {
	out := map[int64]*dmarkets.Market{}
	if len(marketIDs) == 0 {
		return out, nil
	}
	var dbMarkets []models.Market
	if err := r.db.WithContext(ctx).Where("id IN ?", marketIDs).Find(&dbMarkets).Error; err != nil {
		return nil, err
	}
	markets := r.mapMarkets(dbMarkets)
	if err := r.hydrateTagsForMarkets(ctx, markets); err != nil {
		return nil, err
	}
	if err := r.hydrateStewardshipAudits(ctx, markets); err != nil {
		return nil, err
	}
	for _, market := range markets {
		if market != nil {
			out[market.ID] = market
		}
	}
	return out, nil
}

func (r *GormRepository) adminReviewGroupsByID(ctx context.Context, groupIDs []int64) (map[int64]*dmarkets.MarketGroup, map[int64][]*dmarkets.Market, error) {
	groupByID := map[int64]*dmarkets.MarketGroup{}
	childrenByGroup := map[int64][]*dmarkets.Market{}
	if len(groupIDs) == 0 {
		return groupByID, childrenByGroup, nil
	}

	var dbGroups []models.MarketGroup
	if err := r.db.WithContext(ctx).Where("id IN ?", groupIDs).Find(&dbGroups).Error; err != nil {
		return nil, nil, err
	}
	for _, dbGroup := range dbGroups {
		group := modelMarketGroupToDomain(dbGroup)
		groupByID[group.ID] = &group
	}

	var dbMembers []models.MarketGroupMember
	if err := r.db.WithContext(ctx).
		Where("group_id IN ?", groupIDs).
		Order("group_id ASC, display_order ASC, id ASC").
		Find(&dbMembers).Error; err != nil {
		return nil, nil, err
	}
	membersByGroup := map[int64][]dmarkets.MarketGroupMember{}
	childIDs := []int64{}
	for _, dbMember := range dbMembers {
		member := modelMarketGroupMemberToDomain(dbMember)
		membersByGroup[member.GroupID] = append(membersByGroup[member.GroupID], member)
		childIDs = append(childIDs, member.MarketID)
	}
	for groupID, members := range membersByGroup {
		if group := groupByID[groupID]; group != nil {
			group.Members = dmarkets.OrderedMarketGroupMembers(members)
		}
	}

	childByID, err := r.adminReviewMarketsByID(ctx, childIDs)
	if err != nil {
		return nil, nil, err
	}
	for groupID, members := range membersByGroup {
		for _, member := range dmarkets.OrderedMarketGroupMembers(members) {
			if child := childByID[member.MarketID]; child != nil {
				childrenByGroup[groupID] = append(childrenByGroup[groupID], child)
			}
		}
	}
	return groupByID, childrenByGroup, nil
}

func adminReviewUniqueTags(markets []*dmarkets.Market) []dmarkets.MarketTag {
	seen := map[string]bool{}
	tags := []dmarkets.MarketTag{}
	for _, market := range markets {
		if market == nil {
			continue
		}
		for _, tag := range market.Tags {
			key := tag.Slug
			if key == "" {
				key = strconv.FormatInt(tag.ID, 10)
			}
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			tags = append(tags, tag)
		}
	}
	return tags
}
