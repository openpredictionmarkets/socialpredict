package markets

import (
	"context"
)

// SearchMarkets searches for markets by query with fallback logic.
func (s *Service) SearchMarkets(ctx context.Context, query string, filters SearchFilters) (*SearchResults, error) {
	if err := s.searchPolicy.ValidateQuery(query); err != nil {
		return nil, err
	}

	filters = s.searchPolicy.NormalizeFilters(filters)

	primaryResults, err := s.repo.Search(ctx, query, filters)
	if err != nil {
		return nil, err
	}

	if primaryResults == nil {
		primaryResults = []*Market{}
	}

	results := s.searchPolicy.NewSearchResults(query, filters.Status, primaryResults)
	if !s.searchPolicy.ShouldFetchFallback(primaryResults, filters.Status) {
		return results, nil
	}

	fallbackFilters := s.searchPolicy.BuildFallbackFilters(filters)

	allResults, err := s.repo.Search(ctx, query, fallbackFilters)
	if err != nil {
		return results, nil
	}

	fallbackResults := s.searchPolicy.SelectFallback(primaryResults, allResults, filters.Limit)
	if len(fallbackResults) == 0 {
		return results, nil
	}

	results.FallbackResults = fallbackResults
	results.FallbackCount = len(fallbackResults)
	results.TotalCount = results.PrimaryCount + results.FallbackCount
	results.FallbackUsed = true

	return results, nil
}
