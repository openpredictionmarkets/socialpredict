package markets

import (
	"context"
	"sort"
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

// GetMarketPositionsPage returns a display page of market positions. It keeps
// the full canonical position calculation in the repository/service layer and
// only pages the display result.
func (s *Service) GetMarketPositionsPage(ctx context.Context, marketID int64, p Page) (MarketPositions, error) {
	positions, err := s.GetMarketPositions(ctx, marketID)
	if err != nil {
		return nil, err
	}
	positions = activeMarketPositions(positions)
	sortMarketPositionsByTotalShares(positions)
	p = s.statusPolicy.NormalizePage(p, 20, 100)
	return paginateMarketPositions(positions, p), nil
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

func activeMarketPositions(positions MarketPositions) MarketPositions {
	active := make(MarketPositions, 0, len(positions))
	for _, pos := range positions {
		if pos == nil {
			continue
		}
		if pos.YesSharesOwned <= 0 && pos.NoSharesOwned <= 0 {
			continue
		}
		active = append(active, pos)
	}
	return active
}

func sortMarketPositionsByTotalShares(positions MarketPositions) {
	sort.Slice(positions, func(i, j int) bool {
		left := positions[i].YesSharesOwned + positions[i].NoSharesOwned
		right := positions[j].YesSharesOwned + positions[j].NoSharesOwned
		if left == right {
			return positions[i].Username < positions[j].Username
		}
		return left > right
	})
}

func paginateMarketPositions(positions MarketPositions, p Page) MarketPositions {
	if len(positions) == 0 || p.Offset >= len(positions) {
		return MarketPositions{}
	}
	end := p.Offset + p.Limit
	if end > len(positions) {
		end = len(positions)
	}
	return positions[p.Offset:end]
}
