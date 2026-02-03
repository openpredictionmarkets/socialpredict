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

// Compile-time interface compliance checks.
var (
	_ ServiceInterface   = (*Service)(nil)
	_ CoreService        = (*Service)(nil)
	_ SearchService      = (*Service)(nil)
	_ LeaderboardService = (*Service)(nil)
	_ PositionsService   = (*Service)(nil)
	_ ProjectionService  = (*Service)(nil)
)
