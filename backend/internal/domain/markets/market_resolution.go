package markets

import (
	"context"
)

// ResolveMarket resolves a market with a given outcome.
func (s *Service) ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error {
	outcome, err := s.resolutionPolicy.NormalizeResolution(resolution)
	if err != nil {
		return err
	}

	market, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return ErrMarketNotFound
	}
	if market == nil {
		return ErrMarketNotFound
	}

	if err := s.resolutionPolicy.ValidateResolutionRequest(market, username); err != nil {
		return err
	}

	return s.resolutionPolicy.Resolve(ctx, s.repo, s.userService, marketID, outcome)
}
