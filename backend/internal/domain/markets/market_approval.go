package markets

import (
	"context"
	"strings"
	"time"

	users "socialpredict/internal/domain/users"
)

// MarketApprovalRepository persists admin review decisions for proposed markets.
type MarketApprovalRepository interface {
	ApproveMarket(ctx context.Context, id int64, actorUsername string, approvedAt time.Time) error
	RejectMarket(ctx context.Context, id int64, actorUsername string, rejectedAt time.Time, reason string) error
}

func (s *Service) ApproveProposedMarket(ctx context.Context, marketID int64, actorUsername string, confirmed bool) (*Market, error) {
	if marketID <= 0 || strings.TrimSpace(actorUsername) == "" || !confirmed {
		return nil, ErrInvalidInput
	}

	market, err := s.GetMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if NormalizeLifecycleStatus(market.LifecycleStatus) != MarketLifecycleProposed {
		return nil, ErrInvalidState
	}

	now := s.clock.Now()
	if err := market.Publish(now); err != nil {
		return nil, err
	}
	market.ApprovedBy = actorUsername
	market.ApprovedAt = &now

	repo, err := s.approvalRepository()
	if err != nil {
		return nil, err
	}
	if err := repo.ApproveMarket(ctx, marketID, actorUsername, now); err != nil {
		return nil, err
	}
	return market, nil
}

func (s *Service) RejectProposedMarket(ctx context.Context, marketID int64, actorUsername string, reason string) (*Market, error) {
	if marketID <= 0 || strings.TrimSpace(actorUsername) == "" || strings.TrimSpace(reason) == "" {
		return nil, ErrInvalidInput
	}

	market, err := s.GetMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if NormalizeLifecycleStatus(market.LifecycleStatus) != MarketLifecycleProposed {
		return nil, ErrInvalidState
	}

	now := s.clock.Now()
	if err := market.Reject(now); err != nil {
		return nil, err
	}
	market.RejectedBy = actorUsername
	market.RejectedAt = &now
	market.RejectionReason = strings.TrimSpace(reason)

	repo, err := s.approvalRepository()
	if err != nil {
		return nil, err
	}
	if err := repo.RejectMarket(ctx, marketID, actorUsername, now, market.RejectionReason); err != nil {
		return nil, err
	}
	refundAmount := market.ProposalCost
	if refundAmount == 0 {
		refundAmount = s.config.CreateMarketCost
	}
	if refundAmount > 0 && s.userService != nil {
		if err := s.userService.ApplyTransaction(ctx, market.CreatorUsername, refundAmount, users.TransactionRefund); err != nil {
			return nil, err
		}
	}
	return market, nil
}

func (s *Service) approvalRepository() (MarketApprovalRepository, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	repo, ok := s.repo.(MarketApprovalRepository)
	if !ok {
		return nil, ErrInvalidInput
	}
	return repo, nil
}
