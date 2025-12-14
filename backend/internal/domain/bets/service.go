package bets

import (
	"context"
	"strings"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/models"
	"socialpredict/setup"
)

// Repository exposes the persistence layer needed by the bets domain service.
type Repository interface {
	Create(ctx context.Context, bet *models.Bet) error
	UserHasBet(ctx context.Context, marketID uint, username string) (bool, error)
}

type MarketService interface {
	GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error)
	GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error)
}

type UserService interface {
	GetUser(ctx context.Context, username string) (*dusers.User, error)
	ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error
}

// Clock allows time to be mocked in tests.
type Clock interface {
	Now() time.Time
}

type serviceClock struct{}

func (serviceClock) Now() time.Time { return time.Now() }

// ServiceInterface defines the behaviour offered by the bets domain.
type ServiceInterface interface {
	Place(ctx context.Context, req PlaceRequest) (*PlacedBet, error)
	Sell(ctx context.Context, req SellRequest) (*SellResult, error)
}

// Service implements the bets domain logic.
type Service struct {
	repo    Repository
	markets MarketService
	users   UserService
	econ    *setup.EconomicConfig
	clock   Clock
}

// NewService constructs a bets service.
func NewService(repo Repository, markets MarketService, users UserService, econ *setup.EconomicConfig, clock Clock) *Service {
	if clock == nil {
		clock = serviceClock{}
	}
	return &Service{
		repo:    repo,
		markets: markets,
		users:   users,
		econ:    econ,
		clock:   clock,
	}
}

// Place creates a buy bet after validating market status and user balance.
func (s *Service) Place(ctx context.Context, req PlaceRequest) (*PlacedBet, error) {
	outcome, err := validatePlaceRequest(req)
	if err != nil {
		return nil, err
	}

	market, err := s.markets.GetMarket(ctx, int64(req.MarketID))
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()
	if err := ensureMarketOpen(market, now); err != nil {
		return nil, err
	}

	user, err := s.users.GetUser(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	hasBet, err := s.repo.UserHasBet(ctx, req.MarketID, req.Username)
	if err != nil {
		return nil, err
	}

	fees := s.calculateBetFees(hasBet, req.Amount)
	if err := ensureSufficientBalance(user.AccountBalance, fees.totalCost, int64(s.econ.Economics.User.MaximumDebtAllowed)); err != nil {
		return nil, err
	}

	if err := s.users.ApplyTransaction(ctx, req.Username, fees.totalCost, dusers.TransactionBuy); err != nil {
		return nil, err
	}

	bet := &models.Bet{
		Username: req.Username,
		MarketID: req.MarketID,
		Amount:   req.Amount,
		Outcome:  outcome,
		PlacedAt: now,
	}

	if err := s.repo.Create(ctx, bet); err != nil {
		// attempt to roll back user deduction
		_ = s.users.ApplyTransaction(ctx, req.Username, fees.totalCost, dusers.TransactionRefund)
		return nil, err
	}

	return placedBetFromModel(bet), nil
}

// Sell processes a sell request for credits.
func (s *Service) Sell(ctx context.Context, req SellRequest) (*SellResult, error) {
	outcome, err := validateSellRequest(req)
	if err != nil {
		return nil, err
	}

	market, err := s.markets.GetMarket(ctx, int64(req.MarketID))
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()
	if err := ensureMarketOpen(market, now); err != nil {
		return nil, err
	}

	position, err := s.markets.GetUserPositionInMarket(ctx, int64(req.MarketID), req.Username)
	if err != nil {
		return nil, err
	}

	sharesOwned, err := sharesOwnedForOutcome(position, outcome)
	if err != nil {
		return nil, err
	}

	sharesToSell, saleValue, dust, err := s.calculateSale(position, sharesOwned, req.Amount)
	if err != nil {
		return nil, err
	}
	if sharesToSell == 0 {
		return nil, ErrInsufficientShares
	}

	if err := s.users.ApplyTransaction(ctx, req.Username, saleValue, dusers.TransactionSale); err != nil {
		return nil, err
	}

	bet := &models.Bet{
		Username: req.Username,
		MarketID: req.MarketID,
		Amount:   -sharesToSell,
		Outcome:  outcome,
		PlacedAt: now,
	}
	if err := s.repo.Create(ctx, bet); err != nil {
		// Roll back the credit deposited
		_ = s.users.ApplyTransaction(ctx, req.Username, saleValue, dusers.TransactionBuy)
		return nil, err
	}

	return &SellResult{
		Username:      req.Username,
		MarketID:      req.MarketID,
		SharesSold:    sharesToSell,
		SaleValue:     saleValue,
		Dust:          dust,
		Outcome:       outcome,
		TransactionAt: now,
	}, nil
}

func normalizeOutcome(outcome string) string {
	switch strings.ToUpper(strings.TrimSpace(outcome)) {
	case "YES":
		return "YES"
	case "NO":
		return "NO"
	default:
		return ""
	}
}

func sharesOwnedForOutcome(pos *dmarkets.UserPosition, outcome string) (int64, error) {
	switch outcome {
	case "YES":
		if pos.YesSharesOwned == 0 {
			return 0, ErrNoPosition
		}
		return pos.YesSharesOwned, nil
	case "NO":
		if pos.NoSharesOwned == 0 {
			return 0, ErrNoPosition
		}
		return pos.NoSharesOwned, nil
	default:
		return 0, ErrInvalidOutcome
	}
}

func (s *Service) calculateSale(pos *dmarkets.UserPosition, sharesOwned int64, creditsRequested int64) (int64, int64, int64, error) {
	if pos.Value <= 0 {
		return 0, 0, 0, ErrNoPosition
	}
	valuePerShare := pos.Value / sharesOwned
	if valuePerShare <= 0 {
		return 0, 0, 0, ErrNoPosition
	}
	if creditsRequested < valuePerShare {
		return 0, 0, 0, ErrInvalidAmount
	}

	sharesToSell := creditsRequested / valuePerShare
	if sharesToSell > sharesOwned {
		sharesToSell = sharesOwned
	}
	if sharesToSell == 0 {
		return 0, 0, 0, ErrInsufficientShares
	}

	saleValue := sharesToSell * valuePerShare
	dust := creditsRequested - saleValue
	if dust < 0 {
		dust = 0
	}

	cap := s.econ.Economics.Betting.MaxDustPerSale
	if cap > 0 && dust > cap {
		return 0, 0, 0, ErrDustCapExceeded{Cap: cap, Requested: dust}
	}

	return sharesToSell, saleValue, dust, nil
}

func validatePlaceRequest(req PlaceRequest) (string, error) {
	outcome := normalizeOutcome(req.Outcome)
	if outcome == "" {
		return "", ErrInvalidOutcome
	}
	if req.Amount <= 0 {
		return "", ErrInvalidAmount
	}
	return outcome, nil
}

func validateSellRequest(req SellRequest) (string, error) {
	outcome := normalizeOutcome(req.Outcome)
	if outcome == "" {
		return "", ErrInvalidOutcome
	}
	if req.Amount <= 0 {
		return "", ErrInvalidAmount
	}
	return outcome, nil
}

func ensureMarketOpen(market *dmarkets.Market, now time.Time) error {
	if market.Status == "resolved" || now.After(market.ResolutionDateTime) {
		return ErrMarketClosed
	}
	return nil
}

type betFees struct {
	initialFee     int64
	transactionFee int64
	totalCost      int64
}

func (s *Service) calculateBetFees(hasBet bool, amount int64) betFees {
	fees := betFees{
		initialFee:     0,
		transactionFee: int64(s.econ.Economics.Betting.BetFees.BuySharesFee),
	}
	if !hasBet {
		fees.initialFee = int64(s.econ.Economics.Betting.BetFees.InitialBetFee)
	}
	fees.totalCost = amount + fees.initialFee + fees.transactionFee
	return fees
}

func ensureSufficientBalance(balance, totalCost, maxDebt int64) error {
	if balance-totalCost < -maxDebt {
		return ErrInsufficientBalance
	}
	return nil
}

func placedBetFromModel(bet *models.Bet) *PlacedBet {
	return &PlacedBet{
		Username: bet.Username,
		MarketID: bet.MarketID,
		Amount:   bet.Amount,
		Outcome:  bet.Outcome,
		PlacedAt: bet.PlacedAt,
	}
}
