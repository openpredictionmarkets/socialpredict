package markets

import (
	"context"
	"strings"
	"time"

	users "socialpredict/internal/domain/users"
)

// MarketStewardshipRepository persists steward reassignment and audit facts.
type MarketStewardshipRepository interface {
	ReassignMarketSteward(ctx context.Context, marketID int64, fromStewardUsername string, toStewardUsername string, actorUsername string, reason string, changedAt time.Time) error
}

type MarketGroupStewardshipRepository interface {
	ReassignMarketGroupSteward(ctx context.Context, groupID int64, fromStewardUsername string, toStewardUsername string, actorUsername string, reason string, changedAt time.Time) error
}

// MarketStewardshipAuditRecord captures a market stewardship reassignment fact.
type MarketStewardshipAuditRecord struct {
	ID                  int64
	MarketID            int64
	FromStewardUsername string
	ToStewardUsername   string
	ActorUsername       string
	Reason              string
	CreatedAt           time.Time
}

func (s *Service) ReassignMarketSteward(ctx context.Context, marketID int64, newStewardUsername string, actorUsername string, reason string) (*Market, error) {
	newStewardUsername = strings.TrimSpace(newStewardUsername)
	actorUsername = strings.TrimSpace(actorUsername)
	reason = strings.TrimSpace(reason)
	if marketID <= 0 || newStewardUsername == "" || actorUsername == "" || reason == "" {
		return nil, ErrInvalidInput
	}

	market, err := s.GetMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if market.IsResolved() {
		return nil, ErrInvalidState
	}

	if err := s.validateAssignableSteward(ctx, newStewardUsername); err != nil {
		return nil, err
	}

	fromSteward := market.CurrentStewardUsername()
	if strings.EqualFold(fromSteward, newStewardUsername) {
		return market, nil
	}

	repo, err := s.stewardshipRepository()
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()
	if err := repo.ReassignMarketSteward(ctx, marketID, fromSteward, newStewardUsername, actorUsername, reason, now); err != nil {
		return nil, err
	}
	market.StewardUsername = newStewardUsername
	market.UpdatedAt = now
	market.StewardshipAudits = append(market.StewardshipAudits, MarketStewardshipAuditRecord{
		MarketID:            marketID,
		FromStewardUsername: fromSteward,
		ToStewardUsername:   newStewardUsername,
		ActorUsername:       actorUsername,
		Reason:              reason,
		CreatedAt:           now,
	})
	return market, nil
}

func (s *Service) ReassignMarketGroupSteward(ctx context.Context, groupID int64, newStewardUsername string, actorUsername string, reason string) (*MarketGroup, error) {
	newStewardUsername = strings.TrimSpace(newStewardUsername)
	actorUsername = strings.TrimSpace(actorUsername)
	reason = strings.TrimSpace(reason)
	if groupID <= 0 || newStewardUsername == "" || actorUsername == "" || reason == "" {
		return nil, ErrInvalidInput
	}

	groupRepo, err := s.marketGroupRepository()
	if err != nil {
		return nil, err
	}
	group, err := groupRepo.GetMarketGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrMarketGroupNotFound
	}
	switch NormalizeLifecycleStatus(group.LifecycleStatus) {
	case MarketLifecycleRejected, MarketLifecycleCancelled, MarketLifecycleResolved:
		return nil, ErrInvalidState
	}

	if err := s.validateAssignableSteward(ctx, newStewardUsername); err != nil {
		return nil, err
	}

	fromSteward := strings.TrimSpace(group.StewardUsername)
	if fromSteward == "" {
		fromSteward = strings.TrimSpace(group.CreatorUsername)
	}
	if strings.EqualFold(fromSteward, newStewardUsername) {
		return group, nil
	}

	groupStewardshipRepo, err := s.marketGroupStewardshipRepository()
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	if err := groupStewardshipRepo.ReassignMarketGroupSteward(ctx, groupID, fromSteward, newStewardUsername, actorUsername, reason, now); err != nil {
		return nil, err
	}
	group.StewardUsername = newStewardUsername
	group.UpdatedAt = now
	return group, nil
}

func (s *Service) validateAssignableSteward(ctx context.Context, username string) error {
	if s.userService == nil {
		return ErrUnauthorized
	}
	user, err := s.userService.GetPublicUser(ctx, username)
	if err != nil {
		return ErrUserNotFound
	}
	if user == nil || users.NormalizeUserType(user.UserType) != users.UserTypeModerator || users.NormalizeModeratorStatus(user.UserType, string(user.ModeratorStatus)) != users.ModeratorStatusActive {
		return ErrUnauthorized
	}
	return nil
}

func (s *Service) ensureMarketGovernanceActor(ctx context.Context, market *Market, username string) error {
	username = strings.TrimSpace(username)
	if market == nil || username == "" || s.userService == nil {
		return ErrUnauthorized
	}
	actor, err := s.userService.GetPublicUser(ctx, username)
	if err != nil {
		return ErrUserNotFound
	}
	if actor == nil {
		return ErrUnauthorized
	}

	userType := users.NormalizeUserType(actor.UserType)
	moderatorStatus := users.NormalizeModeratorStatus(actor.UserType, string(actor.ModeratorStatus))
	if userType == users.UserTypeAdmin {
		return nil
	}
	if userType == users.UserTypeModerator && moderatorStatus == users.ModeratorStatusSuspended {
		return ErrUnauthorized
	}
	if market.StewardedBy(username) {
		if !s.moderatorModeEnabled() || (userType == users.UserTypeModerator && moderatorStatus == users.ModeratorStatusActive) {
			return nil
		}
	}
	return ErrUnauthorized
}

func (s *Service) stewardshipRepository() (MarketStewardshipRepository, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	repo, ok := s.repo.(MarketStewardshipRepository)
	if !ok {
		return nil, ErrInvalidInput
	}
	return repo, nil
}

func (s *Service) marketGroupStewardshipRepository() (MarketGroupStewardshipRepository, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	repo, ok := s.repo.(MarketGroupStewardshipRepository)
	if !ok {
		return nil, ErrInvalidInput
	}
	return repo, nil
}
