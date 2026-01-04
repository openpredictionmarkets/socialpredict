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

	if _, err := s.repo.GetByID(ctx, marketID); err != nil {
		return nil, err
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
	if _, err := s.repo.GetByID(ctx, marketID); err != nil {
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
