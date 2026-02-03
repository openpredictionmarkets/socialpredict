package markets

import "time"

// Page represents pagination parameters.
type Page struct {
	Limit  int
	Offset int
}

// ListFilters represents filters for listing markets.
type ListFilters struct {
	Status    string
	CreatedBy string
	Limit     int
	Offset    int
}

// SearchFilters represents filters for searching markets.
type SearchFilters struct {
	Status string
	Limit  int
	Offset int
}

// SearchResults represents the result of a market search with fallback.
type SearchResults struct {
	PrimaryResults  []*Market `json:"primaryResults"`
	FallbackResults []*Market `json:"fallbackResults"`
	Query           string    `json:"query"`
	PrimaryStatus   string    `json:"primaryStatus"`
	PrimaryCount    int       `json:"primaryCount"`
	FallbackCount   int       `json:"fallbackCount"`
	TotalCount      int       `json:"totalCount"`
	FallbackUsed    bool      `json:"fallbackUsed"`
}

// CreatorSummary captures lightweight information about a market creator.
type CreatorSummary struct {
	Username      string
	DisplayName   string
	PersonalEmoji string
}

// ProbabilityPoint records a market probability at a specific moment.
type ProbabilityPoint struct {
	Probability float64
	Timestamp   time.Time
}

// MarketOverview represents enriched market data with calculations.
type MarketOverview struct {
	Market             *Market
	Creator            *CreatorSummary
	ProbabilityChanges []ProbabilityPoint
	LastProbability    float64
	NumUsers           int
	TotalVolume        int64
	MarketDust         int64
}

// LeaderboardRow represents a single row in the market leaderboard.
type LeaderboardRow struct {
	Username       string
	Profit         int64
	CurrentValue   int64
	TotalSpent     int64
	Position       string
	YesSharesOwned int64
	NoSharesOwned  int64
	Rank           int
}

// ProbabilityProjectionRequest represents a request for probability projection.
type ProbabilityProjectionRequest struct {
	MarketID int64
	Amount   int64
	Outcome  string
}

// ProbabilityProjection represents the result of a probability projection.
type ProbabilityProjection struct {
	CurrentProbability   float64
	ProjectedProbability float64
}

// BetDisplayInfo represents a bet with probability information.
type BetDisplayInfo struct {
	Username    string    `json:"username"`
	Outcome     string    `json:"outcome"`
	Amount      int64     `json:"amount"`
	Probability float64   `json:"probability"`
	PlacedAt    time.Time `json:"placedAt"`
}
