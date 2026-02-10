package markets

import (
	"context"
	"strings"

	dwallet "socialpredict/internal/domain/wallet"
)

// ResolveMarket resolves a market with a given outcome.
func (s *Service) ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error {
	outcome, err := normalizeResolution(resolution)
	if err != nil {
		return err
	}

	market, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return ErrMarketNotFound
	}

	if err := validateResolutionRequest(market, username); err != nil {
		return err
	}

	if err := s.repo.ResolveMarket(ctx, marketID, outcome); err != nil {
		return err
	}

	if outcome == "N/A" {
		return s.refundMarketBets(ctx, marketID)
	}

	return s.payoutWinningPositions(ctx, marketID)
}

func normalizeResolution(resolution string) (string, error) {
	outcome := strings.ToUpper(strings.TrimSpace(resolution))
	switch outcome {
	case "YES", "NO", "N/A":
		return outcome, nil
	default:
		return "", ErrInvalidInput
	}
}

func validateResolutionRequest(market *Market, username string) error {
	if market.CreatorUsername != username {
		return ErrUnauthorized
	}

	if market.Status == "resolved" {
		return ErrInvalidState
	}

	return nil
}

func (s *Service) refundMarketBets(ctx context.Context, marketID int64) error {
	bets, err := s.repo.ListBetsForMarket(ctx, marketID)
	if err != nil {
		return err
	}
	for _, bet := range bets {
		if err := s.walletService.Credit(ctx, bet.Username, bet.Amount, dwallet.TxRefund); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) payoutWinningPositions(ctx context.Context, marketID int64) error {
	positions, err := s.repo.CalculatePayoutPositions(ctx, marketID)
	if err != nil {
		return err
	}
	for _, pos := range positions {
		if pos.Value <= 0 {
			continue
		}
		if err := s.walletService.Credit(ctx, pos.Username, pos.Value, dwallet.TxWin); err != nil {
			return err
		}
	}
	return nil
}
