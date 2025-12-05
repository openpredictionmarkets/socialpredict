package app

import (
	"time"

	"socialpredict/setup"

	"gorm.io/gorm"

	// Domain services
	analytics "socialpredict/internal/domain/analytics"
	dbets "socialpredict/internal/domain/bets"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"

	// Repositories
	rbets "socialpredict/internal/repository/bets"
	rmarkets "socialpredict/internal/repository/markets"
	rusers "socialpredict/internal/repository/users"

	// Handlers
	hmarkets "socialpredict/handlers/markets"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/security"
)

// Clock interface for testability
type Clock interface {
	Now() time.Time
}

// SystemClock implements Clock using system time
type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

// Container holds all the application dependencies
type Container struct {
	db     *gorm.DB
	config *setup.EconomicConfig
	clock  Clock

	// Repositories
	marketsRepo   rmarkets.GormRepository
	usersRepo     rusers.GormRepository
	analyticsRepo analytics.GormRepository
	betsRepo      rbets.GormRepository

	// Domain services
	analyticsService *analytics.Service
	marketsService   *dmarkets.Service
	usersService     *dusers.Service
	betsService      *dbets.Service
	authService      *authsvc.AuthService

	// Handlers
	marketsHandler *hmarkets.Handler
}

// NewContainer creates a new dependency injection container
func NewContainer(db *gorm.DB, config *setup.EconomicConfig) *Container {
	return &Container{
		db:     db,
		config: config,
		clock:  SystemClock{},
	}
}

// InitializeRepositories sets up all repository implementations
func (c *Container) InitializeRepositories() {
	c.marketsRepo = *rmarkets.NewGormRepository(c.db)
	c.usersRepo = *rusers.NewGormRepository(c.db)
	c.analyticsRepo = *analytics.NewGormRepository(c.db)
	c.betsRepo = *rbets.NewGormRepository(c.db)
}

// InitializeServices sets up all domain services with their dependencies
func (c *Container) InitializeServices() {
	// Users service depends on users repository and configuration
	securityService := security.NewSecurityService()
	configLoader := func() *setup.EconomicConfig { return c.config }
	c.analyticsService = analytics.NewService(&c.analyticsRepo, configLoader)
	c.usersService = dusers.NewService(&c.usersRepo, c.analyticsService, securityService.Sanitizer)
	c.authService = authsvc.NewAuthService(c.usersService)

	// Markets service depends on markets repository and users service
	marketsConfig := dmarkets.Config{
		MinimumFutureHours: c.config.Economics.MarketCreation.MinimumFutureHours,
		CreateMarketCost:   c.config.Economics.MarketIncentives.CreateMarketCost,
		MaximumDebtAllowed: c.config.Economics.User.MaximumDebtAllowed,
	}

	c.marketsService = dmarkets.NewService(&c.marketsRepo, c.usersService, c.clock, marketsConfig)

	c.betsService = dbets.NewService(&c.betsRepo, c.marketsService, c.usersService, c.config, c.clock)
}

// InitializeHandlers sets up all HTTP handlers with their service dependencies
func (c *Container) InitializeHandlers() {
	c.marketsHandler = hmarkets.NewHandler(c.marketsService, c.authService)
}

// Initialize sets up the entire dependency graph
func (c *Container) Initialize() {
	c.InitializeRepositories()
	c.InitializeServices()
	c.InitializeHandlers()
}

// GetMarketsHandler returns the markets HTTP handler
func (c *Container) GetMarketsHandler() *hmarkets.Handler {
	return c.marketsHandler
}

// GetUsersService returns the users domain service
func (c *Container) GetUsersService() *dusers.Service {
	return c.usersService
}

// GetAnalyticsService returns the analytics domain service
func (c *Container) GetAnalyticsService() *analytics.Service {
	return c.analyticsService
}

// GetMarketsService returns the markets domain service
func (c *Container) GetMarketsService() *dmarkets.Service {
	return c.marketsService
}

// GetBetsService returns the bets domain service
func (c *Container) GetBetsService() *dbets.Service {
	return c.betsService
}

// GetAuthService returns the authentication façade
func (c *Container) GetAuthService() *authsvc.AuthService {
	return c.authService
}

// BuildApplication creates a fully wired application container
func BuildApplication(db *gorm.DB, config *setup.EconomicConfig) *Container {
	container := NewContainer(db, config)
	container.Initialize()
	return container
}
