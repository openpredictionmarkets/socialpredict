package bets

import (
	"context"
	"fmt"
	"math"
	"sort"

	dmarkets "socialpredict/internal/domain/markets"
)

// Sell processes a sell request for credits.
// Sale settlement is accounting-sensitive and remains a synchronous ledger write,
// not a replayable read seam or background execution candidate.
func (s *Service) Sell(ctx context.Context, req SellRequest) (*SellResult, error) {
	outcome, err := s.sellValidator.Validate(ctx, req)
	if err != nil {
		return nil, err
	}
	if outcome == "" {
		return nil, ErrInvalidOutcome
	}

	if s.sellUnit == nil {
		return nil, ErrSellTransactionUnavailable
	}

	return s.sellInTransaction(ctx, req, outcome)
}

// QuoteSell previews a sell request without mutating account or market state.
func (s *Service) QuoteSell(ctx context.Context, req SellRequest) (*SellQuoteResult, error) {
	outcome, err := s.sellValidator.Validate(ctx, req)
	if err != nil {
		return nil, err
	}
	if outcome == "" {
		return nil, ErrInvalidOutcome
	}

	if _, err := (marketGate{markets: s.markets, clock: s.clock}).Open(ctx, int64(req.MarketID)); err != nil {
		return nil, err
	}

	sharesOwned, position, err := s.loadUserShares(ctx, req, outcome)
	if err != nil {
		return nil, err
	}

	sale, err := s.saleCalculator.Quote(position, sharesOwned, req.Amount)
	if err != nil {
		return nil, err
	}

	allowed := true
	if err := validateDustCap(sale.Dust, s.config.MaxDustPerSale); err != nil {
		allowed = false
	}

	suggested := suggestSaleAmounts(sale, sharesOwned, s.config.MaxDustPerSale)
	return new(SellQuoteResult).Build(req, outcome, sale, s.config.MaxDustPerSale, allowed, suggested, s.clock.Now()), nil
}

func (s *Service) sellInTransaction(ctx context.Context, req SellRequest, outcome string) (*SellResult, error) {
	var result *SellResult
	err := s.sellUnit.SellBetTransaction(ctx, func(txCtx context.Context, repo Repository, markets MarketService, users UserService) error {
		if _, err := (marketGate{markets: markets, clock: s.clock}).Open(txCtx, int64(req.MarketID)); err != nil {
			return err
		}

		sharesOwned, position, err := loadUserSharesFrom(txCtx, markets, req, outcome)
		if err != nil {
			return err
		}

		sale, err := s.saleCalculator.Calculate(position, sharesOwned, req.Amount)
		if err != nil {
			return err
		}
		if sale.SharesToSell == 0 {
			return ErrInsufficientShares
		}

		now := s.clock.Now()
		bet := req.NewSaleBet(outcome, sale.SharesToSell, sale.Dust, now)
		if err := (betLedger{repo: repo, users: users}).CreditSale(txCtx, bet, sale.SaleValue); err != nil {
			return err
		}

		result = new(SellResult).Build(req, outcome, sale, now)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) loadUserShares(ctx context.Context, req SellRequest, outcome string) (int64, *dmarkets.UserPosition, error) {
	return loadUserSharesFrom(ctx, s.markets, req, outcome)
}

func loadUserSharesFrom(ctx context.Context, markets PositionReader, req SellRequest, outcome string) (int64, *dmarkets.UserPosition, error) {
	position, err := markets.GetUserPositionInMarket(ctx, int64(req.MarketID), req.Username)
	if err != nil {
		return 0, nil, err
	}
	if position == nil {
		return 0, nil, ErrNoPosition
	}

	sharesOwned, err := sharesOwnedForOutcome(position, outcome)
	if err != nil {
		return 0, nil, err
	}

	return sharesOwned, position, nil
}

// SaleQuote summarises how a sale request would be executed.
// Exported so alternative SaleCalculator implementations can return it.
type SaleQuote struct {
	RequestedCredits int64
	SharesToSell     int64
	SaleValue        int64
	Dust             int64
	ValuePerShare    int64
}

type saleCalculator struct {
	maxDustPerSale int64
}

func (s saleCalculator) Calculate(pos *dmarkets.UserPosition, sharesOwned int64, creditsRequested int64) (SaleQuote, error) {
	quote, err := s.Quote(pos, sharesOwned, creditsRequested)
	if err != nil {
		return SaleQuote{}, err
	}
	if err := validateDustCap(quote.Dust, s.maxDustPerSale); err != nil {
		return SaleQuote{}, err
	}
	return quote, nil
}

func (s saleCalculator) Quote(pos *dmarkets.UserPosition, sharesOwned int64, creditsRequested int64) (SaleQuote, error) {
	if pos == nil {
		return SaleQuote{}, ErrNoPosition
	}
	if sharesOwned <= 0 {
		return SaleQuote{}, ErrNoPosition
	}
	if err := validatePositionValue(pos.Value); err != nil {
		return SaleQuote{}, err
	}

	valuePerShare := pos.Value / sharesOwned
	if valuePerShare <= 0 {
		return SaleQuote{}, ErrNoPosition
	}
	if creditsRequested < valuePerShare {
		return SaleQuote{}, ErrInvalidAmount
	}

	sharesToSell := creditsRequested / valuePerShare
	if sharesToSell > sharesOwned {
		sharesToSell = sharesOwned
	}
	if sharesToSell == 0 {
		return SaleQuote{}, ErrInsufficientShares
	}

	saleValue := sharesToSell * valuePerShare
	dust := calculateDust(creditsRequested, saleValue)

	return SaleQuote{
		RequestedCredits: creditsRequested,
		SharesToSell:     sharesToSell,
		SaleValue:        saleValue,
		Dust:             dust,
		ValuePerShare:    valuePerShare,
	}, nil
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
		return newDustCapExceeded(cap, dust)
	}
	return nil
}

func dustCapCoverage(maxDust int64, valuePerShare int64) float64 {
	if valuePerShare <= 0 {
		return 0
	}
	if maxDust < 0 {
		maxDust = 0
	}
	coverage := float64(maxDust+1) / float64(valuePerShare)
	if coverage > 1 {
		return 1
	}
	return math.Round(coverage*10000) / 10000
}

func sellQuoteMessage(allowed bool, dust int64, maxDust int64) string {
	if allowed {
		if dust == 0 {
			return "This sale can be submitted with no dust."
		}
		return fmt.Sprintf("This sale can be submitted. It would retain %d dust as a whole-share rounding remainder.", dust)
	}
	return fmt.Sprintf("This sale would create %d dust, above the configured maximum of %d. Try a different requested credit amount.", dust, maxDust)
}

func suggestSaleAmounts(sale SaleQuote, sharesOwned int64, maxDust int64) []int64 {
	if sale.ValuePerShare <= 0 || sharesOwned <= 0 {
		return []int64{}
	}
	if maxDust < 0 {
		maxDust = 0
	}

	seen := make(map[int64]struct{})
	var candidates []int64
	for _, shares := range []int64{sale.SharesToSell - 1, sale.SharesToSell, sale.SharesToSell + 1} {
		if shares <= 0 || shares > sharesOwned {
			continue
		}
		base := shares * sale.ValuePerShare
		for dust := int64(0); dust <= maxDust; dust++ {
			amount := base + dust
			if _, ok := seen[amount]; ok {
				continue
			}
			seen[amount] = struct{}{}
			candidates = append(candidates, amount)
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		leftDistance := absInt64(candidates[i] - sale.RequestedCredits)
		rightDistance := absInt64(candidates[j] - sale.RequestedCredits)
		if leftDistance == rightDistance {
			return candidates[i] < candidates[j]
		}
		return leftDistance < rightDistance
	})

	if len(candidates) > 6 {
		candidates = candidates[:6]
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i] < candidates[j] })
	return candidates
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
