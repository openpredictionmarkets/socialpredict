package users_test

import (
	"testing"
	"time"

	users "socialpredict/internal/domain/users"
)

func TestUserRoleStateDistinguishesModeratorStatuses(t *testing.T) {
	regular := &users.User{Username: "regular", UserType: "regular"}
	regular.NormalizeRoleState()
	if regular.UserType != string(users.UserTypeRegular) || regular.ModeratorStatus != users.ModeratorStatusNone {
		t.Fatalf("regular role state = %s/%s", regular.UserType, regular.ModeratorStatus)
	}
	if regular.IsModerator() || regular.IsActiveModerator() || regular.IsSuspendedModerator() {
		t.Fatalf("regular user must not look like a moderator")
	}

	moderator := &users.User{Username: "mod", UserType: "MODERATOR"}
	moderator.NormalizeRoleState()
	if !moderator.IsActiveModerator() || moderator.IsSuspendedModerator() {
		t.Fatalf("new moderator should normalize active, got %s", moderator.ModeratorStatus)
	}

	suspendedAt := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	if err := moderator.SuspendModerator("admin", "policy review", suspendedAt); err != nil {
		t.Fatalf("SuspendModerator returned error: %v", err)
	}
	if !moderator.IsSuspendedModerator() || moderator.IsActiveModerator() {
		t.Fatalf("suspended moderator not distinguishable: %+v", moderator)
	}
	if moderator.ModeratorSuspendedAt == nil || !moderator.ModeratorSuspendedAt.Equal(suspendedAt) {
		t.Fatalf("suspension timestamp not preserved: %+v", moderator.ModeratorSuspendedAt)
	}

	if err := moderator.UnsuspendModerator(); err != nil {
		t.Fatalf("UnsuspendModerator returned error: %v", err)
	}
	if !moderator.IsActiveModerator() || moderator.ModeratorSuspensionReason != "" || moderator.ModeratorSuspendedAt != nil {
		t.Fatalf("unsuspended moderator retained suspension state: %+v", moderator)
	}
}

func TestRegularUserCannotBeSuspendedAsModerator(t *testing.T) {
	regular := &users.User{Username: "regular", UserType: string(users.UserTypeRegular)}
	if err := regular.SuspendModerator("admin", "policy review", time.Now()); err != users.ErrInvalidModeratorState {
		t.Fatalf("SuspendModerator error = %v, want ErrInvalidModeratorState", err)
	}
}
