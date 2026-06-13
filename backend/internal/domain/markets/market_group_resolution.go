package markets

import (
	"context"
	"strings"
)

// ResolveMarketGroup resolves every child binary market in a grouped market.
// The transaction boundary remains the child market; the parent is only marked
// resolved after all children use the normal binary resolution path.
func (s *Service) ResolveMarketGroup(ctx context.Context, groupID int64, req MarketGroupResolveRequest, username string) (*MarketGroup, error) {
	if groupID <= 0 {
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
	if NormalizeLifecycleStatus(group.LifecycleStatus) != MarketLifecyclePublished {
		return nil, ErrInvalidState
	}

	resolutions, err := s.resolveGroupChildOutcomes(group, req)
	if err != nil {
		return nil, err
	}
	if err := s.validateMarketGroupResolutionChildren(ctx, group, resolutions, username); err != nil {
		return nil, err
	}

	for _, member := range OrderedMarketGroupMembers(group.Members) {
		resolution := resolutions[member.MarketID]
		if err := s.ResolveMarket(ctx, member.MarketID, resolution, username); err != nil {
			return nil, err
		}
	}

	resolvedAt := s.clock.Now()
	if err := groupRepo.MarkMarketGroupResolved(ctx, group.ID, resolvedAt); err != nil {
		return nil, err
	}
	group.LifecycleStatus = MarketLifecycleResolved
	group.UpdatedAt = resolvedAt
	return group, nil
}

func (s *Service) resolveGroupChildOutcomes(group *MarketGroup, req MarketGroupResolveRequest) (map[int64]string, error) {
	childIDs := map[int64]struct{}{}
	for _, member := range group.Members {
		childIDs[member.MarketID] = struct{}{}
	}
	if len(childIDs) == 0 {
		return nil, ErrInvalidState
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	switch mode {
	case MarketGroupResolveModeExclusiveYes:
		if _, ok := childIDs[req.WinningMarketID]; !ok {
			return nil, ErrInvalidInput
		}
		resolutions := make(map[int64]string, len(childIDs))
		for childID := range childIDs {
			resolutions[childID] = "NO"
		}
		resolutions[req.WinningMarketID] = "YES"
		return resolutions, nil
	case MarketGroupResolveModeManual:
		if len(req.Resolutions) != len(childIDs) {
			return nil, ErrInvalidInput
		}
		resolutions := make(map[int64]string, len(childIDs))
		for _, item := range req.Resolutions {
			if _, ok := childIDs[item.MarketID]; !ok {
				return nil, ErrInvalidInput
			}
			outcome, err := s.resolutionPolicy.NormalizeResolution(item.Resolution)
			if err != nil || (outcome != "YES" && outcome != "NO") {
				return nil, ErrInvalidInput
			}
			resolutions[item.MarketID] = outcome
		}
		if len(resolutions) != len(childIDs) {
			return nil, ErrInvalidInput
		}
		return resolutions, nil
	default:
		return nil, ErrInvalidInput
	}
}

func (s *Service) validateMarketGroupResolutionChildren(ctx context.Context, group *MarketGroup, resolutions map[int64]string, username string) error {
	for _, member := range group.Members {
		market, err := s.repo.GetByID(ctx, member.MarketID)
		if err != nil || market == nil {
			return ErrMarketNotFound
		}
		if _, ok := resolutions[member.MarketID]; !ok {
			return ErrInvalidInput
		}
		if market.LifecycleStatus != "" && NormalizeLifecycleStatus(market.LifecycleStatus) != MarketLifecyclePublished {
			return ErrInvalidState
		}
		if err := s.ensureMarketGovernanceActor(ctx, market, username); err != nil {
			return err
		}
		resolverUsername := username
		if !market.StewardedBy(username) {
			resolverUsername = market.CurrentStewardUsername()
		}
		if err := s.resolutionPolicy.ValidateResolutionRequest(market, resolverUsername); err != nil {
			return err
		}
	}
	return nil
}
