package users_test

import (
	"context"
	"errors"
	"testing"
	"time"

	users "socialpredict/internal/domain/users"
	"socialpredict/security"
)

func (f *fakeRepository) CreateModeratorAudit(ctx context.Context, record *users.ModeratorAuditRecord) error {
	if record == nil {
		return users.ErrInvalidUserData
	}
	f.moderatorAudit = append(f.moderatorAudit, *record)
	return nil
}

func TestServiceModeratorPromotionSuspensionAndUnsuspension(t *testing.T) {
	repo := newFakeRepository("candidate")
	service := users.NewServiceWithDependencies(users.ServiceDependencies{
		Reader:         repo,
		Writer:         repo,
		ModeratorAudit: repo,
	}, nil, security.NewSecurityService().Sanitizer)
	ctx := context.Background()

	promoted, err := service.PromoteToModerator(ctx, "candidate", "admin", "trusted participant")
	if err != nil {
		t.Fatalf("PromoteToModerator returned error: %v", err)
	}
	if promoted.UserType != string(users.UserTypeModerator) || promoted.ModeratorStatus != users.ModeratorStatusActive {
		t.Fatalf("unexpected promoted user: %+v", promoted)
	}

	suspendedAt := time.Date(2026, 5, 24, 15, 30, 0, 0, time.UTC)
	suspended, err := service.SuspendModerator(ctx, "candidate", "admin", "policy violation", suspendedAt)
	if err != nil {
		t.Fatalf("SuspendModerator returned error: %v", err)
	}
	if !suspended.IsSuspendedModerator() || suspended.ModeratorSuspensionReason != "policy violation" || suspended.ModeratorSuspendedBy != "admin" {
		t.Fatalf("unexpected suspended user: %+v", suspended)
	}

	unsuspended, err := service.UnsuspendModerator(ctx, "candidate", "admin", "appeal accepted")
	if err != nil {
		t.Fatalf("UnsuspendModerator returned error: %v", err)
	}
	if !unsuspended.IsActiveModerator() || unsuspended.ModeratorSuspendedAt != nil {
		t.Fatalf("unexpected unsuspended user: %+v", unsuspended)
	}

	if len(repo.moderatorAudit) != 3 {
		t.Fatalf("expected 3 audit records, got %d", len(repo.moderatorAudit))
	}
	if repo.moderatorAudit[0].Action != users.ModeratorAuditActionPromote || repo.moderatorAudit[1].Action != users.ModeratorAuditActionSuspend || repo.moderatorAudit[2].Action != users.ModeratorAuditActionUnsuspend {
		t.Fatalf("unexpected audit actions: %+v", repo.moderatorAudit)
	}
}

func TestServiceRejectsSuspendingRegularUser(t *testing.T) {
	repo := newFakeRepository("regular")
	service := users.NewServiceWithDependencies(users.ServiceDependencies{Reader: repo, Writer: repo}, nil, security.NewSecurityService().Sanitizer)

	_, err := service.SuspendModerator(context.Background(), "regular", "admin", "not allowed", time.Now())
	if !errors.Is(err, users.ErrInvalidModeratorState) {
		t.Fatalf("SuspendModerator error = %v, want ErrInvalidModeratorState", err)
	}
}
