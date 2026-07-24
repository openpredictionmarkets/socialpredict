package mcpserver

import "time"

type EmptyInput struct{}
type SlugInput struct {
	Slug string `json:"slug" jsonschema:"market discovery or tag slug"`
}
type MarketListInput struct {
	Status    string `json:"status,omitempty" jsonschema:"active, open, closed, resolved, or all"`
	TagSlug   string `json:"tagSlug,omitempty" jsonschema:"market tag slug"`
	CreatedBy string `json:"createdBy,omitempty" jsonschema:"creator username"`
	Limit     int    `json:"limit,omitempty" jsonschema:"page size, max 100"`
	Offset    int    `json:"offset,omitempty" jsonschema:"zero-based page offset"`
}
type MarketSearchInput struct {
	Query   string `json:"query" jsonschema:"search query"`
	Status  string `json:"status,omitempty" jsonschema:"active, open, closed, resolved, or all"`
	TagSlug string `json:"tagSlug,omitempty" jsonschema:"market tag slug"`
	Limit   int    `json:"limit,omitempty" jsonschema:"page size, max 100"`
	Offset  int    `json:"offset,omitempty" jsonschema:"zero-based page offset"`
}
type MarketDiscoveryInput struct {
	Slug    string `json:"slug" jsonschema:"market discovery page slug; use markets for the top page"`
	Status  string `json:"status,omitempty" jsonschema:"active, open, closed, resolved, or all"`
	TagSlug string `json:"tagSlug,omitempty" jsonschema:"market tag slug"`
	Limit   int    `json:"limit,omitempty" jsonschema:"page size, max 100"`
	Offset  int    `json:"offset,omitempty" jsonschema:"zero-based page offset"`
}
type MarketIDInput struct {
	MarketID int64 `json:"marketId" jsonschema:"market id"`
}
type MarketActivityInput struct {
	MarketID int64 `json:"marketId" jsonschema:"market id"`
	Limit    int   `json:"limit,omitempty" jsonschema:"page size, max 100"`
	Offset   int   `json:"offset,omitempty" jsonschema:"zero-based page offset"`
}
type MarketUserPositionInput struct {
	MarketID int64  `json:"marketId" jsonschema:"market id"`
	Username string `json:"username" jsonschema:"public username"`
}
type MarketGroupActivityInput struct {
	GroupID int64 `json:"groupId" jsonschema:"market group id"`
	Limit   int   `json:"limit,omitempty" jsonschema:"page size, max 100"`
	Offset  int   `json:"offset,omitempty" jsonschema:"zero-based page offset"`
}
type ProbabilityQuoteInput struct {
	MarketID int64  `json:"marketId" jsonschema:"market id"`
	Amount   int64  `json:"amount" jsonschema:"hypothetical bet amount"`
	Outcome  string `json:"outcome" jsonschema:"YES or NO"`
}

type MarketTagOutput struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	ColorKey    string `json:"colorKey,omitempty"`
	SortOrder   int    `json:"sortOrder"`
	IsActive    bool   `json:"isActive"`
}
type CreatorOutput struct {
	Username      string `json:"username"`
	DisplayName   string `json:"displayName,omitempty"`
	PersonalEmoji string `json:"personalEmoji,omitempty"`
}
type MarketGroupLinkOutput struct {
	ID                 int64     `json:"id"`
	QuestionTitle      string    `json:"questionTitle"`
	Description        string    `json:"description,omitempty"`
	GroupType          string    `json:"groupType,omitempty"`
	LifecycleStatus    string    `json:"lifecycleStatus,omitempty"`
	Status             string    `json:"status,omitempty"`
	AnswerLabel        string    `json:"answerLabel,omitempty"`
	DisplayOrder       int       `json:"displayOrder,omitempty"`
	AnswerCount        int       `json:"answerCount"`
	ProposalCost       int64     `json:"proposalCost,omitempty"`
	CreatorUsername    string    `json:"creatorUsername,omitempty"`
	StewardUsername    string    `json:"stewardUsername,omitempty"`
	ResolutionDateTime time.Time `json:"resolutionDateTime,omitempty"`
	CreatedAt          time.Time `json:"createdAt,omitempty"`
	UpdatedAt          time.Time `json:"updatedAt,omitempty"`
}
type MarketOutput struct {
	ID                      int64                  `json:"id"`
	QuestionTitle           string                 `json:"questionTitle"`
	Description             string                 `json:"description,omitempty"`
	OutcomeType             string                 `json:"outcomeType,omitempty"`
	ResolutionDateTime      time.Time              `json:"resolutionDateTime,omitempty"`
	FinalResolutionDateTime time.Time              `json:"finalResolutionDateTime,omitempty"`
	UTCOffset               int                    `json:"utcOffset"`
	IsResolved              bool                   `json:"isResolved"`
	ResolutionResult        string                 `json:"resolutionResult,omitempty"`
	InitialProbability      float64                `json:"initialProbability"`
	CreatorUsername         string                 `json:"creatorUsername,omitempty"`
	StewardUsername         string                 `json:"stewardUsername,omitempty"`
	CreatedAt               time.Time              `json:"createdAt,omitempty"`
	UpdatedAt               time.Time              `json:"updatedAt,omitempty"`
	YesLabel                string                 `json:"yesLabel,omitempty"`
	NoLabel                 string                 `json:"noLabel,omitempty"`
	Status                  string                 `json:"status,omitempty"`
	LifecycleStatus         string                 `json:"lifecycleStatus,omitempty"`
	Tags                    []MarketTagOutput      `json:"tags,omitempty"`
	MarketGroup             *MarketGroupLinkOutput `json:"marketGroup,omitempty"`
}
type MarketOverviewOutput struct {
	Market          MarketOutput   `json:"market"`
	Creator         *CreatorOutput `json:"creator,omitempty"`
	LastProbability float64        `json:"lastProbability"`
	NumUsers        int            `json:"numUsers"`
	TotalVolume     int64          `json:"totalVolume"`
	MarketDust      int64          `json:"marketDust"`
}
type ProbabilityPointOutput struct {
	Probability float64   `json:"probability"`
	Timestamp   time.Time `json:"timestamp"`
}
type FreshnessOutput struct {
	GeneratedAt            time.Time  `json:"generatedAt"`
	Source                 string     `json:"source"`
	TargetFreshnessSeconds int        `json:"targetFreshnessSeconds"`
	TransactionSafeRead    bool       `json:"transactionSafeRead"`
	IsStale                bool       `json:"isStale"`
	StaleReason            string     `json:"staleReason,omitempty"`
	MarkedStaleAt          *time.Time `json:"markedStaleAt,omitempty"`
}
type MarketDetailsOutput struct {
	Market             MarketOutput             `json:"market"`
	Creator            *CreatorOutput           `json:"creator,omitempty"`
	ProbabilityChanges []ProbabilityPointOutput `json:"probabilityChanges"`
	NumUsers           int                      `json:"numUsers"`
	TotalVolume        int64                    `json:"totalVolume"`
	MarketDust         int64                    `json:"marketDust"`
}
type MarketSummaryOutput struct {
	Market             MarketOutput             `json:"market"`
	Creator            *CreatorOutput           `json:"creator,omitempty"`
	ProbabilityChanges []ProbabilityPointOutput `json:"probabilityChanges"`
	NumUsers           int                      `json:"numUsers"`
	TotalVolume        int64                    `json:"totalVolume"`
	MarketDust         int64                    `json:"marketDust"`
	Freshness          FreshnessOutput          `json:"freshness"`
}
type MarketGroupOutput struct {
	ID                         int64     `json:"id"`
	QuestionTitle              string    `json:"questionTitle"`
	Description                string    `json:"description,omitempty"`
	GroupType                  string    `json:"groupType,omitempty"`
	ProbabilityPolicy          string    `json:"probabilityPolicy,omitempty"`
	ResolutionPolicy           string    `json:"resolutionPolicy,omitempty"`
	LifecycleStatus            string    `json:"lifecycleStatus,omitempty"`
	Status                     string    `json:"status,omitempty"`
	ProposalCost               int64     `json:"proposalCost,omitempty"`
	CreatorUsername            string    `json:"creatorUsername,omitempty"`
	StewardUsername            string    `json:"stewardUsername,omitempty"`
	ResolutionDateTime         time.Time `json:"resolutionDateTime,omitempty"`
	AutoApproveAnswerAdditions bool      `json:"autoApproveAnswerAdditions"`
	CreatedAt                  time.Time `json:"createdAt,omitempty"`
	UpdatedAt                  time.Time `json:"updatedAt,omitempty"`
	AnswerCount                int       `json:"answerCount"`
}
type DiscoveryRowOutput struct {
	IsMarketGroup bool                   `json:"isMarketGroup"`
	Market        *MarketOverviewOutput  `json:"market,omitempty"`
	Group         *MarketGroupOutput     `json:"group,omitempty"`
	ChildMarkets  []MarketOverviewOutput `json:"childMarkets,omitempty"`
	TotalVolume   int64                  `json:"totalVolume"`
	MarketDust    int64                  `json:"marketDust"`
}

type BetOutput struct {
	Username    string    `json:"username"`
	Outcome     string    `json:"outcome"`
	Amount      int64     `json:"amount"`
	Probability float64   `json:"probability"`
	PlacedAt    time.Time `json:"placedAt"`
}
type UserPositionOutput struct {
	Username         string `json:"username"`
	MarketID         int64  `json:"marketId"`
	YesSharesOwned   int64  `json:"yesSharesOwned"`
	NoSharesOwned    int64  `json:"noSharesOwned"`
	Value            int64  `json:"value"`
	TotalSpent       int64  `json:"totalSpent"`
	TotalSpentInPlay int64  `json:"totalSpentInPlay"`
	IsResolved       bool   `json:"isResolved"`
	ResolutionResult string `json:"resolutionResult,omitempty"`
}
type LeaderboardRowOutput struct {
	Username       string `json:"username"`
	Profit         int64  `json:"profit"`
	CurrentValue   int64  `json:"currentValue"`
	TotalSpent     int64  `json:"totalSpent"`
	Position       string `json:"position"`
	YesSharesOwned int64  `json:"yesSharesOwned"`
	NoSharesOwned  int64  `json:"noSharesOwned"`
	Rank           int    `json:"rank"`
}
type MarketGroupBetOutput struct {
	AnswerMarketID int64     `json:"answerMarketId"`
	AnswerLabel    string    `json:"answerLabel"`
	DisplayOrder   int       `json:"displayOrder"`
	Username       string    `json:"username"`
	Outcome        string    `json:"outcome"`
	Amount         int64     `json:"amount"`
	Probability    float64   `json:"probability"`
	PlacedAt       time.Time `json:"placedAt"`
}
type MarketGroupPositionAnswerOutput struct {
	AnswerMarketID   int64  `json:"answerMarketId"`
	AnswerLabel      string `json:"answerLabel"`
	DisplayOrder     int    `json:"displayOrder"`
	MarketID         int64  `json:"marketId"`
	YesSharesOwned   int64  `json:"yesSharesOwned"`
	NoSharesOwned    int64  `json:"noSharesOwned"`
	Value            int64  `json:"value"`
	TotalSpent       int64  `json:"totalSpent"`
	TotalSpentInPlay int64  `json:"totalSpentInPlay"`
	IsResolved       bool   `json:"isResolved"`
	ResolutionResult string `json:"resolutionResult,omitempty"`
}
type MarketGroupPositionOutput struct {
	Username         string                            `json:"username"`
	YesSharesOwned   int64                             `json:"yesSharesOwned"`
	NoSharesOwned    int64                             `json:"noSharesOwned"`
	Value            int64                             `json:"value"`
	TotalSpent       int64                             `json:"totalSpent"`
	TotalSpentInPlay int64                             `json:"totalSpentInPlay"`
	Answers          []MarketGroupPositionAnswerOutput `json:"answers"`
}
type MarketGroupLeaderboardAnswerOutput struct {
	AnswerMarketID int64  `json:"answerMarketId"`
	AnswerLabel    string `json:"answerLabel"`
	DisplayOrder   int    `json:"displayOrder"`
	Profit         int64  `json:"profit"`
	CurrentValue   int64  `json:"currentValue"`
	TotalSpent     int64  `json:"totalSpent"`
	Position       string `json:"position"`
	YesSharesOwned int64  `json:"yesSharesOwned"`
	NoSharesOwned  int64  `json:"noSharesOwned"`
}
type MarketGroupLeaderboardRowOutput struct {
	Username       string                               `json:"username"`
	Profit         int64                                `json:"profit"`
	CurrentValue   int64                                `json:"currentValue"`
	TotalSpent     int64                                `json:"totalSpent"`
	Position       string                               `json:"position"`
	YesSharesOwned int64                                `json:"yesSharesOwned"`
	NoSharesOwned  int64                                `json:"noSharesOwned"`
	Rank           int                                  `json:"rank"`
	Answers        []MarketGroupLeaderboardAnswerOutput `json:"answers"`
}
type ProbabilityQuoteOutput struct {
	MarketID             int64   `json:"marketId"`
	CurrentProbability   float64 `json:"currentProbability"`
	ProjectedProbability float64 `json:"projectedProbability"`
	Amount               int64   `json:"amount"`
	Outcome              string  `json:"outcome"`
}
