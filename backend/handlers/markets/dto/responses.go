package dto

import (
	"time"
)

// MarketResponse represents the HTTP response for a market
type MarketResponse struct {
	ID                 int64     `json:"id"`
	QuestionTitle      string    `json:"questionTitle"`
	Description        string    `json:"description"`
	OutcomeType        string    `json:"outcomeType"`
	ResolutionDateTime time.Time `json:"resolutionDateTime"`
	CreatorUsername    string    `json:"creatorUsername"`
	YesLabel           string    `json:"yesLabel"`
	NoLabel            string    `json:"noLabel"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// CreateMarketResponse represents the HTTP response after creating a market
type CreateMarketResponse struct {
	ID                 int64     `json:"id"`
	QuestionTitle      string    `json:"questionTitle"`
	Description        string    `json:"description"`
	OutcomeType        string    `json:"outcomeType"`
	ResolutionDateTime time.Time `json:"resolutionDateTime"`
	CreatorUsername    string    `json:"creatorUsername"`
	YesLabel           string    `json:"yesLabel"`
	NoLabel            string    `json:"noLabel"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
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
}

// SimpleListMarketsResponse represents the HTTP response for simple market listing
type SimpleListMarketsResponse struct {
	Markets []*MarketResponse `json:"markets"`
	Total   int               `json:"total"`
}

// ListMarketsResponse represents the HTTP response for listing markets with enriched data
type ListMarketsResponse struct {
	Markets []*MarketOverviewResponse `json:"markets"`
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
	Username string  `json:"username"`
	Profit   float64 `json:"profit"`
	Volume   int64   `json:"volume"`
	Rank     int     `json:"rank"`
}

// LeaderboardResponse represents the HTTP response for market leaderboard
type LeaderboardResponse struct {
	MarketID    int64            `json:"marketId"`
	Leaderboard []LeaderboardRow `json:"leaderboard"`
	Total       int              `json:"total"`
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
	MarketID           int64       `json:"marketId"`
	Market             interface{} `json:"market"`             // Will be properly typed later
	Creator            interface{} `json:"creator"`            // Will be properly typed later
	ProbabilityChanges interface{} `json:"probabilityChanges"` // Will be properly typed later
	NumUsers           int         `json:"numUsers"`
	TotalVolume        int64       `json:"totalVolume"`
	MarketDust         int64       `json:"marketDust"`
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
	PrimaryResults  []MarketResponse `json:"primaryResults"`
	FallbackResults []MarketResponse `json:"fallbackResults"`
	Query           string           `json:"query"`
	PrimaryStatus   string           `json:"primaryStatus"`
	PrimaryCount    int              `json:"primaryCount"`
	FallbackCount   int              `json:"fallbackCount"`
	TotalCount      int              `json:"totalCount"`
	FallbackUsed    bool             `json:"fallbackUsed"`
}
