package markets

import (
	"context"
	"sort"
)

// GetMarketGroupBetsPage returns one globally sorted bet feed for all child
// markets in a group. Child markets remain the transaction boundary.
func (s *Service) GetMarketGroupBetsPage(ctx context.Context, groupID int64, p Page) (*MarketGroupBetsPage, error) {
	group, members, err := s.marketGroupActivityMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	p = s.statusPolicy.NormalizePage(p, 20, 100)
	rows := make([]*MarketGroupBetDisplayInfo, 0)
	for _, member := range members {
		bets, err := s.getMarketBetDisplayInfos(ctx, member.MarketID)
		if err != nil {
			return nil, err
		}
		for _, bet := range bets {
			if bet == nil {
				continue
			}
			rows = append(rows, &MarketGroupBetDisplayInfo{
				AnswerMarketID: member.MarketID,
				AnswerLabel:    member.AnswerLabel,
				DisplayOrder:   member.DisplayOrder,
				Username:       bet.Username,
				Outcome:        bet.Outcome,
				Amount:         bet.Amount,
				Probability:    bet.Probability,
				PlacedAt:       bet.PlacedAt,
			})
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].PlacedAt.Equal(rows[j].PlacedAt) {
			if rows[i].DisplayOrder == rows[j].DisplayOrder {
				return rows[i].Username < rows[j].Username
			}
			return rows[i].DisplayOrder < rows[j].DisplayOrder
		}
		return rows[i].PlacedAt.After(rows[j].PlacedAt)
	})

	total := len(rows)
	return &MarketGroupBetsPage{
		GroupID: group.ID,
		Bets:    paginateMarketGroupBets(rows, p),
		Total:   total,
	}, nil
}

// GetMarketGroupPositionsPage returns grouped user positions with answer-level
// detail. It is display-only and must not be used for transaction decisions.
func (s *Service) GetMarketGroupPositionsPage(ctx context.Context, groupID int64, p Page) (*MarketGroupPositionsPage, error) {
	group, members, err := s.marketGroupActivityMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	byUser := make(map[string]*MarketGroupPositionRow)
	for _, member := range members {
		positions, err := s.GetMarketPositions(ctx, member.MarketID)
		if err != nil {
			return nil, err
		}
		for _, pos := range activeMarketPositions(positions) {
			if pos == nil || pos.Username == "" {
				continue
			}
			current := byUser[pos.Username]
			if current == nil {
				current = &MarketGroupPositionRow{Username: pos.Username}
				byUser[pos.Username] = current
			}
			current.YesSharesOwned += pos.YesSharesOwned
			current.NoSharesOwned += pos.NoSharesOwned
			current.Value += pos.Value
			current.TotalSpent += pos.TotalSpent
			current.TotalSpentInPlay += pos.TotalSpentInPlay
			current.Answers = append(current.Answers, &MarketGroupPositionAnswer{
				AnswerMarketID:   member.MarketID,
				AnswerLabel:      member.AnswerLabel,
				DisplayOrder:     member.DisplayOrder,
				MarketID:         pos.MarketID,
				YesSharesOwned:   pos.YesSharesOwned,
				NoSharesOwned:    pos.NoSharesOwned,
				Value:            pos.Value,
				TotalSpent:       pos.TotalSpent,
				TotalSpentInPlay: pos.TotalSpentInPlay,
				IsResolved:       pos.IsResolved,
				ResolutionResult: pos.ResolutionResult,
			})
		}
	}

	rows := make([]*MarketGroupPositionRow, 0, len(byUser))
	for _, row := range byUser {
		sort.SliceStable(row.Answers, func(i, j int) bool {
			return row.Answers[i].DisplayOrder < row.Answers[j].DisplayOrder
		})
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		left := rows[i].YesSharesOwned + rows[i].NoSharesOwned
		right := rows[j].YesSharesOwned + rows[j].NoSharesOwned
		if left == right {
			return rows[i].Username < rows[j].Username
		}
		return left > right
	})

	p = s.statusPolicy.NormalizePage(p, 20, 100)
	total := len(rows)
	return &MarketGroupPositionsPage{
		GroupID:   group.ID,
		Positions: paginateMarketGroupPositions(rows, p),
		Total:     total,
	}, nil
}

// GetMarketGroupLeaderboardPage returns an aggregate leaderboard across child
// markets with per-answer profit breakdowns for display.
func (s *Service) GetMarketGroupLeaderboardPage(ctx context.Context, groupID int64, p Page) (*MarketGroupLeaderboardPage, error) {
	group, members, err := s.marketGroupActivityMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	byUser := make(map[string]*MarketGroupLeaderboardRow)
	for _, member := range members {
		rows, err := s.getMarketLeaderboardRows(ctx, member.MarketID)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			if row == nil || row.Username == "" {
				continue
			}
			current := byUser[row.Username]
			if current == nil {
				current = &MarketGroupLeaderboardRow{Username: row.Username}
				byUser[row.Username] = current
			}
			current.Profit += row.Profit
			current.CurrentValue += row.CurrentValue
			current.TotalSpent += row.TotalSpent
			current.YesSharesOwned += row.YesSharesOwned
			current.NoSharesOwned += row.NoSharesOwned
			current.Answers = append(current.Answers, &MarketGroupLeaderboardAnswer{
				AnswerMarketID: member.MarketID,
				AnswerLabel:    member.AnswerLabel,
				DisplayOrder:   member.DisplayOrder,
				Profit:         row.Profit,
				CurrentValue:   row.CurrentValue,
				TotalSpent:     row.TotalSpent,
				Position:       row.Position,
				YesSharesOwned: row.YesSharesOwned,
				NoSharesOwned:  row.NoSharesOwned,
			})
		}
	}

	rows := make([]*MarketGroupLeaderboardRow, 0, len(byUser))
	for _, row := range byUser {
		row.Position = groupedLeaderboardPosition(row.YesSharesOwned, row.NoSharesOwned)
		sort.SliceStable(row.Answers, func(i, j int) bool {
			return row.Answers[i].DisplayOrder < row.Answers[j].DisplayOrder
		})
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Profit == rows[j].Profit {
			if rows[i].CurrentValue == rows[j].CurrentValue {
				return rows[i].Username < rows[j].Username
			}
			return rows[i].CurrentValue > rows[j].CurrentValue
		}
		return rows[i].Profit > rows[j].Profit
	})
	for index, row := range rows {
		row.Rank = index + 1
	}

	p = s.statusPolicy.NormalizePage(p, 20, 100)
	total := len(rows)
	return &MarketGroupLeaderboardPage{
		GroupID:     group.ID,
		Leaderboard: paginateMarketGroupLeaderboard(rows, p),
		Total:       total,
	}, nil
}

func (s *Service) marketGroupActivityMembers(ctx context.Context, groupID int64) (*MarketGroup, []MarketGroupMember, error) {
	if groupID <= 0 {
		return nil, nil, ErrInvalidInput
	}
	groupRepo, err := s.marketGroupRepository()
	if err != nil {
		return nil, nil, err
	}
	group, err := groupRepo.GetMarketGroup(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}
	if group == nil {
		return nil, nil, ErrMarketGroupNotFound
	}
	return group, OrderedMarketGroupMembers(group.Members), nil
}

func groupedLeaderboardPosition(yesShares int64, noShares int64) string {
	if yesShares > 0 && noShares > 0 {
		return "MIXED"
	}
	if yesShares > 0 {
		return "YES"
	}
	if noShares > 0 {
		return "NO"
	}
	return "NEUTRAL"
}

func paginateMarketGroupBets(rows []*MarketGroupBetDisplayInfo, p Page) []*MarketGroupBetDisplayInfo {
	if len(rows) == 0 || p.Offset >= len(rows) {
		return []*MarketGroupBetDisplayInfo{}
	}
	end := p.Offset + p.Limit
	if end > len(rows) {
		end = len(rows)
	}
	return rows[p.Offset:end]
}

func paginateMarketGroupPositions(rows []*MarketGroupPositionRow, p Page) []*MarketGroupPositionRow {
	if len(rows) == 0 || p.Offset >= len(rows) {
		return []*MarketGroupPositionRow{}
	}
	end := p.Offset + p.Limit
	if end > len(rows) {
		end = len(rows)
	}
	return rows[p.Offset:end]
}

func paginateMarketGroupLeaderboard(rows []*MarketGroupLeaderboardRow, p Page) []*MarketGroupLeaderboardRow {
	if len(rows) == 0 || p.Offset >= len(rows) {
		return []*MarketGroupLeaderboardRow{}
	}
	end := p.Offset + p.Limit
	if end > len(rows) {
		end = len(rows)
	}
	return rows[p.Offset:end]
}
