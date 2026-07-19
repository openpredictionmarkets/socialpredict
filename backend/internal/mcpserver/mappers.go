package mcpserver

import (
	"strings"

	dmarkets "socialpredict/internal/domain/markets"
)

func MarketTagOutputFromDomain(tag dmarkets.MarketTag) MarketTagOutput {
	return MarketTagOutput{ID: tag.ID, Slug: tag.Slug, DisplayName: tag.DisplayName, Description: tag.Description, ColorKey: tag.ColorKey, SortOrder: tag.SortOrder, IsActive: tag.IsActive}
}
func MarketTagOutputsFromDomain(tags []dmarkets.MarketTag) []MarketTagOutput {
	out := make([]MarketTagOutput, 0, len(tags))
	for _, tag := range tags {
		out = append(out, MarketTagOutputFromDomain(tag))
	}
	return out
}
func CreatorOutputFromDomain(s *dmarkets.CreatorSummary) *CreatorOutput {
	if s == nil {
		return nil
	}
	return &CreatorOutput{Username: s.Username, DisplayName: s.DisplayName, PersonalEmoji: s.PersonalEmoji}
}

func MarketOutputFromDomain(m *dmarkets.Market) MarketOutput {
	if m == nil {
		return MarketOutput{Tags: []MarketTagOutput{}}
	}
	return MarketOutput{ID: m.ID, QuestionTitle: m.QuestionTitle, Description: m.Description, OutcomeType: m.OutcomeType, ResolutionDateTime: m.ResolutionDateTime, FinalResolutionDateTime: m.FinalResolutionDateTime, UTCOffset: m.UTCOffset, IsResolved: m.IsResolved(), ResolutionResult: m.ResolutionResult, InitialProbability: m.InitialProbability, CreatorUsername: m.CreatorUsername, StewardUsername: m.CurrentStewardUsername(), CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, YesLabel: m.YesLabel, NoLabel: m.NoLabel, Status: m.Status, LifecycleStatus: m.LifecycleStatus, Tags: MarketTagOutputsFromDomain(m.Tags)}
}
func MarketOverviewOutputFromDomain(o *dmarkets.MarketOverview) MarketOverviewOutput {
	if o == nil {
		return MarketOverviewOutput{Market: MarketOutput{Tags: []MarketTagOutput{}}}
	}
	return MarketOverviewOutput{Market: MarketOutputFromDomain(o.Market), Creator: CreatorOutputFromDomain(o.Creator), LastProbability: o.LastProbability, NumUsers: o.NumUsers, TotalVolume: o.TotalVolume, MarketDust: o.MarketDust}
}
func MarketDetailsOutputFromDomain(o *dmarkets.MarketOverview) MarketDetailsOutput {
	if o == nil {
		return MarketDetailsOutput{ProbabilityChanges: []ProbabilityPointOutput{}}
	}
	return MarketDetailsOutput{Market: MarketOutputFromDomain(o.Market), Creator: CreatorOutputFromDomain(o.Creator), ProbabilityChanges: ProbabilityPointsFromDomain(o.ProbabilityChanges), NumUsers: o.NumUsers, TotalVolume: o.TotalVolume, MarketDust: o.MarketDust}
}
func MarketSummaryOutputFromDomain(s *dmarkets.MarketSummaryReadModel) MarketSummaryOutput {
	if s == nil {
		return MarketSummaryOutput{ProbabilityChanges: []ProbabilityPointOutput{}}
	}
	f := s.Accounting.Freshness()
	return MarketSummaryOutput{Market: MarketOutputFromDomain(s.Market), Creator: CreatorOutputFromDomain(s.Creator), ProbabilityChanges: ProbabilityPointsFromDomain(s.Accounting.ProbabilityChanges), NumUsers: s.Accounting.UserCount, TotalVolume: s.Accounting.VolumeWithDust, MarketDust: s.Accounting.MarketDust, Freshness: FreshnessOutput{GeneratedAt: f.GeneratedAt, Source: f.Source, TargetFreshnessSeconds: f.TargetFreshnessSeconds, TransactionSafeRead: f.TransactionSafeRead, IsStale: f.IsStale, StaleReason: f.StaleReason, MarkedStaleAt: f.MarkedStaleAt}}
}
func ProbabilityPointsFromDomain(points []dmarkets.ProbabilityPoint) []ProbabilityPointOutput {
	out := make([]ProbabilityPointOutput, 0, len(points))
	for _, p := range points {
		out = append(out, ProbabilityPointOutput{Probability: p.Probability, Timestamp: p.Timestamp})
	}
	return out
}

func MarketGroupOutputFromDomain(g *dmarkets.MarketGroup) *MarketGroupOutput {
	if g == nil {
		return nil
	}
	status := g.LifecycleStatus
	if strings.EqualFold(status, dmarkets.MarketLifecyclePublished) {
		status = dmarkets.MarketStatusActive
	}
	return &MarketGroupOutput{ID: g.ID, QuestionTitle: g.QuestionTitle, Description: g.Description, GroupType: g.GroupType, ProbabilityPolicy: g.ProbabilityPolicy, ResolutionPolicy: g.ResolutionPolicy, LifecycleStatus: g.LifecycleStatus, Status: status, ProposalCost: g.ProposalCost, CreatorUsername: g.CreatorUsername, StewardUsername: g.CurrentStewardUsername(), ResolutionDateTime: g.ResolutionDateTime, AutoApproveAnswerAdditions: g.AutoApproveAnswerAdditions, CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt, AnswerCount: len(g.Members)}
}
func MarketGroupLinkOutputFromDomain(g *dmarkets.MarketGroup, marketID int64) *MarketGroupLinkOutput {
	if g == nil || g.ID <= 0 {
		return nil
	}
	status := g.LifecycleStatus
	if strings.EqualFold(status, dmarkets.MarketLifecyclePublished) {
		status = dmarkets.MarketStatusActive
	}
	link := &MarketGroupLinkOutput{ID: g.ID, QuestionTitle: g.QuestionTitle, Description: g.Description, GroupType: g.GroupType, LifecycleStatus: g.LifecycleStatus, Status: status, AnswerCount: len(g.Members), ProposalCost: g.ProposalCost, CreatorUsername: g.CreatorUsername, StewardUsername: g.CurrentStewardUsername(), ResolutionDateTime: g.ResolutionDateTime, CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt}
	for _, member := range g.Members {
		if member.MarketID == marketID {
			link.AnswerLabel = member.AnswerLabel
			link.DisplayOrder = member.DisplayOrder
			break
		}
	}
	return link
}
func DiscoveryRowOutputFromDomain(row dmarkets.MarketDiscoveryRow) DiscoveryRowOutput {
	if row.Group != nil {
		children := make([]MarketOverviewOutput, 0, len(row.Children))
		for _, child := range row.Children {
			children = append(children, MarketOverviewOutput{Market: MarketOutputFromDomain(child)})
		}
		return DiscoveryRowOutput{IsMarketGroup: true, Group: MarketGroupOutputFromDomain(row.Group), ChildMarkets: children}
	}
	o := MarketOverviewOutput{Market: MarketOutputFromDomain(row.Market)}
	return DiscoveryRowOutput{Market: &o, TotalVolume: o.TotalVolume, MarketDust: o.MarketDust}
}
func BetOutputFromDomain(b *dmarkets.BetDisplayInfo) BetOutput {
	if b == nil {
		return BetOutput{}
	}
	return BetOutput{Username: b.Username, Outcome: b.Outcome, Amount: b.Amount, Probability: b.Probability, PlacedAt: b.PlacedAt}
}
func UserPositionOutputFromDomain(p *dmarkets.UserPosition) UserPositionOutput {
	if p == nil {
		return UserPositionOutput{}
	}
	return UserPositionOutput{Username: p.Username, MarketID: p.MarketID, YesSharesOwned: p.YesSharesOwned, NoSharesOwned: p.NoSharesOwned, Value: p.Value, TotalSpent: p.TotalSpent, TotalSpentInPlay: p.TotalSpentInPlay, IsResolved: p.IsResolved, ResolutionResult: p.ResolutionResult}
}
func LeaderboardRowOutputFromDomain(r *dmarkets.LeaderboardRow) LeaderboardRowOutput {
	if r == nil {
		return LeaderboardRowOutput{}
	}
	return LeaderboardRowOutput{Username: r.Username, Profit: r.Profit, CurrentValue: r.CurrentValue, TotalSpent: r.TotalSpent, Position: r.Position, YesSharesOwned: r.YesSharesOwned, NoSharesOwned: r.NoSharesOwned, Rank: r.Rank}
}

func MarketGroupBetOutputFromDomain(row *dmarkets.MarketGroupBetDisplayInfo) MarketGroupBetOutput {
	if row == nil {
		return MarketGroupBetOutput{}
	}
	return MarketGroupBetOutput{AnswerMarketID: row.AnswerMarketID, AnswerLabel: row.AnswerLabel, DisplayOrder: row.DisplayOrder, Username: row.Username, Outcome: row.Outcome, Amount: row.Amount, Probability: row.Probability, PlacedAt: row.PlacedAt}
}

func MarketGroupPositionOutputFromDomain(row *dmarkets.MarketGroupPositionRow) MarketGroupPositionOutput {
	if row == nil {
		return MarketGroupPositionOutput{Answers: []MarketGroupPositionAnswerOutput{}}
	}
	answers := make([]MarketGroupPositionAnswerOutput, 0, len(row.Answers))
	for _, answer := range row.Answers {
		if answer == nil {
			continue
		}
		answers = append(answers, MarketGroupPositionAnswerOutput{AnswerMarketID: answer.AnswerMarketID, AnswerLabel: answer.AnswerLabel, DisplayOrder: answer.DisplayOrder, MarketID: answer.MarketID, YesSharesOwned: answer.YesSharesOwned, NoSharesOwned: answer.NoSharesOwned, Value: answer.Value, TotalSpent: answer.TotalSpent, TotalSpentInPlay: answer.TotalSpentInPlay, IsResolved: answer.IsResolved, ResolutionResult: answer.ResolutionResult})
	}
	return MarketGroupPositionOutput{Username: row.Username, YesSharesOwned: row.YesSharesOwned, NoSharesOwned: row.NoSharesOwned, Value: row.Value, TotalSpent: row.TotalSpent, TotalSpentInPlay: row.TotalSpentInPlay, Answers: answers}
}

func MarketGroupLeaderboardRowOutputFromDomain(row *dmarkets.MarketGroupLeaderboardRow) MarketGroupLeaderboardRowOutput {
	if row == nil {
		return MarketGroupLeaderboardRowOutput{Answers: []MarketGroupLeaderboardAnswerOutput{}}
	}
	answers := make([]MarketGroupLeaderboardAnswerOutput, 0, len(row.Answers))
	for _, answer := range row.Answers {
		if answer == nil {
			continue
		}
		answers = append(answers, MarketGroupLeaderboardAnswerOutput{AnswerMarketID: answer.AnswerMarketID, AnswerLabel: answer.AnswerLabel, DisplayOrder: answer.DisplayOrder, Profit: answer.Profit, CurrentValue: answer.CurrentValue, TotalSpent: answer.TotalSpent, Position: answer.Position, YesSharesOwned: answer.YesSharesOwned, NoSharesOwned: answer.NoSharesOwned})
	}
	return MarketGroupLeaderboardRowOutput{Username: row.Username, Profit: row.Profit, CurrentValue: row.CurrentValue, TotalSpent: row.TotalSpent, Position: row.Position, YesSharesOwned: row.YesSharesOwned, NoSharesOwned: row.NoSharesOwned, Rank: row.Rank, Answers: answers}
}
