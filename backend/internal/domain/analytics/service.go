package analytics

import (
	"context"
	"errors"

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

	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	var (
		userCount       = int64(len(users))
		unusedDebt      int64
		realizedProfits int64
	)

	for _, user := range users {
		balance := user.AccountBalance
		if balance > 0 {
			realizedProfits += balance
		}
		usedDebt := int64(0)
		if balance < 0 {
			usedDebt = -balance
		}
		unusedDebt += econ.Economics.User.MaximumDebtAllowed - usedDebt
	}

	totalDebtCapacity := econ.Economics.User.MaximumDebtAllowed * userCount

	markets, err := s.repo.ListMarkets(ctx)
	if err != nil {
		return nil, err
	}

	marketCreationFees := int64(len(markets)) * econ.Economics.MarketIncentives.CreateMarketCost

	var activeBetVolume int64
	for _, market := range markets {
		if market.IsResolved {
			continue
		}

		bets, err := s.repo.ListBetsForMarket(ctx, uint(market.ID))
		if err != nil {
			return nil, err
		}
		activeBetVolume += marketmath.GetMarketVolume(bets)
	}

	betsOrdered, err := s.repo.ListBetsOrdered(ctx)
	if err != nil {
		return nil, err
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

	bonusesPaid := realizedProfits
	totalUtilized := unusedDebt + activeBetVolume + marketCreationFees + participationFees + bonusesPaid
	surplus := totalDebtCapacity - totalUtilized
	balanced := surplus == 0

	metrics := &SystemMetrics{
		MoneyCreated: MoneyCreated{
			UserDebtCapacity: MetricWithExplanation{
				Value:       totalDebtCapacity,
				Formula:     "numUsers × maxDebtPerUser",
				Explanation: "Total credit capacity made available to all users",
			},
			NumUsers: MetricWithExplanation{
				Value:       userCount,
				Explanation: "Total number of registered users",
			},
		},
		MoneyUtilized: MoneyUtilized{
			UnusedDebt: MetricWithExplanation{
				Value:       unusedDebt,
				Formula:     "Σ(maxDebtPerUser - max(0, -balance))",
				Explanation: "Remaining borrowing capacity available to users",
			},
			ActiveBetVolume: MetricWithExplanation{
				Value:       activeBetVolume,
				Formula:     "Σ(unresolved_market_volumes)",
				Explanation: "Total value of bets currently active in unresolved markets (excludes fees and subsidies)",
			},
			MarketCreationFees: MetricWithExplanation{
				Value:       marketCreationFees,
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
