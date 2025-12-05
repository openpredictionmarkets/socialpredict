package analytics

import (
	"context"
	"errors"
	"sort"
	"time"

	marketmath "socialpredict/internal/domain/math/market"
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
	"socialpredict/setup"
)

// Repository exposes the data access required by the analytics domain service.
type Repository interface {
	ListUsers(ctx context.Context) ([]models.User, error)
	ListMarkets(ctx context.Context) ([]models.Market, error)
	ListBetsForMarket(ctx context.Context, marketID uint) ([]models.Bet, error)
	ListBetsOrdered(ctx context.Context) ([]models.Bet, error)
	UserMarketPositions(ctx context.Context, username string) ([]positionsmath.MarketPosition, error)
}

// Service implements analytics calculations.
type Service struct {
	repo       Repository
	econLoader setup.EconConfigLoader
}

// NewService constructs an analytics service.
func NewService(repo Repository, econLoader setup.EconConfigLoader) *Service {
	return &Service{repo: repo, econLoader: econLoader}
}

// ComputeUserFinancials calculates comprehensive financial metrics for a user.
func (s *Service) ComputeUserFinancials(ctx context.Context, req FinancialSnapshotRequest) (*FinancialSnapshot, error) {
	if req.Username == "" {
		return nil, errors.New("username is required")
	}

	if s.econLoader == nil {
		return nil, errors.New("economic configuration loader not provided")
	}

	positions, err := s.repo.UserMarketPositions(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	econConfig := s.econLoader()

	snapshot := &FinancialSnapshot{
		AccountBalance:     req.AccountBalance,
		MaximumDebtAllowed: econConfig.Economics.User.MaximumDebtAllowed,
	}

	for _, pos := range positions {
		profit := pos.Value - pos.TotalSpent

		snapshot.AmountInPlay += pos.Value
		snapshot.TotalSpent += pos.TotalSpent
		snapshot.TradingProfits += profit

		if pos.IsResolved {
			snapshot.RealizedProfits += profit
			snapshot.RealizedValue += pos.Value
		} else {
			snapshot.PotentialProfits += profit
			snapshot.PotentialValue += pos.Value
			snapshot.AmountInPlayActive += pos.Value
			snapshot.TotalSpentInPlay += pos.TotalSpentInPlay
		}
	}

	if req.AccountBalance < 0 {
		snapshot.AmountBorrowed = -req.AccountBalance
	}

	snapshot.RetainedEarnings = req.AccountBalance - snapshot.AmountInPlay
	snapshot.Equity = snapshot.RetainedEarnings + snapshot.AmountInPlay - snapshot.AmountBorrowed

	// Placeholder for future work-based profits integration.	snapshot.WorkProfits = 0
	snapshot.TotalProfits = snapshot.TradingProfits + snapshot.WorkProfits

	return snapshot, nil
}

// ComputeSystemMetrics aggregates system-wide monetary metrics.
func (s *Service) ComputeSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	if s.econLoader == nil {
		return nil, errors.New("economic configuration loader not provided")
	}
	econ := s.econLoader()

	debtStats, err := s.computeDebtStats(ctx, econ)
	if err != nil {
		return nil, err
	}

	volumeStats, err := s.computeMarketVolumes(ctx, econ)
	if err != nil {
		return nil, err
	}

	participationFees, err := s.computeParticipationFees(ctx, econ)
	if err != nil {
		return nil, err
	}

	bonusesPaid := debtStats.realizedProfits
	totalUtilized := debtStats.unusedDebt + volumeStats.activeBetVolume + volumeStats.marketCreationFees + participationFees + bonusesPaid
	surplus := debtStats.totalDebtCapacity - totalUtilized
	balanced := surplus == 0

	metrics := &SystemMetrics{
		MoneyCreated: MoneyCreated{
			UserDebtCapacity: MetricWithExplanation{
				Value:       debtStats.totalDebtCapacity,
				Formula:     "numUsers × maxDebtPerUser",
				Explanation: "Total credit capacity made available to all users",
			},
			NumUsers: MetricWithExplanation{
				Value:       debtStats.userCount,
				Explanation: "Total number of registered users",
			},
		},
		MoneyUtilized: MoneyUtilized{
			UnusedDebt: MetricWithExplanation{
				Value:       debtStats.unusedDebt,
				Formula:     "Σ(maxDebtPerUser - max(0, -balance))",
				Explanation: "Remaining borrowing capacity available to users",
			},
			ActiveBetVolume: MetricWithExplanation{
				Value:       volumeStats.activeBetVolume,
				Formula:     "Σ(unresolved_market_volumes)",
				Explanation: "Total value of bets currently active in unresolved markets (excludes fees and subsidies)",
			},
			MarketCreationFees: MetricWithExplanation{
				Value:       volumeStats.marketCreationFees,
				Formula:     "number_of_markets × creation_fee_per_market",
				Explanation: "Fees collected from users creating new markets",
			},
			ParticipationFees: MetricWithExplanation{
				Value:       participationFees,
				Formula:     "Σ(first_bet_per_user_per_market × participation_fee)",
				Explanation: "Fees collected from first-time participation in each market",
			},
			BonusesPaid: MetricWithExplanation{
				Value:       bonusesPaid,
				Explanation: "System bonuses paid to users and realized profits currently held in user balances",
			},
			TotalUtilized: MetricWithExplanation{
				Value:       totalUtilized,
				Formula:     "unusedDebt + activeBetVolume + marketCreationFees + participationFees + bonusesPaid",
				Explanation: "Total debt capacity that has been utilized across all categories",
			},
		},
		Verification: Verification{
			Balanced: MetricWithExplanation{
				Value:       balanced,
				Explanation: "Whether total created equals total utilized (perfect accounting balance)",
			},
			Surplus: MetricWithExplanation{
				Value:       surplus,
				Formula:     "userDebtCapacity - totalUtilized",
				Explanation: "Positive = unused capacity, Negative = over-utilization (indicates accounting error)",
			},
		},
	}

	return metrics, nil
}

type debtStats struct {
	userCount         int64
	unusedDebt        int64
	realizedProfits   int64
	totalDebtCapacity int64
}

func (s *Service) computeDebtStats(ctx context.Context, econ setup.EconConfig) (*debtStats, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	stats := &debtStats{
		userCount: int64(len(users)),
	}

	for _, user := range users {
		balance := user.AccountBalance
		if balance > 0 {
			stats.realizedProfits += balance
		}
		usedDebt := int64(0)
		if balance < 0 {
			usedDebt = -balance
		}
		stats.unusedDebt += econ.Economics.User.MaximumDebtAllowed - usedDebt
	}

	stats.totalDebtCapacity = econ.Economics.User.MaximumDebtAllowed * stats.userCount
	return stats, nil
}

type marketVolumeStats struct {
	marketCreationFees int64
	activeBetVolume    int64
}

func (s *Service) computeMarketVolumes(ctx context.Context, econ setup.EconConfig) (*marketVolumeStats, error) {
	markets, err := s.repo.ListMarkets(ctx)
	if err != nil {
		return nil, err
	}

	stats := &marketVolumeStats{
		marketCreationFees: int64(len(markets)) * econ.Economics.MarketIncentives.CreateMarketCost,
	}

	for _, market := range markets {
		if market.IsResolved {
			continue
		}

		bets, err := s.repo.ListBetsForMarket(ctx, uint(market.ID))
		if err != nil {
			return nil, err
		}
		stats.activeBetVolume += marketmath.GetMarketVolume(bets)
	}

	return stats, nil
}

func (s *Service) computeParticipationFees(ctx context.Context, econ setup.EconConfig) (int64, error) {
	betsOrdered, err := s.repo.ListBetsOrdered(ctx)
	if err != nil {
		return 0, err
	}

	type userMarket struct {
		marketID uint
		username string
	}

	seen := make(map[userMarket]bool)
	var participationFees int64

	for _, b := range betsOrdered {
		if b.Amount <= 0 {
			continue
		}
		key := userMarket{marketID: b.MarketID, username: b.Username}
		if !seen[key] {
			participationFees += econ.Economics.Betting.BetFees.InitialBetFee
			seen[key] = true
		}
	}

	return participationFees, nil
}

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
