package dto

import "time"

// MarketResponse represents the HTTP response for a market
type MarketResponse struct {
	ID                 int64               `json:"id"`
	QuestionTitle      string              `json:"questionTitle"`
	Description        string              `json:"description"`
	OutcomeType        string              `json:"outcomeType"`
	ResolutionDateTime time.Time           `json:"resolutionDateTime"`
	CreatorUsername    string              `json:"creatorUsername"`
	StewardUsername    string              `json:"stewardUsername"`
	YesLabel           string              `json:"yesLabel"`
	NoLabel            string              `json:"noLabel"`
	Status             string              `json:"status"`
	LifecycleStatus    string              `json:"lifecycleStatus,omitempty"`
	ApprovedBy         string              `json:"approvedBy,omitempty"`
	ApprovedAt         *time.Time          `json:"approvedAt,omitempty"`
	RejectedBy         string              `json:"rejectedBy,omitempty"`
	RejectedAt         *time.Time          `json:"rejectedAt,omitempty"`
	RejectionReason    string              `json:"rejectionReason,omitempty"`
	ProposalCost       int64               `json:"proposalCost,omitempty"`
	IsResolved         bool                `json:"isResolved"`
	ResolutionResult   string              `json:"resolutionResult"`
	CreatedAt          time.Time           `json:"createdAt"`
	UpdatedAt          time.Time           `json:"updatedAt"`
	Tags               []MarketTagResponse `json:"tags,omitempty"`
	MarketGroup        *MarketGroupLink    `json:"marketGroup,omitempty"`
}

// CreateMarketResponse represents the HTTP response after creating a market
type CreateMarketResponse struct {
	ID                 int64               `json:"id"`
	QuestionTitle      string              `json:"questionTitle"`
	Description        string              `json:"description"`
	OutcomeType        string              `json:"outcomeType"`
	ResolutionDateTime time.Time           `json:"resolutionDateTime"`
	CreatorUsername    string              `json:"creatorUsername"`
	StewardUsername    string              `json:"stewardUsername"`
	YesLabel           string              `json:"yesLabel"`
	NoLabel            string              `json:"noLabel"`
	Status             string              `json:"status"`
	LifecycleStatus    string              `json:"lifecycleStatus,omitempty"`
	ProposalCost       int64               `json:"proposalCost,omitempty"`
	CreatedAt          time.Time           `json:"createdAt"`
	Tags               []MarketTagResponse `json:"tags,omitempty"`
}

// MarketGroupResponse represents a multiple-choice binary parent market.
type MarketGroupResponse struct {
	ID                 int64      `json:"id"`
	QuestionTitle      string     `json:"questionTitle"`
	Description        string     `json:"description"`
	GroupType          string     `json:"groupType"`
	ProbabilityPolicy  string     `json:"probabilityPolicy"`
	ResolutionPolicy   string     `json:"resolutionPolicy"`
	LifecycleStatus    string     `json:"lifecycleStatus"`
	Status             string     `json:"status"`
	ProposalCost       int64      `json:"proposalCost"`
	CreatorUsername    string     `json:"creatorUsername"`
	StewardUsername    string     `json:"stewardUsername"`
	ApprovedBy         string     `json:"approvedBy,omitempty"`
	ApprovedAt         *time.Time `json:"approvedAt,omitempty"`
	RejectedBy         string     `json:"rejectedBy,omitempty"`
	RejectedAt         *time.Time `json:"rejectedAt,omitempty"`
	RejectionReason    string     `json:"rejectionReason,omitempty"`
	ResolutionDateTime time.Time  `json:"resolutionDateTime"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	AnswerCount        int        `json:"answerCount"`
}

// MarketGroupAnswerResponse links one answer option to its tradable child market.
type MarketGroupAnswerResponse struct {
	ID                    int64                                `json:"id"`
	GroupID               int64                                `json:"groupId"`
	MarketID              int64                                `json:"marketId"`
	AnswerLabel           string                               `json:"answerLabel"`
	DisplayOrder          int                                  `json:"displayOrder"`
	Market                *MarketOverviewResponse              `json:"market,omitempty"`
	ProbabilityChanges    []ProbabilityChangeResponse          `json:"probabilityChanges,omitempty"`
	DescriptionAmendments []MarketDescriptionAmendmentResponse `json:"descriptionAmendments,omitempty"`
}

// MarketGroupDetailsResponse returns group metadata with child market overviews.
type MarketGroupDetailsResponse struct {
	Group   *MarketGroupResponse        `json:"group"`
	Creator *CreatorResponse            `json:"creator"`
	Answers []MarketGroupAnswerResponse `json:"answers"`
}

type MarketTagResponse struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	ColorKey    string `json:"colorKey,omitempty"`
	SortOrder   int    `json:"sortOrder"`
	IsActive    bool   `json:"isActive"`
}

type MarketTagsResponse struct {
	Tags  []MarketTagResponse `json:"tags"`
	Total int                 `json:"total"`
}

// CreatorResponse represents the creator information for frontend display
type CreatorResponse struct {
	Username      string `json:"username"`
	PersonalEmoji string `json:"personalEmoji"`
	DisplayName   string `json:"displayname,omitempty"`
}

// MarketOverviewResponse represents enriched market data for list display
type MarketOverviewResponse struct {
	Market          *MarketResponse  `json:"market"`
	Creator         *CreatorResponse `json:"creator"` // Properly typed creator info
	LastProbability float64          `json:"lastProbability"`
	NumUsers        int              `json:"numUsers"`
	TotalVolume     int64            `json:"totalVolume"`
	MarketDust      int64            `json:"marketDust"`
}

// PublicMarketResponse represents the legacy public market payload.
type PublicMarketResponse struct {
	ID                      int64               `json:"id"`
	QuestionTitle           string              `json:"questionTitle"`
	Description             string              `json:"description"`
	OutcomeType             string              `json:"outcomeType"`
	ResolutionDateTime      time.Time           `json:"resolutionDateTime"`
	FinalResolutionDateTime time.Time           `json:"finalResolutionDateTime"`
	UTCOffset               int                 `json:"utcOffset"`
	IsResolved              bool                `json:"isResolved"`
	ResolutionResult        string              `json:"resolutionResult"`
	InitialProbability      float64             `json:"initialProbability"`
	CreatorUsername         string              `json:"creatorUsername"`
	StewardUsername         string              `json:"stewardUsername"`
	CreatedAt               time.Time           `json:"createdAt"`
	YesLabel                string              `json:"yesLabel"`
	NoLabel                 string              `json:"noLabel"`
	Tags                    []MarketTagResponse `json:"tags,omitempty"`
	MarketGroup             *MarketGroupLink    `json:"marketGroup,omitempty"`
}

// MarketGroupLink binds a normal child market back to its parent group.
type MarketGroupLink struct {
	ID                 int64      `json:"id"`
	QuestionTitle      string     `json:"questionTitle"`
	Description        string     `json:"description,omitempty"`
	GroupType          string     `json:"groupType"`
	LifecycleStatus    string     `json:"lifecycleStatus"`
	Status             string     `json:"status"`
	AnswerLabel        string     `json:"answerLabel,omitempty"`
	DisplayOrder       int        `json:"displayOrder,omitempty"`
	AnswerCount        int        `json:"answerCount"`
	ProposalCost       int64      `json:"proposalCost,omitempty"`
	CreatorUsername    string     `json:"creatorUsername,omitempty"`
	StewardUsername    string     `json:"stewardUsername,omitempty"`
	ApprovedBy         string     `json:"approvedBy,omitempty"`
	ApprovedAt         *time.Time `json:"approvedAt,omitempty"`
	RejectedBy         string     `json:"rejectedBy,omitempty"`
	RejectedAt         *time.Time `json:"rejectedAt,omitempty"`
	RejectionReason    string     `json:"rejectionReason,omitempty"`
	ResolutionDateTime time.Time  `json:"resolutionDateTime"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// ProbabilityChangeResponse represents WPAM probability history.
type ProbabilityChangeResponse struct {
	Probability float64   `json:"probability"`
	Timestamp   time.Time `json:"timestamp"`
}

// SimpleListMarketsResponse represents the HTTP response for simple market listing
type SimpleListMarketsResponse struct {
	Markets []*MarketResponse `json:"markets"`
	Total   int               `json:"total"`
}

// ListMarketsResponse represents the HTTP response for listing markets with enriched data
type ListMarketsResponse struct {
	Markets []*MarketOverviewResponse `json:"markets"`
	Total   int                       `json:"total"`
}

// MarketOverview represents backward compatibility type for market overview data
type MarketOverview struct {
	Market          interface{} `json:"market"`
	Creator         interface{} `json:"creator"`
	LastProbability float64     `json:"lastProbability"`
	NumUsers        int         `json:"numUsers"`
	TotalVolume     int64       `json:"totalVolume"`
	MarketDust      int64       `json:"marketDust"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// ResolveMarketResponse represents the HTTP response after resolving a market
type ResolveMarketResponse struct {
	Message string `json:"message"`
}

// LeaderboardRow represents a single row in the market leaderboard
type LeaderboardRow struct {
	Username       string `json:"username"`
	Profit         int64  `json:"profit"`
	CurrentValue   int64  `json:"currentValue"`
	TotalSpent     int64  `json:"totalSpent"`
	Position       string `json:"position"`
	YesSharesOwned int64  `json:"yesSharesOwned"`
	NoSharesOwned  int64  `json:"noSharesOwned"`
	Rank           int    `json:"rank"`
}

// LeaderboardResponse represents the HTTP response for market leaderboard
type LeaderboardResponse struct {
	MarketID    int64            `json:"marketId"`
	Leaderboard []LeaderboardRow `json:"leaderboard"`
	Total       int              `json:"total"`
	Freshness   *Freshness       `json:"freshness,omitempty"`
}

// Freshness represents display/read-model freshness metadata.
type Freshness struct {
	GeneratedAt            time.Time  `json:"generatedAt"`
	Source                 string     `json:"source"`
	TargetFreshnessSeconds int        `json:"targetFreshnessSeconds"`
	TransactionSafeRead    bool       `json:"transactionSafeRead"`
	IsStale                bool       `json:"isStale"`
	StaleReason            string     `json:"staleReason,omitempty"`
	MarkedStaleAt          *time.Time `json:"markedStaleAt,omitempty"`
}

// ProbabilityProjectionResponse represents the HTTP response for probability projection
type ProbabilityProjectionResponse struct {
	MarketID             int64   `json:"marketId"`
	CurrentProbability   float64 `json:"currentProbability"`
	ProjectedProbability float64 `json:"projectedProbability"`
	Amount               int64   `json:"amount"`
	Outcome              string  `json:"outcome"`
}

// MarketDetailsResponse represents the HTTP response for market details
type MarketDetailsResponse struct {
	Market                PublicMarketResponse                 `json:"market"`
	Creator               *CreatorResponse                     `json:"creator"`
	ProbabilityChanges    []ProbabilityChangeResponse          `json:"probabilityChanges"`
	NumUsers              int                                  `json:"numUsers"`
	TotalVolume           int64                                `json:"totalVolume"`
	MarketDust            int64                                `json:"marketDust"`
	DescriptionAmendments []MarketDescriptionAmendmentResponse `json:"descriptionAmendments"`
}

type MarketDescriptionAmendmentResponse struct {
	ID                         int64                                `json:"id"`
	MarketID                   int64                                `json:"marketId"`
	MarketTitle                string                               `json:"marketTitle,omitempty"`
	MarketDescription          string                               `json:"marketDescription,omitempty"`
	Version                    int                                  `json:"version"`
	Body                       string                               `json:"body"`
	BodyFormat                 string                               `json:"bodyFormat"`
	Status                     string                               `json:"status"`
	CreatedBy                  string                               `json:"createdBy"`
	CreatedAt                  time.Time                            `json:"createdAt"`
	UpdatedAt                  time.Time                            `json:"updatedAt"`
	ApprovedBy                 string                               `json:"approvedBy,omitempty"`
	ApprovedAt                 *time.Time                           `json:"approvedAt,omitempty"`
	RejectedBy                 string                               `json:"rejectedBy,omitempty"`
	RejectedAt                 *time.Time                           `json:"rejectedAt,omitempty"`
	RejectionReason            string                               `json:"rejectionReason,omitempty"`
	SubmitReason               string                               `json:"submitReason,omitempty"`
	PreviousApprovedAmendments []MarketDescriptionAmendmentResponse `json:"previousApprovedAmendments,omitempty"`
}

// MarketDetailHandlerResponse - backward compatibility type for tests
type MarketDetailHandlerResponse struct {
	Market             interface{} `json:"market"`
	Creator            interface{} `json:"creator"`
	ProbabilityChanges interface{} `json:"probabilityChanges"`
	NumUsers           int         `json:"numUsers"`
	TotalVolume        int64       `json:"totalVolume"`
	MarketDust         int64       `json:"marketDust"`
}

// SearchResponse represents the HTTP response for market search with fallback logic
type SearchResponse struct {
	PrimaryResults  []*MarketOverviewResponse `json:"primaryResults"`
	FallbackResults []*MarketOverviewResponse `json:"fallbackResults"`
	Query           string                    `json:"query"`
	PrimaryStatus   string                    `json:"primaryStatus"`
	PrimaryCount    int                       `json:"primaryCount"`
	FallbackCount   int                       `json:"fallbackCount"`
	TotalCount      int                       `json:"totalCount"`
	FallbackUsed    bool                      `json:"fallbackUsed"`
}
