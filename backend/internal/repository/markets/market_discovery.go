package markets

import (
	"context"
	"strings"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"

	"gorm.io/gorm"
)

var _ dmarkets.MarketDiscoveryRepository = (*GormRepository)(nil)

// ListMarketDiscovery returns public discovery rows grouped before pagination.
func (r *GormRepository) ListMarketDiscovery(ctx context.Context, filters dmarkets.ListFilters) (*dmarkets.MarketDiscoveryPage, error) {
	filters = normalizeDiscoveryListFilters(filters)

	query := r.db.WithContext(ctx).Model(&models.Market{})
	query = applyListStatusFilter(query, filters.Status)
	query = applyDiscoverySearchTerm(query, filters.Query)
	query = applyCreatedByFilter(query, filters.CreatedBy)
	query = applyOwnedByFilter(query, filters.OwnedBy)
	query = applyTagSlugFilter(query, filters.TagSlug)
	query = query.Distinct("markets.*").Order("markets.created_at DESC")

	var dbMarkets []models.Market
	if err := query.Find(&dbMarkets).Error; err != nil {
		return nil, err
	}

	markets := r.mapMarkets(dbMarkets)
	if err := r.hydrateTagsForMarkets(ctx, markets); err != nil {
		return nil, err
	}
	return r.discoveryPageFromMarkets(ctx, markets, filters.Limit, filters.Offset)
}

// ListLifecycleMarketDiscovery returns private/admin lifecycle rows grouped
// before pagination. It intentionally includes proposed/rejected lifecycle
// rows that public discovery excludes.
func (r *GormRepository) ListLifecycleMarketDiscovery(ctx context.Context, filters dmarkets.ListFilters) (*dmarkets.MarketDiscoveryPage, error) {
	filters = normalizeDiscoveryListFilters(filters)

	query := r.db.WithContext(ctx).Model(&models.Market{})
	query = applyLifecycleAdminStatusFilter(query, filters.Status)
	query = applyDiscoverySearchTerm(query, filters.Query)
	query = applyCreatedByFilter(query, filters.CreatedBy)
	query = applyOwnedByFilter(query, filters.OwnedBy)
	query = applyTagSlugFilter(query, filters.TagSlug)
	query = query.Distinct("markets.*").Order("markets.created_at DESC")

	var dbMarkets []models.Market
	if err := query.Find(&dbMarkets).Error; err != nil {
		return nil, err
	}

	markets := r.mapMarkets(dbMarkets)
	if err := r.hydrateTagsForMarkets(ctx, markets); err != nil {
		return nil, err
	}
	if err := r.hydrateStewardshipAudits(ctx, markets); err != nil {
		return nil, err
	}
	return r.discoveryPageFromMarkets(ctx, markets, filters.Limit, filters.Offset)
}

// SearchMarketDiscovery searches public discovery rows by child market text,
// parent group text, and grouped answer labels, then groups before pagination.
func (r *GormRepository) SearchMarketDiscovery(ctx context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.MarketDiscoveryPage, error) {
	filters = normalizeDiscoverySearchFilters(filters)

	dbQuery := r.db.WithContext(ctx).Model(&models.Market{})
	dbQuery = applyDiscoverySearchTerm(dbQuery, query)
	dbQuery = applyStatusFilter(dbQuery, filters.Status)
	dbQuery = applyTagSlugFilter(dbQuery, filters.TagSlug)
	dbQuery = dbQuery.Distinct("markets.*").Order("markets.created_at DESC")

	var dbMarkets []models.Market
	if err := dbQuery.Find(&dbMarkets).Error; err != nil {
		return nil, err
	}

	markets := r.mapMarkets(dbMarkets)
	if err := r.hydrateTagsForMarkets(ctx, markets); err != nil {
		return nil, err
	}
	return r.discoveryPageFromMarkets(ctx, markets, filters.Limit, filters.Offset)
}

func normalizeDiscoveryListFilters(filters dmarkets.ListFilters) dmarkets.ListFilters {
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	return filters
}

func normalizeDiscoverySearchFilters(filters dmarkets.SearchFilters) dmarkets.SearchFilters {
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 50 {
		filters.Limit = 50
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	return filters
}

func applyDiscoverySearchTerm(dbQuery *gorm.DB, value string) *gorm.DB {
	term := strings.TrimSpace(value)
	if term == "" {
		return dbQuery
	}
	searchPattern := "%" + strings.ToLower(term) + "%"
	return dbQuery.
		Joins("LEFT JOIN market_group_members discovery_mgm ON discovery_mgm.market_id = markets.id AND discovery_mgm.deleted_at IS NULL").
		Joins("LEFT JOIN market_groups discovery_mg ON discovery_mg.id = discovery_mgm.group_id AND discovery_mg.deleted_at IS NULL").
		Where(
			`(
				LOWER(markets.question_title) LIKE ?
				OR LOWER(markets.description) LIKE ?
				OR LOWER(discovery_mg.question_title) LIKE ?
				OR LOWER(discovery_mg.description) LIKE ?
				OR LOWER(discovery_mgm.answer_label) LIKE ?
			)`,
			searchPattern,
			searchPattern,
			searchPattern,
			searchPattern,
			searchPattern,
		)
}

type discoveryGroupData struct {
	memberByMarket  map[int64]dmarkets.MarketGroupMember
	groupByID       map[int64]*dmarkets.MarketGroup
	childrenByGroup map[int64][]*dmarkets.Market
}

func (r *GormRepository) discoveryPageFromMarkets(ctx context.Context, markets []*dmarkets.Market, limit int, offset int) (*dmarkets.MarketDiscoveryPage, error) {
	groupData, err := r.loadDiscoveryGroupData(ctx, markets)
	if err != nil {
		return nil, err
	}

	rows := make([]dmarkets.MarketDiscoveryRow, 0, len(markets))
	seenGroups := make(map[int64]bool)
	for _, market := range markets {
		if market == nil {
			continue
		}
		member, grouped := groupData.memberByMarket[market.ID]
		if !grouped {
			rows = append(rows, dmarkets.MarketDiscoveryRow{Market: market})
			continue
		}
		if seenGroups[member.GroupID] {
			continue
		}
		seenGroups[member.GroupID] = true

		group := groupData.groupByID[member.GroupID]
		children := groupData.childrenByGroup[member.GroupID]
		representative := firstDiscoveryChild(children)
		if representative == nil {
			representative = market
		}
		rows = append(rows, dmarkets.MarketDiscoveryRow{
			Market:   representative,
			Group:    group,
			Children: children,
		})
	}

	total := len(rows)
	if offset >= total {
		return &dmarkets.MarketDiscoveryPage{Rows: []dmarkets.MarketDiscoveryRow{}, Total: total}, nil
	}
	end := total
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return &dmarkets.MarketDiscoveryPage{Rows: rows[offset:end], Total: total}, nil
}

func firstDiscoveryChild(children []*dmarkets.Market) *dmarkets.Market {
	for _, child := range children {
		if child != nil {
			return child
		}
	}
	return nil
}

func (r *GormRepository) loadDiscoveryGroupData(ctx context.Context, markets []*dmarkets.Market) (discoveryGroupData, error) {
	data := discoveryGroupData{
		memberByMarket:  map[int64]dmarkets.MarketGroupMember{},
		groupByID:       map[int64]*dmarkets.MarketGroup{},
		childrenByGroup: map[int64][]*dmarkets.Market{},
	}
	if len(markets) == 0 {
		return data, nil
	}

	marketIDs := make([]int64, 0, len(markets))
	for _, market := range markets {
		if market != nil && market.ID > 0 {
			marketIDs = append(marketIDs, market.ID)
		}
	}
	if len(marketIDs) == 0 {
		return data, nil
	}

	var matchedMembers []models.MarketGroupMember
	if err := r.db.WithContext(ctx).
		Where("market_id IN ?", marketIDs).
		Order("group_id ASC, display_order ASC, id ASC").
		Find(&matchedMembers).Error; err != nil {
		return data, err
	}
	if len(matchedMembers) == 0 {
		return data, nil
	}

	groupIDs := make([]int64, 0)
	groupSeen := map[int64]bool{}
	for _, dbMember := range matchedMembers {
		member := modelMarketGroupMemberToDomain(dbMember)
		data.memberByMarket[member.MarketID] = member
		if !groupSeen[member.GroupID] {
			groupSeen[member.GroupID] = true
			groupIDs = append(groupIDs, member.GroupID)
		}
	}

	var dbGroups []models.MarketGroup
	if err := r.db.WithContext(ctx).
		Where("id IN ?", groupIDs).
		Find(&dbGroups).Error; err != nil {
		return data, err
	}
	for _, dbGroup := range dbGroups {
		group := modelMarketGroupToDomain(dbGroup)
		data.groupByID[group.ID] = &group
	}

	var dbAllMembers []models.MarketGroupMember
	if err := r.db.WithContext(ctx).
		Where("group_id IN ?", groupIDs).
		Order("group_id ASC, display_order ASC, id ASC").
		Find(&dbAllMembers).Error; err != nil {
		return data, err
	}

	membersByGroup := make(map[int64][]dmarkets.MarketGroupMember)
	childIDs := make([]int64, 0, len(dbAllMembers))
	for _, dbMember := range dbAllMembers {
		member := modelMarketGroupMemberToDomain(dbMember)
		membersByGroup[member.GroupID] = append(membersByGroup[member.GroupID], member)
		childIDs = append(childIDs, member.MarketID)
	}
	for groupID, members := range membersByGroup {
		if group := data.groupByID[groupID]; group != nil {
			group.Members = dmarkets.OrderedMarketGroupMembers(members)
		}
	}

	var dbChildren []models.Market
	if err := r.db.WithContext(ctx).
		Where("id IN ?", childIDs).
		Find(&dbChildren).Error; err != nil {
		return data, err
	}
	children := r.mapMarkets(dbChildren)
	if err := r.hydrateTagsForMarkets(ctx, children); err != nil {
		return data, err
	}
	if err := r.hydrateStewardshipAudits(ctx, children); err != nil {
		return data, err
	}

	childByID := make(map[int64]*dmarkets.Market, len(children))
	for _, child := range children {
		if child != nil {
			childByID[child.ID] = child
		}
	}
	for groupID, members := range membersByGroup {
		ordered := dmarkets.OrderedMarketGroupMembers(members)
		for _, member := range ordered {
			if child := childByID[member.MarketID]; child != nil {
				data.childrenByGroup[groupID] = append(data.childrenByGroup[groupID], child)
			}
		}
	}

	return data, nil
}
