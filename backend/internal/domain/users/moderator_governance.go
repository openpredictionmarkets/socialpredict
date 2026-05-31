package users

import (
	"context"
	"strings"
	"time"
)

// ModeratorAuditWriter persists moderator role/status governance events.
type ModeratorAuditWriter interface {
	CreateModeratorAudit(ctx context.Context, record *ModeratorAuditRecord) error
}

func (s *Service) PromoteToModerator(ctx context.Context, username, actorUsername, reason string) (*User, error) {
	if err := validateModeratorTransitionInput(username, actorUsername); err != nil {
		return nil, err
	}

	var audit *ModeratorAuditRecord
	user, err := s.updateUserProfile(ctx, username, func(user *User) error {
		audit = newModeratorAuditRecord(user, actorUsername, ModeratorAuditActionPromote, reason)
		if err := user.PromoteToModerator(); err != nil {
			return err
		}
		audit.ToUserType = user.UserType
		audit.ToModeratorStatus = user.ModeratorStatus
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := s.writeModeratorAudit(ctx, audit); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) SuspendModerator(ctx context.Context, username, actorUsername, reason string, suspendedAt time.Time) (*User, error) {
	if err := validateModeratorTransitionInput(username, actorUsername); err != nil {
		return nil, err
	}
	if strings.TrimSpace(reason) == "" {
		return nil, ErrInvalidUserData
	}

	var audit *ModeratorAuditRecord
	user, err := s.updateUserProfile(ctx, username, func(user *User) error {
		audit = newModeratorAuditRecord(user, actorUsername, ModeratorAuditActionSuspend, reason)
		if err := user.SuspendModerator(actorUsername, reason, suspendedAt); err != nil {
			return err
		}
		audit.ToUserType = user.UserType
		audit.ToModeratorStatus = user.ModeratorStatus
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := s.writeModeratorAudit(ctx, audit); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) UnsuspendModerator(ctx context.Context, username, actorUsername, reason string) (*User, error) {
	if err := validateModeratorTransitionInput(username, actorUsername); err != nil {
		return nil, err
	}

	var audit *ModeratorAuditRecord
	user, err := s.updateUserProfile(ctx, username, func(user *User) error {
		audit = newModeratorAuditRecord(user, actorUsername, ModeratorAuditActionUnsuspend, reason)
		if err := user.UnsuspendModerator(); err != nil {
			return err
		}
		audit.ToUserType = user.UserType
		audit.ToModeratorStatus = user.ModeratorStatus
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := s.writeModeratorAudit(ctx, audit); err != nil {
		return nil, err
	}
	return user, nil
}

func validateModeratorTransitionInput(username, actorUsername string) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	if err := validateUsername(actorUsername); err != nil {
		return err
	}
	return nil
}

func newModeratorAuditRecord(user *User, actorUsername string, action ModeratorAuditAction, reason string) *ModeratorAuditRecord {
	if user == nil {
		return nil
	}
	return &ModeratorAuditRecord{
		Username:            user.Username,
		ActorUsername:       actorUsername,
		Action:              action,
		FromUserType:        user.UserType,
		ToUserType:          user.UserType,
		FromModeratorStatus: NormalizeModeratorStatus(user.UserType, string(user.ModeratorStatus)),
		ToModeratorStatus:   NormalizeModeratorStatus(user.UserType, string(user.ModeratorStatus)),
		Reason:              strings.TrimSpace(reason),
	}
}

func (s *Service) writeModeratorAudit(ctx context.Context, record *ModeratorAuditRecord) error {
	if record == nil || s == nil || s.moderatorAudit == nil {
		return nil
	}
	return s.moderatorAudit.CreateModeratorAudit(ctx, record)
}
