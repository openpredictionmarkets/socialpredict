package markets

import (
	"context"
	"strings"
)

// SearchMarkets searches for markets by query with fallback logic.
func (s *Service) SearchMarkets(ctx context.Context, query string, filters SearchFilters) (*SearchResults, error) {
	if err := validateSearchQuery(query); err != nil {
		return nil, err
	}

	filters = normalizeSearchFilters(filters)

	primaryResults, err := s.repo.Search(ctx, query, filters)
	if err != nil {
		return nil, err
	}

	results := newSearchResults(query, filters.Status, primaryResults)
	if !shouldFetchFallback(primaryResults, filters.Status) {
		return results, nil
	}

	fallbackResults, fallbackErr := s.fetchFallbackMarkets(ctx, query, filters, primaryResults)
	if fallbackErr != nil || len(fallbackResults) == 0 {
		return results, nil
	}

	results.FallbackResults = fallbackResults
	results.FallbackCount = len(fallbackResults)
	results.TotalCount = results.PrimaryCount + results.FallbackCount
	results.FallbackUsed = true

	return results, nil
}

func validateSearchQuery(query string) error {
	if strings.TrimSpace(query) == "" {
		return ErrInvalidInput
	}
	return nil
}

func normalizeSearchFilters(filters SearchFilters) SearchFilters {
	if filters.Limit <= 0 || filters.Limit > 50 {
		filters.Limit = 20
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	return filters
}

func newSearchResults(query string, status string, primaryResults []*Market) *SearchResults {
	return &SearchResults{
		PrimaryResults:  primaryResults,
		FallbackResults: []*Market{},
		Query:           query,
		PrimaryStatus:   status,
		PrimaryCount:    len(primaryResults),
		FallbackCount:   0,
		TotalCount:      len(primaryResults),
		FallbackUsed:    false,
	}
}

func shouldFetchFallback(primaryResults []*Market, status string) bool {
	return len(primaryResults) <= 5 && status != "" && status != "all"
}

func (s *Service) fetchFallbackMarkets(ctx context.Context, query string, filters SearchFilters, primaryResults []*Market) ([]*Market, error) {
	allFilters := SearchFilters{
		Status: "",
		Limit:  filters.Limit * 2,
		Offset: 0,
	}

	allResults, err := s.repo.Search(ctx, query, allFilters)
	if err != nil {
		return nil, err
	}

	primaryIDs := make(map[int64]bool)
	for _, market := range primaryResults {
		primaryIDs[market.ID] = true
	}

	var fallbackResults []*Market
	for _, market := range allResults {
		if primaryIDs[market.ID] {
			continue
		}
		fallbackResults = append(fallbackResults, market)
		if len(fallbackResults) >= filters.Limit {
			break
		}
	}

	return fallbackResults, nil
}
