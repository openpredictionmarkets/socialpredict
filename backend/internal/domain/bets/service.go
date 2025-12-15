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

	marketGate     marketGate
	fees           feeCalculator
	balances       balanceGuard
	ledger         betLedger
	saleCalculator saleCalculator
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

		marketGate:     marketGate{markets: markets, clock: clock},
		fees:           feeCalculator{econ: econ},
		balances:       balanceGuard{maxDebtAllowed: int64(econ.Economics.User.MaximumDebtAllowed)},
		ledger:         betLedger{repo: repo, users: users},
		saleCalculator: saleCalculator{maxDustPerSale: int64(econ.Economics.Betting.MaxDustPerSale)},
	}
}

// Place creates a buy bet after validating market status and user balance.
func (s *Service) Place(ctx context.Context, req PlaceRequest) (*PlacedBet, error) {
	outcome, err := validatePlaceRequest(req)
	if err != nil {
		return nil, err
	}

	if _, err := s.marketGate.Open(ctx, int64(req.MarketID)); err != nil {
		return nil, err
	}

	user, hasBet, err := s.loadUserAndBetStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	fees := s.fees.Calculate(hasBet, req.Amount)
	if err := s.balances.EnsureSufficient(user.AccountBalance, fees.totalCost); err != nil {
		return nil, err
	}

	now := s.clock.Now()
	bet := &models.Bet{
		Username: req.Username,
		MarketID: req.MarketID,
		Amount:   req.Amount,
		Outcome:  outcome,
		PlacedAt: now,
	}

	if err := s.ledger.ChargeAndRecord(ctx, bet, fees.totalCost); err != nil {
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

	if _, err := s.marketGate.Open(ctx, int64(req.MarketID)); err != nil {
		return nil, err
	}

	sharesOwned, position, err := s.loadUserShares(ctx, req, outcome)
	if err != nil {
		return nil, err
	}

	sale, err := s.saleCalculator.Calculate(position, sharesOwned, req.Amount)
	if err != nil {
		return nil, err
	}
	if sale.sharesToSell == 0 {
		return nil, ErrInsufficientShares
	}

	now := s.clock.Now()
	bet := &models.Bet{
		Username: req.Username,
		MarketID: req.MarketID,
		Amount:   -sale.sharesToSell,
		Outcome:  outcome,
		PlacedAt: now,
	}
	if err := s.ledger.CreditSale(ctx, bet, sale.saleValue); err != nil {
		return nil, err
	}

	return &SellResult{
		Username:      req.Username,
		MarketID:      req.MarketID,
		SharesSold:    sale.sharesToSell,
		SaleValue:     sale.saleValue,
		Dust:          sale.dust,
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

type saleResult struct {
	sharesToSell int64
	saleValue    int64
	dust         int64
}

func (s *Service) loadUserShares(ctx context.Context, req SellRequest, outcome string) (int64, *dmarkets.UserPosition, error) {
	position, err := s.markets.GetUserPositionInMarket(ctx, int64(req.MarketID), req.Username)
	if err != nil {
		return 0, nil, err
	}

	sharesOwned, err := sharesOwnedForOutcome(position, outcome)
	if err != nil {
		return 0, nil, err
	}

	return sharesOwned, position, nil
}

func validatePositionValue(value int64) error {
	if value <= 0 {
		return ErrNoPosition
	}
	return nil
}

func calculateDust(requested, saleValue int64) int64 {
	dust := requested - saleValue
	if dust < 0 {
		return 0
	}
	return dust
}

func validateDustCap(dust int64, cap int64) error {
	if cap > 0 && dust > cap {
		return ErrDustCapExceeded{Cap: cap, Requested: dust}
	}
	return nil
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

func placedBetFromModel(bet *models.Bet) *PlacedBet {
	return &PlacedBet{
		Username: bet.Username,
		MarketID: bet.MarketID,
		Amount:   bet.Amount,
		Outcome:  bet.Outcome,
		PlacedAt: bet.PlacedAt,
	}
}

func (s *Service) loadUserAndBetStatus(ctx context.Context, req PlaceRequest) (*dusers.User, bool, error) {
	user, err := s.users.GetUser(ctx, req.Username)
	if err != nil {
		return nil, false, err
	}

	hasBet, err := s.repo.UserHasBet(ctx, req.MarketID, req.Username)
	if err != nil {
		return nil, false, err
	}

	return user, hasBet, nil
}

// marketGate ensures markets are open before interacting with them.
type marketGate struct {
	markets MarketService
	clock   Clock
}

func (g marketGate) Open(ctx context.Context, marketID int64) (*dmarkets.Market, error) {
	market, err := g.markets.GetMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if err := ensureMarketOpen(market, g.clock.Now()); err != nil {
		return nil, err
	}
	return market, nil
}

type feeCalculator struct {
	econ *setup.EconomicConfig
}

func (f feeCalculator) Calculate(hasBet bool, amount int64) betFees {
	fees := betFees{
		initialFee:     0,
		transactionFee: int64(f.econ.Economics.Betting.BetFees.BuySharesFee),
	}
	if !hasBet {
		fees.initialFee = int64(f.econ.Economics.Betting.BetFees.InitialBetFee)
	}
	fees.totalCost = amount + fees.initialFee + fees.transactionFee
	return fees
}

type balanceGuard struct {
	maxDebtAllowed int64
}

func (g balanceGuard) EnsureSufficient(balance, totalCost int64) error {
	if balance-totalCost < -g.maxDebtAllowed {
		return ErrInsufficientBalance
	}
	return nil
}

type betLedger struct {
	repo  Repository
	users UserService
}

func (l betLedger) ChargeAndRecord(ctx context.Context, bet *models.Bet, totalCost int64) error {
	if err := l.users.ApplyTransaction(ctx, bet.Username, totalCost, dusers.TransactionBuy); err != nil {
		return err
	}

	if err := l.repo.Create(ctx, bet); err != nil {
		_ = l.users.ApplyTransaction(ctx, bet.Username, totalCost, dusers.TransactionRefund)
		return err
	}
	return nil
}

func (l betLedger) CreditSale(ctx context.Context, bet *models.Bet, saleValue int64) error {
	if err := l.users.ApplyTransaction(ctx, bet.Username, saleValue, dusers.TransactionSale); err != nil {
		return err
	}
	if err := l.repo.Create(ctx, bet); err != nil {
		// Roll back the credit deposited
		_ = l.users.ApplyTransaction(ctx, bet.Username, saleValue, dusers.TransactionBuy)
		return err
	}
	return nil
}

type saleCalculator struct {
	maxDustPerSale int64
}

func (s saleCalculator) Calculate(pos *dmarkets.UserPosition, sharesOwned int64, creditsRequested int64) (saleResult, error) {
	if err := validatePositionValue(pos.Value); err != nil {
		return saleResult{}, err
	}

	valuePerShare := pos.Value / sharesOwned
	if valuePerShare <= 0 {
		return saleResult{}, ErrNoPosition
	}
	if creditsRequested < valuePerShare {
		return saleResult{}, ErrInvalidAmount
	}

	sharesToSell := creditsRequested / valuePerShare
	if sharesToSell > sharesOwned {
		sharesToSell = sharesOwned
	}
	if sharesToSell == 0 {
		return saleResult{}, ErrInsufficientShares
	}

	saleValue := sharesToSell * valuePerShare
	dust := calculateDust(creditsRequested, saleValue)

	if err := validateDustCap(dust, s.maxDustPerSale); err != nil {
		return saleResult{}, err
	}

	return saleResult{sharesToSell: sharesToSell, saleValue: saleValue, dust: dust}, nil
}
