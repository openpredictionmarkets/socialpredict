package markets

import "context"

// ListActiveMarkets returns markets that are not resolved and active.
func (s *Service) ListActiveMarkets(ctx context.Context, limit int) ([]*Market, error) {
	filters := ListFilters{
		Status: "active",
		Limit:  limit,
	}
	return s.repo.List(ctx, filters)
}

// ListClosedMarkets returns markets that are closed but not resolved.
func (s *Service) ListClosedMarkets(ctx context.Context, limit int) ([]*Market, error) {
	filters := ListFilters{
		Status: "closed",
		Limit:  limit,
	}
	return s.repo.List(ctx, filters)
}

// ListResolvedMarkets returns markets that have been resolved.
func (s *Service) ListResolvedMarkets(ctx context.Context, limit int) ([]*Market, error) {
	filters := ListFilters{
		Status: "resolved",
		Limit:  limit,
	}
	return s.repo.List(ctx, filters)
}

// ListByStatus returns markets filtered by status with pagination.
func (s *Service) ListByStatus(ctx context.Context, status string, p Page) ([]*Market, error) {
	switch status {
	case "active", "closed", "resolved", "all":
	default:
		return nil, ErrInvalidInput
	}

	if p.Limit <= 0 {
		p.Limit = 100
	}
	if p.Limit > 1000 {
		p.Limit = 1000
	}
	if p.Offset < 0 {
		p.Offset = 0
	}

	return s.repo.ListByStatus(ctx, status, p)
}
