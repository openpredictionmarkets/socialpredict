package analytics

import (
	"context"
	"errors"
	"time"
)

// RefreshUserFinancialMetricSnapshot recomputes and stores an authenticated
// display/read-model financial snapshot from canonical position data.
func (s *Service) RefreshUserFinancialMetricSnapshot(ctx context.Context, req FinancialSnapshotRequest, generatedAt time.Time) (*UserFinancialMetricSnapshot, error) {
	if req.Username == "" {
		return nil, errors.New("username is required")
	}
	if s.financialsRepo == nil {
		return nil, errors.New("financials repository not provided")
	}

	snapshotRepo, ok := s.repo.(UserFinancialMetricSnapshotRepository)
	if !ok {
		return nil, errors.New("user financial metric snapshot repository not provided")
	}

	positions, err := s.financialsRepo.UserMarketPositions(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	workProfits, err := s.computeUserWorkProfits(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	req.WorkProfits = workProfits
	unrealizedWorkIncome, unrealizedWorkProfits, err := s.computeUserUnrealizedWorkFinancials(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	req.UnrealizedWorkIncome = unrealizedWorkIncome
	req.UnrealizedWorkProfits = unrealizedWorkProfits

	snapshot := NewUserFinancialMetricSnapshotCalculator(s.config).
		Calculate(req, positions, generatedAt)
	if err := snapshotRepo.UpsertUserFinancialMetricSnapshot(ctx, snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// GetUserFinancialMetricReadModel returns a stored authenticated display
// snapshot with freshness metadata. A missing snapshot is not an error.
func (s *Service) GetUserFinancialMetricReadModel(ctx context.Context, username string) (*UserFinancialMetricReadModel, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}

	snapshotRepo, ok := s.repo.(UserFinancialMetricSnapshotRepository)
	if !ok {
		return nil, errors.New("user financial metric snapshot repository not provided")
	}

	snapshot, err := snapshotRepo.GetUserFinancialMetricSnapshot(ctx, username)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return nil, nil
	}

	return &UserFinancialMetricReadModel{
		Snapshot:  *snapshot,
		Freshness: snapshot.Freshness(),
	}, nil
}

// MarkUserFinancialMetricSnapshotStale marks an authenticated display snapshot
// stale after canonical user-affecting activity. It does not mutate balances,
// positions, payouts, or any transaction truth.
func (s *Service) MarkUserFinancialMetricSnapshotStale(ctx context.Context, username string, reason string) error {
	if username == "" {
		return errors.New("username is required")
	}

	snapshotRepo, ok := s.repo.(UserFinancialMetricSnapshotRepository)
	if !ok {
		return errors.New("user financial metric snapshot repository not provided")
	}
	return snapshotRepo.MarkUserFinancialMetricSnapshotStale(ctx, username, reason)
}
