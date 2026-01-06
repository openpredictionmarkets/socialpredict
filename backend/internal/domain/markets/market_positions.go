package markets

import (
	"context"
	"strings"
)

// GetMarketPositions returns all user positions in a market.
func (s *Service) GetMarketPositions(ctx context.Context, marketID int64) (MarketPositions, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}

	market, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if market == nil {
		return nil, ErrMarketNotFound
	}

	positions, err := s.repo.ListMarketPositions(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if positions == nil {
		return MarketPositions{}, nil
	}
	return positions, nil
}

// GetUserPositionInMarket returns a specific user's position in a market.
func (s *Service) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*UserPosition, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}

	market, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return nil, ErrMarketNotFound
	}
	if market == nil {
		return nil, ErrMarketNotFound
	}

	if strings.TrimSpace(username) == "" {
		return nil, ErrInvalidInput
	}

	position, err := s.repo.GetUserPosition(ctx, marketID, username)
	if err != nil {
		return nil, err
	}
	return position, nil
}
