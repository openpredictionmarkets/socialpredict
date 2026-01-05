package analytics

import (
	"context"
	"sort"
	"time"

	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
)

// GlobalUserProfitability summarises a user's profitability across all markets.
type GlobalUserProfitability struct {
	Username          string    `json:"username"`
	TotalProfit       int64     `json:"totalProfit"`
	TotalCurrentValue int64     `json:"totalCurrentValue"`
	TotalSpent        int64     `json:"totalSpent"`
	ActiveMarkets     int       `json:"activeMarkets"`
	ResolvedMarkets   int       `json:"resolvedMarkets"`
	EarliestBet       time.Time `json:"earliestBet"`
	Rank              int       `json:"rank"`
}

// ComputeGlobalLeaderboard ranks users by profitability across all markets.
func (s *Service) ComputeGlobalLeaderboard(ctx context.Context) ([]GlobalUserProfitability, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return []GlobalUserProfitability{}, nil
	}

	markets, err := s.repo.ListMarkets(ctx)
	if err != nil {
		return nil, err
	}
	if len(markets) == 0 {
		return []GlobalUserProfitability{}, nil
	}

	marketData, err := s.loadLeaderboardMarketData(ctx, markets)
	if err != nil {
		return nil, err
	}
	if len(marketData) == 0 {
		return []GlobalUserProfitability{}, nil
	}

	aggregates := aggregateLeaderboardUserStats(marketData)
	if len(aggregates) == 0 {
		return []GlobalUserProfitability{}, nil
	}

	earliestBets := findEarliestBetsPerUser(marketData, aggregates)
	leaderboard := assembleLeaderboardEntries(aggregates, earliestBets)
	return rankLeaderboardEntries(leaderboard), nil
}

type leaderboardMarketData struct {
	snapshot  positionsmath.MarketSnapshot
	positions []positionsmath.MarketPosition
	bets      []models.Bet
}

type leaderboardAggregate struct {
	totalProfit       int64
	totalCurrentValue int64
	totalSpent        int64
	activeMarkets     int
	resolvedMarkets   int
}

func (s *Service) loadLeaderboardMarketData(ctx context.Context, markets []models.Market) ([]leaderboardMarketData, error) {
	data := make([]leaderboardMarketData, 0, len(markets))

	for _, market := range markets {
		bets, err := s.repo.ListBetsForMarket(ctx, uint(market.ID))
		if err != nil {
			return nil, err
		}

		snapshot := positionsmath.MarketSnapshot{
			ID:               int64(market.ID),
			CreatedAt:        market.CreatedAt,
			IsResolved:       market.IsResolved,
			ResolutionResult: market.ResolutionResult,
		}

		marketPositions, err := positionsmath.CalculateMarketPositions_WPAM_DBPM(snapshot, bets)
		if err != nil {
			return nil, err
		}

		data = append(data, leaderboardMarketData{
			snapshot:  snapshot,
			positions: marketPositions,
			bets:      bets,
		})
	}

	return data, nil
}

func aggregateLeaderboardUserStats(markets []leaderboardMarketData) map[string]*leaderboardAggregate {
	aggregates := make(map[string]*leaderboardAggregate)

	for _, market := range markets {
		for _, pos := range market.positions {
			agg := aggregates[pos.Username]
			if agg == nil {
				agg = &leaderboardAggregate{}
				aggregates[pos.Username] = agg
			}

			profit := pos.Value - pos.TotalSpent
			agg.totalProfit += profit
			agg.totalCurrentValue += pos.Value
			agg.totalSpent += pos.TotalSpent
			if pos.IsResolved {
				agg.resolvedMarkets++
			} else {
				agg.activeMarkets++
			}
		}
	}

	return aggregates
}

func findEarliestBetsPerUser(markets []leaderboardMarketData, aggregates map[string]*leaderboardAggregate) map[string]time.Time {
	earliest := make(map[string]time.Time)

	for _, market := range markets {
		for _, bet := range market.bets {
			if aggregates[bet.Username] == nil {
				continue
			}
			if ts, ok := earliest[bet.Username]; !ok || bet.PlacedAt.Before(ts) {
				earliest[bet.Username] = bet.PlacedAt
			}
		}
	}

	return earliest
}

func assembleLeaderboardEntries(aggregates map[string]*leaderboardAggregate, earliest map[string]time.Time) []GlobalUserProfitability {
	leaderboard := make([]GlobalUserProfitability, 0, len(aggregates))

	for username, agg := range aggregates {
		firstBet, ok := earliest[username]
		if !ok {
			continue
		}
		leaderboard = append(leaderboard, GlobalUserProfitability{
			Username:          username,
			TotalProfit:       agg.totalProfit,
			TotalCurrentValue: agg.totalCurrentValue,
			TotalSpent:        agg.totalSpent,
			ActiveMarkets:     agg.activeMarkets,
			ResolvedMarkets:   agg.resolvedMarkets,
			EarliestBet:       firstBet,
		})
	}

	return leaderboard
}

func rankLeaderboardEntries(entries []GlobalUserProfitability) []GlobalUserProfitability {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].TotalProfit == entries[j].TotalProfit {
			return entries[i].EarliestBet.Before(entries[j].EarliestBet)
		}
		return entries[i].TotalProfit > entries[j].TotalProfit
	})

	for i := range entries {
		entries[i].Rank = i + 1
	}

	return entries
}
