package markets

import (
	"context"
	"time"

	users "socialpredict/internal/domain/users"
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

// CreatorProfileService defines creator existence and profile lookups used by markets.
type CreatorProfileService interface {
	ValidateUserExists(ctx context.Context, username string) error
	GetPublicUser(ctx context.Context, username string) (*users.PublicUser, error)
}

// WalletService defines balance operations used by markets.
type WalletService interface {
	ValidateBalance(ctx context.Context, username string, amount int64, maxDebt int64) error
	Debit(ctx context.Context, username string, amount int64, maxDebt int64, txType string) error
	Credit(ctx context.Context, username string, amount int64, txType string) error
}

// UserService is a temporary compatibility interface for legacy constructor wiring.
// It will be removed once callers inject creator/profile and wallet dependencies directly.
type UserService interface {
	CreatorProfileService
	ValidateUserBalance(ctx context.Context, username string, requiredAmount int64, maxDebt int64) error
	DeductBalance(ctx context.Context, username string, amount int64) error
	ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error
}

// --- Focused Service Interfaces ---

// CoreService handles core market CRUD operations.
type CoreService interface {
	CreateMarket(ctx context.Context, req MarketCreateRequest, creatorUsername string) (*Market, error)
	SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error
	GetMarket(ctx context.Context, id int64) (*Market, error)
	GetPublicMarket(ctx context.Context, marketID int64) (*PublicMarket, error)
	ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error
}

// SearchService handles market search and listing.
type SearchService interface {
	ListMarkets(ctx context.Context, filters ListFilters) ([]*Market, error)
	SearchMarkets(ctx context.Context, query string, filters SearchFilters) (*SearchResults, error)
	ListByStatus(ctx context.Context, status string, p Page) ([]*Market, error)
}

// LeaderboardService handles market leaderboard calculations.
type LeaderboardService interface {
	GetMarketLeaderboard(ctx context.Context, marketID int64, p Page) ([]*LeaderboardRow, error)
}

// PositionsService handles user position queries.
type PositionsService interface {
	GetMarketPositions(ctx context.Context, marketID int64) (MarketPositions, error)
	GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*UserPosition, error)
}

// ProjectionService handles probability projections and market details.
type ProjectionService interface {
	ProjectProbability(ctx context.Context, req ProbabilityProjectionRequest) (*ProbabilityProjection, error)
	GetMarketDetails(ctx context.Context, marketID int64) (*MarketOverview, error)
	GetMarketBets(ctx context.Context, marketID int64) ([]*BetDisplayInfo, error)
	CalculateMarketVolume(ctx context.Context, marketID int64) (int64, error)
}

// ServiceInterface combines all market service capabilities.
// Prefer using focused interfaces (CoreService, SearchService, etc.) for new code.
type ServiceInterface interface {
	CoreService
	SearchService
	LeaderboardService
	PositionsService
	ProjectionService
}
