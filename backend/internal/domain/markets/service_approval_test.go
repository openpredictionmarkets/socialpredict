package markets_test

import (
	"context"
	"errors"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
)

func proposedMarket(id int64, now time.Time) *markets.Market {
	return &markets.Market{
		ID:                 id,
		QuestionTitle:      "Proposal",
		Status:             markets.MarketLifecycleProposed,
		LifecycleStatus:    markets.MarketLifecycleProposed,
		ResolutionDateTime: now.Add(24 * time.Hour),
	}
}

func TestApproveProposedMarketPublishesAndRecordsApproval(t *testing.T) {
	now := marketsTestTime()
	market := proposedMarket(77, now)
	var approvedBy string
	var approvedAt time.Time
	repo := newProjectionRepo(withProjectionRepoMarket(market), func(repo *projectionRepo) {
		repo.approveMarketFunc = func(_ context.Context, id int64, actor string, at time.Time) error {
			if id != market.ID {
				t.Fatalf("id = %d, want %d", id, market.ID)
			}
			approvedBy = actor
			approvedAt = at
			return nil
		}
	})
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	approved, err := service.ApproveProposedMarket(context.Background(), market.ID, "admin", true)
	if err != nil {
		t.Fatalf("ApproveProposedMarket returned error: %v", err)
	}
	if approved.Status != markets.MarketStatusActive || approved.LifecycleStatus != markets.MarketLifecyclePublished {
		t.Fatalf("unexpected approved market: %+v", approved)
	}
	if approvedBy != "admin" || !approvedAt.Equal(now) || approved.ApprovedBy != "admin" || approved.ApprovedAt == nil || !approved.ApprovedAt.Equal(now) {
		t.Fatalf("approval metadata mismatch: market=%+v repoBy=%q repoAt=%s", approved, approvedBy, approvedAt)
	}
}

func TestRejectProposedMarketRecordsReason(t *testing.T) {
	now := marketsTestTime()
	market := proposedMarket(78, now)
	var rejectedBy string
	var rejectionReason string
	repo := newProjectionRepo(withProjectionRepoMarket(market), func(repo *projectionRepo) {
		repo.rejectMarketFunc = func(_ context.Context, id int64, actor string, at time.Time, reason string) error {
			rejectedBy = actor
			rejectionReason = reason
			return nil
		}
	})
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	rejected, err := service.RejectProposedMarket(context.Background(), market.ID, "admin", " out of scope ")
	if err != nil {
		t.Fatalf("RejectProposedMarket returned error: %v", err)
	}
	if rejected.Status != markets.MarketLifecycleRejected || rejected.LifecycleStatus != markets.MarketLifecycleRejected {
		t.Fatalf("unexpected rejected market: %+v", rejected)
	}
	if rejectedBy != "admin" || rejectionReason != "out of scope" || rejected.RejectionReason != "out of scope" {
		t.Fatalf("rejection metadata mismatch: market=%+v repoBy=%q repoReason=%q", rejected, rejectedBy, rejectionReason)
	}
}

func TestRejectProposedMarketRefundsCreationCost(t *testing.T) {
	now := marketsTestTime()
	market := proposedMarket(80, now)
	market.CreatorUsername = "moderator"
	market.ProposalCost = 7
	repo := newProjectionRepo(withProjectionRepoMarket(market), func(repo *projectionRepo) {
		repo.rejectMarketFunc = func(context.Context, int64, string, time.Time, string) error {
			return nil
		}
	})
	var refundedUsername string
	var refundedAmount int64
	var refundType string
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.applyTransactionFunc = func(_ context.Context, username string, amount int64, transactionType string) error {
			refundedUsername = username
			refundedAmount = amount
			refundType = transactionType
			return nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{CreateMarketCost: 10})

	if _, err := service.RejectProposedMarket(context.Background(), market.ID, "admin", "duplicate"); err != nil {
		t.Fatalf("RejectProposedMarket returned error: %v", err)
	}
	if refundedUsername != "moderator" || refundedAmount != 7 || refundType != dusers.TransactionRefund {
		t.Fatalf("refund mismatch: username=%q amount=%d type=%q", refundedUsername, refundedAmount, refundType)
	}
}

func TestRejectProposedMarketFallsBackToConfigCostForLegacyProposals(t *testing.T) {
	now := marketsTestTime()
	market := proposedMarket(81, now)
	market.CreatorUsername = "moderator"
	repo := newProjectionRepo(withProjectionRepoMarket(market), func(repo *projectionRepo) {
		repo.rejectMarketFunc = func(context.Context, int64, string, time.Time, string) error {
			return nil
		}
	})
	var refundedAmount int64
	usersSvc := newNoopUserService(func(service *noopUserService) {
		service.applyTransactionFunc = func(_ context.Context, _ string, amount int64, _ string) error {
			refundedAmount = amount
			return nil
		}
	})
	service := markets.NewService(repo, usersSvc, newFixedClock(now), markets.Config{CreateMarketCost: 10})

	if _, err := service.RejectProposedMarket(context.Background(), market.ID, "admin", "duplicate"); err != nil {
		t.Fatalf("RejectProposedMarket returned error: %v", err)
	}
	if refundedAmount != 10 {
		t.Fatalf("fallback refund amount = %d, want 10", refundedAmount)
	}
}

func TestReviewProposedMarketRejectsWrongStateAndMissingConfirmation(t *testing.T) {
	now := marketsTestTime()
	published := &markets.Market{ID: 79, LifecycleStatus: markets.MarketLifecyclePublished, Status: markets.MarketStatusActive, ResolutionDateTime: now.Add(24 * time.Hour)}
	repo := newProjectionRepo(withProjectionRepoMarket(published), func(repo *projectionRepo) {
		repo.approveMarketFunc = func(context.Context, int64, string, time.Time) error {
			t.Fatalf("approve should not be persisted for wrong state")
			return nil
		}
		repo.rejectMarketFunc = func(context.Context, int64, string, time.Time, string) error {
			t.Fatalf("reject should not be persisted for wrong state")
			return nil
		}
	})
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	if _, err := service.ApproveProposedMarket(context.Background(), published.ID, "admin", false); !errors.Is(err, markets.ErrInvalidInput) {
		t.Fatalf("ApproveProposedMarket missing confirmation error = %v, want ErrInvalidInput", err)
	}
	if _, err := service.ApproveProposedMarket(context.Background(), published.ID, "admin", true); !errors.Is(err, markets.ErrInvalidState) {
		t.Fatalf("ApproveProposedMarket wrong state error = %v, want ErrInvalidState", err)
	}
	if _, err := service.RejectProposedMarket(context.Background(), published.ID, "admin", "bad"); !errors.Is(err, markets.ErrInvalidState) {
		t.Fatalf("RejectProposedMarket wrong state error = %v, want ErrInvalidState", err)
	}
	if _, err := service.RejectProposedMarket(context.Background(), published.ID, "admin", " "); !errors.Is(err, markets.ErrInvalidInput) {
		t.Fatalf("RejectProposedMarket missing reason error = %v, want ErrInvalidInput", err)
	}
}
