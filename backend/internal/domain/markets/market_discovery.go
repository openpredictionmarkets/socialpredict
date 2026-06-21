package markets

import (
	"context"
	"strconv"
)

// MarketDiscoveryRow is a display/read-model row. Standalone rows contain one
// Market; grouped rows contain the parent Group and every child answer market.
type MarketDiscoveryRow struct {
	Market   *Market
	Group    *MarketGroup
	Children []*Market
}

// MarketDiscoveryPage is paginated after grouped child markets have been
// collapsed into one discovery row.
type MarketDiscoveryPage struct {
	Rows  []MarketDiscoveryRow
	Total int
}

// MarketDiscoverySearchResults mirrors SearchResults with grouped rows.
type MarketDiscoverySearchResults struct {
	PrimaryRows   []MarketDiscoveryRow
	FallbackRows  []MarketDiscoveryRow
	Query         string
	PrimaryStatus string
	PrimaryCount  int
	FallbackCount int
	TotalCount    int
	FallbackUsed  bool
}

// MarketDiscoveryRepository exposes display discovery rows grouped before
// pagination. Transaction paths must not use this interface.
type MarketDiscoveryRepository interface {
	ListMarketDiscovery(ctx context.Context, filters ListFilters) (*MarketDiscoveryPage, error)
	ListLifecycleMarketDiscovery(ctx context.Context, filters ListFilters) (*MarketDiscoveryPage, error)
	SearchMarketDiscovery(ctx context.Context, query string, filters SearchFilters) (*MarketDiscoveryPage, error)
}

// ListMarketDiscovery returns market discovery rows grouped before pagination.
func (s *Service) ListMarketDiscovery(ctx context.Context, filters ListFilters) (*MarketDiscoveryPage, error) {
	if filters.Status != "" {
		if err := s.statusPolicy.ValidateStatus(filters.Status); err != nil {
			return nil, err
		}
	}
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	if repo, ok := s.repo.(MarketDiscoveryRepository); ok {
		return repo.ListMarketDiscovery(ctx, filters)
	}

	markets, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, err
	}
	rows := make([]MarketDiscoveryRow, 0, len(markets))
	for _, market := range markets {
		rows = append(rows, MarketDiscoveryRow{Market: market})
	}
	return &MarketDiscoveryPage{Rows: rows, Total: len(rows)}, nil
}

// ListLifecycleMarketDiscovery returns private/admin lifecycle rows grouped
// before pagination. It is for display queues only, not transaction paths.
func (s *Service) ListLifecycleMarketDiscovery(ctx context.Context, filters ListFilters) (*MarketDiscoveryPage, error) {
	status := NormalizeLifecycleStatus(filters.Status)
	switch status {
	case MarketStatusAll, MarketLifecycleProposed, MarketLifecyclePublished, MarketLifecycleRejected, MarketLifecycleClosed, MarketLifecycleResolved, MarketLifecycleCancelled:
	default:
		return nil, ErrInvalidInput
	}
	filters.Status = status
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	if repo, ok := s.repo.(MarketDiscoveryRepository); ok {
		return repo.ListLifecycleMarketDiscovery(ctx, filters)
	}

	lifecycleRepo, ok := s.repo.(LifecycleReadRepository)
	if !ok {
		return nil, ErrInvalidInput
	}
	markets, err := lifecycleRepo.ListByLifecycle(ctx, filters)
	if err != nil {
		return nil, err
	}
	rows := make([]MarketDiscoveryRow, 0, len(markets))
	for _, market := range markets {
		rows = append(rows, MarketDiscoveryRow{Market: market})
	}
	return &MarketDiscoveryPage{Rows: rows, Total: len(rows)}, nil
}

// SearchMarketDiscovery searches display discovery rows grouped before
// pagination. It preserves the existing primary/fallback search behavior.
func (s *Service) SearchMarketDiscovery(ctx context.Context, query string, filters SearchFilters) (*MarketDiscoverySearchResults, error) {
	if err := s.searchPolicy.ValidateQuery(query); err != nil {
		return nil, err
	}
	filters = s.searchPolicy.NormalizeFilters(filters)

	repo, ok := s.repo.(MarketDiscoveryRepository)
	if !ok {
		legacy, err := s.SearchMarkets(ctx, query, filters)
		if err != nil {
			return nil, err
		}
		return marketDiscoverySearchResultsFromLegacy(legacy), nil
	}

	primary, err := repo.SearchMarketDiscovery(ctx, query, filters)
	if err != nil {
		return nil, err
	}
	if primary == nil {
		primary = &MarketDiscoveryPage{}
	}
	results := &MarketDiscoverySearchResults{
		PrimaryRows:   primary.Rows,
		Query:         query,
		PrimaryStatus: filters.Status,
		PrimaryCount:  len(primary.Rows),
		TotalCount:    len(primary.Rows),
	}
	if !s.searchPolicy.ShouldFetchFallback(marketsFromDiscoveryRows(primary.Rows), filters.Status) {
		return results, nil
	}

	fallbackFilters := s.searchPolicy.BuildFallbackFilters(filters)
	all, err := repo.SearchMarketDiscovery(ctx, query, fallbackFilters)
	if err != nil || all == nil || len(all.Rows) == 0 {
		return results, nil
	}

	primaryKeys := map[string]bool{}
	for _, row := range primary.Rows {
		primaryKeys[marketDiscoveryRowKey(row)] = true
	}
	for _, row := range all.Rows {
		if primaryKeys[marketDiscoveryRowKey(row)] {
			continue
		}
		results.FallbackRows = append(results.FallbackRows, row)
		if len(results.FallbackRows) >= filters.Limit {
			break
		}
	}
	results.FallbackCount = len(results.FallbackRows)
	results.TotalCount = results.PrimaryCount + results.FallbackCount
	results.FallbackUsed = results.FallbackCount > 0
	return results, nil
}

func marketDiscoverySearchResultsFromLegacy(legacy *SearchResults) *MarketDiscoverySearchResults {
	if legacy == nil {
		return &MarketDiscoverySearchResults{}
	}
	return &MarketDiscoverySearchResults{
		PrimaryRows:   discoveryRowsFromMarkets(legacy.PrimaryResults),
		FallbackRows:  discoveryRowsFromMarkets(legacy.FallbackResults),
		Query:         legacy.Query,
		PrimaryStatus: legacy.PrimaryStatus,
		PrimaryCount:  legacy.PrimaryCount,
		FallbackCount: legacy.FallbackCount,
		TotalCount:    legacy.TotalCount,
		FallbackUsed:  legacy.FallbackUsed,
	}
}

func discoveryRowsFromMarkets(markets []*Market) []MarketDiscoveryRow {
	rows := make([]MarketDiscoveryRow, 0, len(markets))
	for _, market := range markets {
		rows = append(rows, MarketDiscoveryRow{Market: market})
	}
	return rows
}

func marketsFromDiscoveryRows(rows []MarketDiscoveryRow) []*Market {
	markets := make([]*Market, 0, len(rows))
	for _, row := range rows {
		if row.Market != nil {
			markets = append(markets, row.Market)
		}
	}
	return markets
}

func marketDiscoveryRowKey(row MarketDiscoveryRow) string {
	if row.Group != nil && row.Group.ID > 0 {
		return "group:" + strconv.FormatInt(row.Group.ID, 10)
	}
	if row.Market != nil {
		return "market:" + strconv.FormatInt(row.Market.ID, 10)
	}
	return "empty"
}
