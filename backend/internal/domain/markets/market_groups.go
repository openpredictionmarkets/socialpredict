package markets

import (
	"context"
	"sort"
	"strings"
	"time"
)

const (
	MarketGroupTypeMultipleChoiceBinary = "MULTIPLE_CHOICE_BINARY"

	MarketGroupProbabilityPolicyIndependentBinary = "INDEPENDENT_BINARY"

	MarketGroupResolutionPolicyIndependentChildren = "INDEPENDENT_CHILDREN"
	MarketGroupResolutionPolicyExclusiveHelper     = "EXCLUSIVE_HELPER"

	MinMarketGroupAnswers = 2
	MaxMarketGroupAnswers = 20
	MaxAnswerLabelLength  = 160
)

// MarketGroup is a display and governance parent for normal binary child markets.
// It does not own trading math; child markets remain the transaction boundary.
type MarketGroup struct {
	ID                 int64
	QuestionTitle      string
	Description        string
	GroupType          string
	ProbabilityPolicy  string
	ResolutionPolicy   string
	LifecycleStatus    string
	ProposalCost       int64
	CreatorUsername    string
	StewardUsername    string
	ApprovedBy         string
	ApprovedAt         *time.Time
	RejectedBy         string
	RejectedAt         *time.Time
	RejectionReason    string
	ResolutionDateTime time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Members            []MarketGroupMember
}

// MarketGroupCreateRequest captures a participant-created grouped binary
// market. Each answer label becomes a normal binary child market.
type MarketGroupCreateRequest struct {
	QuestionTitle      string
	Description        string
	ResolutionDateTime time.Time
	AnswerLabels       []string
	TagSlugs           []string
}

// MarketGroupMember links one normal binary child market to a parent group.
type MarketGroupMember struct {
	ID           int64
	GroupID      int64
	MarketID     int64
	AnswerLabel  string
	DisplayOrder int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// MarketGroupRepository persists parent groups and child-market links. It is
// intentionally separate from the existing binary market Repository so current
// transaction paths cannot accidentally depend on grouped display state.
type MarketGroupRepository interface {
	CreateMarketGroup(ctx context.Context, group *MarketGroup, members []MarketGroupMember) error
	GetMarketGroup(ctx context.Context, groupID int64) (*MarketGroup, error)
	ListMarketGroupMembers(ctx context.Context, groupID int64) ([]MarketGroupMember, error)
}

// MarketGroupLookupRepository resolves a child market back to its parent group.
// Display/read paths use this to bind normal binary child markets into a group
// without changing child-market transaction ownership.
type MarketGroupLookupRepository interface {
	GetMarketGroupForMarket(ctx context.Context, marketID int64) (*MarketGroup, error)
}

// MarketGroupOverview is the parent group read shape with child market
// overviews. It is display data, not transaction state.
type MarketGroupOverview struct {
	Group   *MarketGroup
	Creator *CreatorSummary
	Answers []MarketGroupAnswerOverview
}

type MarketGroupAnswerOverview struct {
	Member   MarketGroupMember
	Overview *MarketOverview
}

// NormalizeMarketGroupDefaults fills policy defaults while preserving explicit
// values for future market group classes.
func NormalizeMarketGroupDefaults(group *MarketGroup) {
	if group == nil {
		return
	}
	if strings.TrimSpace(group.GroupType) == "" {
		group.GroupType = MarketGroupTypeMultipleChoiceBinary
	}
	if strings.TrimSpace(group.ProbabilityPolicy) == "" {
		group.ProbabilityPolicy = MarketGroupProbabilityPolicyIndependentBinary
	}
	if strings.TrimSpace(group.ResolutionPolicy) == "" {
		group.ResolutionPolicy = MarketGroupResolutionPolicyIndependentChildren
	}
	if strings.TrimSpace(group.LifecycleStatus) == "" {
		group.LifecycleStatus = MarketLifecycleProposed
	}
	if strings.TrimSpace(group.StewardUsername) == "" {
		group.StewardUsername = strings.TrimSpace(group.CreatorUsername)
	}
}

// ValidateMarketGroupMembers enforces baseline multiple-choice child-market
// invariants before persistence.
func ValidateMarketGroupMembers(members []MarketGroupMember) error {
	if len(members) < MinMarketGroupAnswers || len(members) > MaxMarketGroupAnswers {
		return ErrInvalidInput
	}

	seenLabels := map[string]bool{}
	seenOrders := map[int]bool{}
	seenMarkets := map[int64]bool{}
	for _, member := range members {
		label := strings.TrimSpace(member.AnswerLabel)
		normalizedLabel := strings.ToLower(label)
		if label == "" || len(label) > MaxAnswerLabelLength {
			return ErrInvalidInput
		}
		if seenLabels[normalizedLabel] {
			return ErrInvalidInput
		}
		if seenOrders[member.DisplayOrder] {
			return ErrInvalidInput
		}
		if member.MarketID <= 0 || seenMarkets[member.MarketID] {
			return ErrInvalidInput
		}
		seenLabels[normalizedLabel] = true
		seenOrders[member.DisplayOrder] = true
		seenMarkets[member.MarketID] = true
	}
	return nil
}

// OrderedMarketGroupMembers returns a stable display order copy.
func OrderedMarketGroupMembers(members []MarketGroupMember) []MarketGroupMember {
	ordered := append([]MarketGroupMember(nil), members...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].DisplayOrder == ordered[j].DisplayOrder {
			return ordered[i].ID < ordered[j].ID
		}
		return ordered[i].DisplayOrder < ordered[j].DisplayOrder
	})
	return ordered
}
