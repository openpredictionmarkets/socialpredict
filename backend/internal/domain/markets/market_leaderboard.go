package markets

import (
	"context"
	"strings"

	positionsmath "socialpredict/internal/domain/math/positions"
)

// GetMarketLeaderboard returns the leaderboard for a specific market.
func (s *Service) GetMarketLeaderboard(ctx context.Context, marketID int64, p Page) ([]*LeaderboardRow, error) {
	rows, err := s.getMarketLeaderboardRows(ctx, marketID)
	if err != nil {
		return nil, err
	}

	p = s.statusPolicy.NormalizePage(p, 100, 1000)
	return paginateLeaderboardRows(rows, p), nil
}

func (s *Service) getMarketLeaderboardRows(ctx context.Context, marketID int64) ([]*LeaderboardRow, error) {
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

	bets, err := s.repo.ListBetsForMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}

	if len(bets) == 0 {
		return []*LeaderboardRow{}, nil
	}

	boundaryBets := convertToBoundaryBets(bets)
	snapshot := marketSnapshotFromModel(market)

	profitability, err := s.leaderboardCalculator.Calculate(snapshot, boundaryBets)
	if err != nil {
		return nil, err
	}

	if len(profitability) == 0 {
		return []*LeaderboardRow{}, nil
	}

	return mapLeaderboardRows(profitability), nil
}

func marketSnapshotFromModel(market *Market) positionsmath.MarketSnapshot {
	return positionsmath.MarketSnapshot{
		ID:               market.ID,
		CreatedAt:        market.CreatedAt,
		IsResolved:       strings.EqualFold(market.Status, "resolved"),
		ResolutionResult: market.ResolutionResult,
	}
}

func paginateProfitability(profitability []positionsmath.UserProfitability, p Page) []positionsmath.UserProfitability {
	start := p.Offset
	if start > len(profitability) {
		start = len(profitability)
	}
	end := start + p.Limit
	if end > len(profitability) {
		end = len(profitability)
	}
	return profitability[start:end]
}

func mapLeaderboardRows(rows []positionsmath.UserProfitability) []*LeaderboardRow {
	if len(rows) == 0 {
		return []*LeaderboardRow{}
	}
	leaderboard := make([]*LeaderboardRow, len(rows))
	for i, row := range rows {
		leaderboard[i] = &LeaderboardRow{
			Username:       row.Username,
			Profit:         row.Profit,
			CurrentValue:   row.CurrentValue,
			TotalSpent:     row.TotalSpent,
			Position:       row.Position,
			YesSharesOwned: row.YesSharesOwned,
			NoSharesOwned:  row.NoSharesOwned,
			Rank:           row.Rank,
		}
	}
	return leaderboard
}
