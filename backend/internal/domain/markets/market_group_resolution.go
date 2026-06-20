package markets

import (
	"context"
	"strings"

	users "socialpredict/internal/domain/users"
)

// ResolveMarketGroup resolves every child binary market in a grouped market.
// The transaction boundary remains the child market; the parent is only marked
// resolved after all children use the normal binary resolution path.
func (s *Service) ResolveMarketGroup(ctx context.Context, groupID int64, req MarketGroupResolveRequest, username string) (*MarketGroup, error) {
	if uow, ok := s.groupedMarketUnitOfWork(); ok {
		var resolved *MarketGroup
		err := uow.GroupedMarketTransaction(ctx, func(txCtx context.Context, repo Repository, users UserService) error {
			var err error
			resolved, err = s.withTransactionDependencies(repo, users).resolveMarketGroup(txCtx, groupID, req, username)
			return err
		})
		if err != nil {
			return nil, err
		}
		return resolved, nil
	}
	return s.resolveMarketGroup(ctx, groupID, req, username)
}

func (s *Service) resolveMarketGroup(ctx context.Context, groupID int64, req MarketGroupResolveRequest, username string) (*MarketGroup, error) {
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

	resolverUsername := username
	if !group.StewardedBy(username) {
		resolverUsername = group.CurrentStewardUsername()
	}

	for _, member := range OrderedMarketGroupMembers(group.Members) {
		resolution := resolutions[member.MarketID]
		if err := s.resolveMarket(ctx, member.MarketID, resolution, username, false); err != nil {
			return nil, err
		}
	}

	resolvedAt := s.clock.Now()
	if err := groupRepo.MarkMarketGroupResolved(ctx, group.ID, resolvedAt); err != nil {
		return nil, err
	}
	group.LifecycleStatus = MarketLifecycleResolved
	group.UpdatedAt = resolvedAt
	if err := s.applyMarketGroupWorkProfit(ctx, group, resolverUsername); err != nil {
		return nil, err
	}
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
			return &MarketGroupChildNotPublishedError{
				MarketID:        member.MarketID,
				AnswerLabel:     member.AnswerLabel,
				LifecycleStatus: NormalizeLifecycleStatus(market.LifecycleStatus),
			}
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

func (s *Service) applyMarketGroupWorkProfit(ctx context.Context, group *MarketGroup, stewardUsername string) error {
	if group == nil || stewardUsername == "" || s.config.InitialBetFee <= 0 {
		return nil
	}
	income, err := s.calculateMarketGroupWorkFeePayout(ctx, group)
	if err != nil {
		return err
	}
	if income <= 0 {
		return nil
	}
	return s.userService.ApplyTransaction(ctx, stewardUsername, income, users.TransactionWorkProfit)
}

func (s *Service) calculateMarketGroupWorkFeePayout(ctx context.Context, group *MarketGroup) (int64, error) {
	if group == nil {
		return 0, nil
	}
	betsByAnswer := make([][]*Bet, 0, len(group.Members))
	for _, member := range OrderedMarketGroupMembers(group.Members) {
		bets, err := s.repo.ListBetsForMarket(ctx, member.MarketID)
		if err != nil {
			return 0, err
		}
		betsByAnswer = append(betsByAnswer, bets)
	}
	return ModeratorGroupWorkFeeIncome(betsByAnswer, s.config.InitialBetFee), nil
}

// ModeratorGroupWorkFeeIncome derives first-participation fee income for a
// grouped multiple-choice binary market. A participant counts once across the
// group, even if they trade several child answer markets.
func ModeratorGroupWorkFeeIncome(betsByAnswer [][]*Bet, initialBetFee int64) int64 {
	if initialBetFee <= 0 {
		return 0
	}
	participants := make(map[string]struct{})
	for _, answerBets := range betsByAnswer {
		for _, bet := range answerBets {
			if bet == nil || bet.Amount <= 0 || bet.Username == "" {
				continue
			}
			participants[bet.Username] = struct{}{}
		}
	}
	return int64(len(participants)) * initialBetFee
}

// ModeratorGroupWorkProfitIncome returns net grouped work profit after
// subtracting the parent proposal cost from collected fee income. It can be
// negative; the resolution payout itself is fee income.
func ModeratorGroupWorkProfitIncome(betsByAnswer [][]*Bet, initialBetFee int64, creationCost int64) int64 {
	return ModeratorGroupWorkFeeIncome(betsByAnswer, initialBetFee) - creationCost
}
