package markets_test

import (
	"errors"
	"testing"

	"socialpredict/internal/domain/markets"
)

func TestNormalizeMarketGroupDefaults(t *testing.T) {
	group := &markets.MarketGroup{
		CreatorUsername: "moderator",
	}

	markets.NormalizeMarketGroupDefaults(group)

	if group.GroupType != markets.MarketGroupTypeMultipleChoiceBinary {
		t.Fatalf("GroupType = %q", group.GroupType)
	}
	if group.ProbabilityPolicy != markets.MarketGroupProbabilityPolicyIndependentBinary {
		t.Fatalf("ProbabilityPolicy = %q", group.ProbabilityPolicy)
	}
	if group.ResolutionPolicy != markets.MarketGroupResolutionPolicyIndependentChildren {
		t.Fatalf("ResolutionPolicy = %q", group.ResolutionPolicy)
	}
	if group.LifecycleStatus != markets.MarketLifecycleProposed {
		t.Fatalf("LifecycleStatus = %q", group.LifecycleStatus)
	}
	if group.StewardUsername != "moderator" {
		t.Fatalf("StewardUsername = %q", group.StewardUsername)
	}
}

func TestValidateMarketGroupMembers(t *testing.T) {
	tests := []struct {
		name    string
		members []markets.MarketGroupMember
		wantErr error
	}{
		{
			name: "valid",
			members: []markets.MarketGroupMember{
				{MarketID: 10, AnswerLabel: "Team A", DisplayOrder: 0},
				{MarketID: 11, AnswerLabel: "Team B", DisplayOrder: 1},
			},
		},
		{
			name: "too few answers",
			members: []markets.MarketGroupMember{
				{MarketID: 10, AnswerLabel: "Team A", DisplayOrder: 0},
			},
			wantErr: markets.ErrInvalidInput,
		},
		{
			name: "duplicate labels normalize case",
			members: []markets.MarketGroupMember{
				{MarketID: 10, AnswerLabel: "Team A", DisplayOrder: 0},
				{MarketID: 11, AnswerLabel: " team a ", DisplayOrder: 1},
			},
			wantErr: markets.ErrInvalidInput,
		},
		{
			name: "duplicate child market",
			members: []markets.MarketGroupMember{
				{MarketID: 10, AnswerLabel: "Team A", DisplayOrder: 0},
				{MarketID: 10, AnswerLabel: "Team B", DisplayOrder: 1},
			},
			wantErr: markets.ErrInvalidInput,
		},
		{
			name: "duplicate display order",
			members: []markets.MarketGroupMember{
				{MarketID: 10, AnswerLabel: "Team A", DisplayOrder: 0},
				{MarketID: 11, AnswerLabel: "Team B", DisplayOrder: 0},
			},
			wantErr: markets.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := markets.ValidateMarketGroupMembers(tt.members)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateMarketGroupMembers error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrderedMarketGroupMembers(t *testing.T) {
	members := []markets.MarketGroupMember{
		{ID: 3, MarketID: 30, AnswerLabel: "C", DisplayOrder: 2},
		{ID: 1, MarketID: 10, AnswerLabel: "A", DisplayOrder: 0},
		{ID: 2, MarketID: 20, AnswerLabel: "B", DisplayOrder: 1},
	}

	ordered := markets.OrderedMarketGroupMembers(members)

	if ordered[0].AnswerLabel != "A" || ordered[1].AnswerLabel != "B" || ordered[2].AnswerLabel != "C" {
		t.Fatalf("unexpected order: %+v", ordered)
	}
	if members[0].AnswerLabel != "C" {
		t.Fatalf("OrderedMarketGroupMembers mutated input slice")
	}
}
