package markets

import (
	"context"
	"time"

	users "socialpredict/internal/domain/users"
)

const (
	MaxQuestionTitleLength = 160
	MaxDescriptionLength   = 2000
	MaxLabelLength         = 20
	MinLabelLength         = 1
)

// Clock provides time functionality for testability.
type Clock interface {
	Now() time.Time
}

// Repository defines the interface for market data access.
type Repository interface {
	Create(ctx context.Context, market *Market) error
	GetByID(ctx context.Context, id int64) (*Market, error)
	UpdateLabels(ctx context.Context, id int64, yesLabel, noLabel string) error
	List(ctx context.Context, filters ListFilters) ([]*Market, error)
	ListByStatus(ctx context.Context, status string, p Page) ([]*Market, error)
	Search(ctx context.Context, query string, filters SearchFilters) ([]*Market, error)
	Delete(ctx context.Context, id int64) error
	ResolveMarket(ctx context.Context, id int64, resolution string) error
	GetUserPosition(ctx context.Context, marketID int64, username string) (*UserPosition, error)
	ListMarketPositions(ctx context.Context, marketID int64) (MarketPositions, error)
	ListBetsForMarket(ctx context.Context, marketID int64) ([]*Bet, error)
	CalculatePayoutPositions(ctx context.Context, marketID int64) ([]*PayoutPosition, error)
	GetPublicMarket(ctx context.Context, marketID int64) (*PublicMarket, error)
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

// UserService defines the interface for user-related operations.
type UserService interface {
	ValidateUserExists(ctx context.Context, username string) error
	ValidateUserBalance(ctx context.Context, username string, requiredAmount int64, maxDebt int64) error
	DeductBalance(ctx context.Context, username string, amount int64) error
	ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error
	GetPublicUser(ctx context.Context, username string) (*users.PublicUser, error)
}

// Config holds configuration for the markets service.
type Config struct {
	MinimumFutureHours float64
	CreateMarketCost   int64
	MaximumDebtAllowed int64
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

// ServiceInterface defines the interface for market service operations.
type ServiceInterface interface {
	CreateMarket(ctx context.Context, req MarketCreateRequest, creatorUsername string) (*Market, error)
	SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error
	GetMarket(ctx context.Context, id int64) (*Market, error)
	ListMarkets(ctx context.Context, filters ListFilters) ([]*Market, error)
	SearchMarkets(ctx context.Context, query string, filters SearchFilters) (*SearchResults, error)
	ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error
	ListByStatus(ctx context.Context, status string, p Page) ([]*Market, error)
	GetMarketLeaderboard(ctx context.Context, marketID int64, p Page) ([]*LeaderboardRow, error)
	ProjectProbability(ctx context.Context, req ProbabilityProjectionRequest) (*ProbabilityProjection, error)
	GetMarketDetails(ctx context.Context, marketID int64) (*MarketOverview, error)
	GetMarketBets(ctx context.Context, marketID int64) ([]*BetDisplayInfo, error)
	GetMarketPositions(ctx context.Context, marketID int64) (MarketPositions, error)
	GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*UserPosition, error)
	CalculateMarketVolume(ctx context.Context, marketID int64) (int64, error)
	GetPublicMarket(ctx context.Context, marketID int64) (*PublicMarket, error)
}

// Service implements the core market business logic.
type Service struct {
	repo        Repository
	userService UserService
	clock       Clock
	config      Config
}

// NewService creates a new markets service.
func NewService(repo Repository, userService UserService, clock Clock, config Config) *Service {
	return &Service{
		repo:        repo,
		userService: userService,
		clock:       clock,
		config:      config,
	}
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

// Page represents pagination parameters.
type Page struct {
	Limit  int
	Offset int
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
