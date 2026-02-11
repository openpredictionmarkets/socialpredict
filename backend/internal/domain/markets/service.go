package markets

// Constants for validation limits.
const (
	MaxQuestionTitleLength = 160
	MaxDescriptionLength   = 2000
	MaxLabelLength         = 20
	MinLabelLength         = 1
)

// Config holds configuration for the markets service.
type Config struct {
	MinimumFutureHours float64
	CreateMarketCost   int64
	MaximumDebtAllowed int64
}

// Service implements the core market business logic.
type Service struct {
	repo                  Repository
	creatorProfileService CreatorProfileService
	walletService         WalletService
	clock                 Clock
	config                Config
}

// NewServiceWithWallet creates a markets service with explicit creator/profile and wallet dependencies.
func NewServiceWithWallet(repo Repository, creatorProfileService CreatorProfileService, walletService WalletService, clock Clock, config Config) *Service {
	return &Service{
		repo:                  repo,
		creatorProfileService: creatorProfileService,
		walletService:         walletService,
		clock:                 clock,
		config:                config,
	}
}

// Compile-time interface compliance checks.
var (
	_ ServiceInterface   = (*Service)(nil)
	_ CoreService        = (*Service)(nil)
	_ SearchService      = (*Service)(nil)
	_ LeaderboardService = (*Service)(nil)
	_ PositionsService   = (*Service)(nil)
	_ ProjectionService  = (*Service)(nil)
)
