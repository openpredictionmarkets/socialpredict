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

type MarketGroupApprovalRepository interface {
	ApproveMarketGroup(ctx context.Context, groupID int64, actorUsername string, approvedAt time.Time) error
	RejectMarketGroup(ctx context.Context, groupID int64, actorUsername string, rejectedAt time.Time, reason string) error
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

func (s *Service) ApproveProposedMarketGroup(ctx context.Context, groupID int64, actorUsername string, confirmed bool) (*MarketGroup, error) {
	if groupID <= 0 || strings.TrimSpace(actorUsername) == "" || !confirmed {
		return nil, ErrInvalidInput
	}

	groupReadRepo, err := s.marketGroupRepository()
	if err != nil {
		return nil, err
	}
	groupWriteRepo, err := s.marketGroupApprovalRepository()
	if err != nil {
		return nil, err
	}
	group, err := groupReadRepo.GetMarketGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if NormalizeLifecycleStatus(group.LifecycleStatus) != MarketLifecycleProposed {
		return nil, ErrInvalidState
	}

	now := s.clock.Now()
	if err := groupWriteRepo.ApproveMarketGroup(ctx, groupID, actorUsername, now); err != nil {
		return nil, err
	}
	group.LifecycleStatus = MarketLifecyclePublished
	group.ApprovedBy = actorUsername
	group.ApprovedAt = &now
	group.RejectedBy = ""
	group.RejectedAt = nil
	group.RejectionReason = ""
	return group, nil
}

func (s *Service) RejectProposedMarketGroup(ctx context.Context, groupID int64, actorUsername string, reason string) (*MarketGroup, error) {
	rejectionReason := strings.TrimSpace(reason)
	if groupID <= 0 || strings.TrimSpace(actorUsername) == "" || rejectionReason == "" {
		return nil, ErrInvalidInput
	}

	groupReadRepo, err := s.marketGroupRepository()
	if err != nil {
		return nil, err
	}
	groupWriteRepo, err := s.marketGroupApprovalRepository()
	if err != nil {
		return nil, err
	}
	group, err := groupReadRepo.GetMarketGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if NormalizeLifecycleStatus(group.LifecycleStatus) != MarketLifecycleProposed {
		return nil, ErrInvalidState
	}

	now := s.clock.Now()
	if err := groupWriteRepo.RejectMarketGroup(ctx, groupID, actorUsername, now, rejectionReason); err != nil {
		return nil, err
	}
	refundAmount := group.ProposalCost
	if refundAmount == 0 {
		refundAmount = s.config.CreateMarketCost
	}
	if refundAmount > 0 && s.userService != nil {
		if err := s.userService.ApplyTransaction(ctx, group.CreatorUsername, refundAmount, users.TransactionRefund); err != nil {
			return nil, err
		}
	}
	group.LifecycleStatus = MarketLifecycleRejected
	group.RejectedBy = actorUsername
	group.RejectedAt = &now
	group.RejectionReason = rejectionReason
	return group, nil
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

func (s *Service) marketGroupApprovalRepository() (MarketGroupApprovalRepository, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	repo, ok := s.repo.(MarketGroupApprovalRepository)
	if !ok {
		return nil, ErrInvalidInput
	}
	return repo, nil
}
