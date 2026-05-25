package users

import (
	"strings"
	"time"
)

type UserType string

const (
	UserTypeAdmin     UserType = "ADMIN"
	UserTypeRegular   UserType = "REGULAR"
	UserTypeModerator UserType = "MODERATOR"
)

type ModeratorStatus string

const (
	ModeratorStatusNone      ModeratorStatus = "none"
	ModeratorStatusActive    ModeratorStatus = "active"
	ModeratorStatusSuspended ModeratorStatus = "suspended"
)

type ModeratorAuditAction string

const (
	ModeratorAuditActionPromote   ModeratorAuditAction = "promote"
	ModeratorAuditActionSuspend   ModeratorAuditAction = "suspend"
	ModeratorAuditActionUnsuspend ModeratorAuditAction = "unsuspend"
)

// ModeratorAuditRecord captures the governance seam for moderator role/status changes.
type ModeratorAuditRecord struct {
	ID                  int64
	Username            string
	ActorUsername       string
	Action              ModeratorAuditAction
	FromUserType        string
	ToUserType          string
	FromModeratorStatus ModeratorStatus
	ToModeratorStatus   ModeratorStatus
	Reason              string
	CreatedAt           time.Time
}

func NormalizeUserType(value string) UserType {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case string(UserTypeAdmin):
		return UserTypeAdmin
	case string(UserTypeModerator):
		return UserTypeModerator
	case "", "USER", string(UserTypeRegular):
		return UserTypeRegular
	default:
		return UserType(strings.ToUpper(strings.TrimSpace(value)))
	}
}

func NormalizeModeratorStatus(userType string, value string) ModeratorStatus {
	if NormalizeUserType(userType) != UserTypeModerator {
		return ModeratorStatusNone
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ModeratorStatusSuspended):
		return ModeratorStatusSuspended
	case string(ModeratorStatusActive), "":
		return ModeratorStatusActive
	default:
		return ModeratorStatus(strings.ToLower(strings.TrimSpace(value)))
	}
}

func (u *User) NormalizeRoleState() {
	if u == nil {
		return
	}
	u.UserType = string(NormalizeUserType(u.UserType))
	u.ModeratorStatus = NormalizeModeratorStatus(u.UserType, string(u.ModeratorStatus))
	if u.ModeratorStatus != ModeratorStatusSuspended {
		u.ModeratorSuspensionReason = ""
		u.ModeratorSuspendedBy = ""
		u.ModeratorSuspendedAt = nil
	}
}

func (u *User) IsModerator() bool {
	if u == nil {
		return false
	}
	return NormalizeUserType(u.UserType) == UserTypeModerator
}

func (u *User) IsActiveModerator() bool {
	if !u.IsModerator() {
		return false
	}
	return NormalizeModeratorStatus(u.UserType, string(u.ModeratorStatus)) == ModeratorStatusActive
}

func (u *User) IsSuspendedModerator() bool {
	if !u.IsModerator() {
		return false
	}
	return NormalizeModeratorStatus(u.UserType, string(u.ModeratorStatus)) == ModeratorStatusSuspended
}

func (u *User) PromoteToModerator() error {
	if u == nil {
		return ErrInvalidUserData
	}
	u.UserType = string(UserTypeModerator)
	u.ModeratorStatus = ModeratorStatusActive
	u.ModeratorSuspensionReason = ""
	u.ModeratorSuspendedBy = ""
	u.ModeratorSuspendedAt = nil
	return nil
}

func (u *User) SuspendModerator(actorUsername, reason string, suspendedAt time.Time) error {
	if u == nil {
		return ErrInvalidUserData
	}
	if !u.IsModerator() {
		return ErrInvalidModeratorState
	}
	if actorUsername == "" || strings.TrimSpace(reason) == "" {
		return ErrInvalidUserData
	}
	if suspendedAt.IsZero() {
		suspendedAt = time.Now().UTC()
	}

	u.UserType = string(UserTypeModerator)
	u.ModeratorStatus = ModeratorStatusSuspended
	u.ModeratorSuspensionReason = strings.TrimSpace(reason)
	u.ModeratorSuspendedBy = actorUsername
	u.ModeratorSuspendedAt = &suspendedAt
	return nil
}

func (u *User) UnsuspendModerator() error {
	if u == nil {
		return ErrInvalidUserData
	}
	if !u.IsModerator() {
		return ErrInvalidModeratorState
	}
	u.UserType = string(UserTypeModerator)
	u.ModeratorStatus = ModeratorStatusActive
	u.ModeratorSuspensionReason = ""
	u.ModeratorSuspendedBy = ""
	u.ModeratorSuspendedAt = nil
	return nil
}
