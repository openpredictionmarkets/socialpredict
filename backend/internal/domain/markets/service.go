package markets

import (
	"context"
	"time"

	positionsmath "socialpredict/internal/domain/math/positions"
	users "socialpredict/internal/domain/users"
	"socialpredict/models"
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

type serviceClock struct{}

func (serviceClock) Now() time.Time { return time.Now() }

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

// CreationPolicy governs validation and construction of markets.
type CreationPolicy interface {
	ValidateCreateRequest(req MarketCreateRequest) error
	ValidateCustomLabels(yesLabel, noLabel string) error
	NormalizeLabels(yesLabel, noLabel string) labelPair
	ValidateResolutionTime(now time.Time, resolution time.Time, minimumFutureHours float64) error
	EnsureCreateMarketBalance(ctx context.Context, users UserService, creatorUsername string, cost int64, maxDebt int64) error
	BuildMarketEntity(now time.Time, req MarketCreateRequest, creatorUsername string, labels labelPair) *Market
}

// ResolutionPolicy encapsulates resolution rules and post-resolution actions.
type ResolutionPolicy interface {
	NormalizeResolution(resolution string) (string, error)
	ValidateResolutionRequest(market *Market, username string) error
	Resolve(ctx context.Context, repo Repository, userService UserService, marketID int64, outcome string) error
}

// ProbabilityChange represents a change in market probability at a timestamp.
type ProbabilityChange struct {
	Probability float64
	Timestamp   time.Time
}

// ProbabilityEngine provides probability calculations and projections.
type ProbabilityEngine interface {
	Calculate(createdAt time.Time, bets []models.Bet) []ProbabilityChange
	Project(createdAt time.Time, bets []models.Bet, newBet models.Bet) ProbabilityProjection
}

// ProbabilityValidator validates projection requests and market eligibility.
type ProbabilityValidator interface {
	ValidateRequest(req ProbabilityProjectionRequest) error
	ValidateMarket(market *Market, now time.Time) error
}

// SearchPolicy encapsulates search query validation and fallback logic.
type SearchPolicy interface {
	ValidateQuery(query string) error
	NormalizeFilters(filters SearchFilters) SearchFilters
	ShouldFetchFallback(primary []*Market, status string) bool
	NewSearchResults(query string, status string, primary []*Market) *SearchResults
	BuildFallbackFilters(primary SearchFilters) SearchFilters
	SelectFallback(primary []*Market, all []*Market, limit int) []*Market
}

// MetricsCalculator computes volume and dust metrics.
type MetricsCalculator interface {
	Volume(bets []models.Bet) int64
	VolumeWithDust(bets []models.Bet) int64
	Dust(bets []models.Bet) int64
}

// LeaderboardCalculator computes leaderboard standings.
type LeaderboardCalculator interface {
	Calculate(snapshot positionsmath.MarketSnapshot, bets []models.Bet) ([]positionsmath.UserProfitability, error)
}

// StatusPolicy validates status filters and pagination rules.
type StatusPolicy interface {
	ValidateStatus(status string) error
	NormalizePage(p Page, defaultLimit, maxLimit int) Page
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

	creationPolicy        CreationPolicy
	resolutionPolicy      ResolutionPolicy
	probabilityEngine     ProbabilityEngine
	probabilityValidator  ProbabilityValidator
	searchPolicy          SearchPolicy
	metricsCalculator     MetricsCalculator
	leaderboardCalculator LeaderboardCalculator
	statusPolicy          StatusPolicy
}

// NewService creates a new markets service.
func NewService(repo Repository, userService UserService, clock Clock, config Config) *Service {
	if clock == nil {
		clock = serviceClock{}
	}
	return &Service{
		repo:        repo,
		userService: userService,
		clock:       clock,
		config:      config,

		creationPolicy:        defaultCreationPolicy{config: config},
		resolutionPolicy:      defaultResolutionPolicy{},
		probabilityEngine:     defaultProbabilityEngine{},
		probabilityValidator:  defaultProbabilityValidator{},
		searchPolicy:          defaultSearchPolicy{},
		metricsCalculator:     defaultMetricsCalculator{},
		leaderboardCalculator: defaultLeaderboardCalculator{},
		statusPolicy:          defaultStatusPolicy{},
	}
}

var (
	_ ServiceInterface = (*Service)(nil)
	_ Clock            = serviceClock{}
)

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
