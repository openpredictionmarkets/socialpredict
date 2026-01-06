package markets

import (
	"context"
	"strings"

	positionsmath "socialpredict/internal/domain/math/positions"
)

// GetMarketLeaderboard returns the leaderboard for a specific market.
func (s *Service) GetMarketLeaderboard(ctx context.Context, marketID int64, p Page) ([]*LeaderboardRow, error) {
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

	p = s.statusPolicy.NormalizePage(p, 100, 1000)

	bets, err := s.repo.ListBetsForMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}

	if len(bets) == 0 {
		return []*LeaderboardRow{}, nil
	}

	modelBets := convertToModelBets(bets)
	snapshot := marketSnapshotFromModel(market)

	profitability, err := s.leaderboardCalculator.Calculate(snapshot, modelBets)
	if err != nil {
		return nil, err
	}

	if len(profitability) == 0 {
		return []*LeaderboardRow{}, nil
	}

	paged := paginateProfitability(profitability, p)
	return mapLeaderboardRows(paged), nil
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
