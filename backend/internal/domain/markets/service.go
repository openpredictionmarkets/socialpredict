package markets

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	MaxQuestionTitleLength = 160
	MaxDescriptionLength   = 2000
	MaxLabelLength         = 20
	MinLabelLength         = 1
)

// Clock provides time functionality for testability
type Clock interface {
	Now() time.Time
}

// Repository defines the interface for market data access
type Repository interface {
	Create(ctx context.Context, market *Market) error
	GetByID(ctx context.Context, id int64) (*Market, error)
	UpdateLabels(ctx context.Context, id int64, yesLabel, noLabel string) error
	List(ctx context.Context, filters ListFilters) ([]*Market, error)
	Search(ctx context.Context, query string, filters SearchFilters) ([]*Market, error)
	Delete(ctx context.Context, id int64) error
	ResolveMarket(ctx context.Context, id int64, resolution string) error
}

// UserService defines the interface for user-related operations
type UserService interface {
	ValidateUserExists(ctx context.Context, username string) error
	ValidateUserBalance(ctx context.Context, username string, requiredAmount float64, maxDebt float64) error
	DeductBalance(ctx context.Context, username string, amount float64) error
}

// Config holds configuration for the markets service
type Config struct {
	MinimumFutureHours float64
	CreateMarketCost   float64
	MaximumDebtAllowed float64
}

// ListFilters represents filters for listing markets
type ListFilters struct {
	Status    string
	CreatedBy string
	Limit     int
	Offset    int
}

// SearchFilters represents filters for searching markets
type SearchFilters struct {
	Status string
	Limit  int
	Offset int
}

// Service implements the core market business logic
type Service struct {
	repo        Repository
	userService UserService
	clock       Clock
	config      Config
}

// NewService creates a new markets service
func NewService(repo Repository, userService UserService, clock Clock, config Config) *Service {
	return &Service{
		repo:        repo,
		userService: userService,
		clock:       clock,
		config:      config,
	}
}

// CreateMarket creates a new market with validation
func (s *Service) CreateMarket(ctx context.Context, req MarketCreateRequest, creatorUsername string) (*Market, error) {
	// Validate question title length
	if err := s.validateQuestionTitle(req.QuestionTitle); err != nil {
		return nil, err
	}

	// Validate description length
	if err := s.validateDescription(req.Description); err != nil {
		return nil, err
	}

	// Validate custom labels
	if err := s.validateCustomLabels(req.YesLabel, req.NoLabel); err != nil {
		return nil, err
	}

	// Set default labels if not provided
	yesLabel := strings.TrimSpace(req.YesLabel)
	if yesLabel == "" {
		yesLabel = "YES"
	}

	noLabel := strings.TrimSpace(req.NoLabel)
	if noLabel == "" {
		noLabel = "NO"
	}

	// Validate user exists
	if err := s.userService.ValidateUserExists(ctx, creatorUsername); err != nil {
		return nil, ErrUserNotFound
	}

	// Validate market resolution time
	if err := s.ValidateMarketResolutionTime(req.ResolutionDateTime); err != nil {
		return nil, err
	}

	// Check user balance and deduct fee
	if err := s.userService.ValidateUserBalance(ctx, creatorUsername, s.config.CreateMarketCost, s.config.MaximumDebtAllowed); err != nil {
		return nil, ErrInsufficientBalance
	}

	// Deduct market creation fee
	if err := s.userService.DeductBalance(ctx, creatorUsername, s.config.CreateMarketCost); err != nil {
		return nil, err
	}

	// Create market object
	market := &Market{
		QuestionTitle:      req.QuestionTitle,
		Description:        req.Description,
		OutcomeType:        req.OutcomeType,
		ResolutionDateTime: req.ResolutionDateTime,
		CreatorUsername:    creatorUsername,
		YesLabel:           yesLabel,
		NoLabel:            noLabel,
		Status:             "active", // Default status
		CreatedAt:          s.clock.Now(),
		UpdatedAt:          s.clock.Now(),
	}

	// Create market in repository
	if err := s.repo.Create(ctx, market); err != nil {
		return nil, err
	}

	return market, nil
}

// SetCustomLabels updates the custom labels for a market
func (s *Service) SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error {
	// Validate labels
	if err := s.validateCustomLabels(yesLabel, noLabel); err != nil {
		return err
	}

	// Check market exists
	_, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return ErrMarketNotFound
	}

	// Update labels
	return s.repo.UpdateLabels(ctx, marketID, yesLabel, noLabel)
}

// GetMarket retrieves a market by ID
func (s *Service) GetMarket(ctx context.Context, id int64) (*Market, error) {
	return s.repo.GetByID(ctx, id)
}

// MarketOverview represents enriched market data with calculations
type MarketOverview struct {
	Market          *Market
	Creator         interface{} // Will be replaced with proper user type
	LastProbability float64
	NumUsers        int
	TotalVolume     int64
	MarketDust      int64
}

// ListMarkets returns a list of markets with filters
func (s *Service) ListMarkets(ctx context.Context, filters ListFilters) ([]*Market, error) {
	return s.repo.List(ctx, filters)
}

// GetMarketOverviews returns enriched market data with calculations
func (s *Service) GetMarketOverviews(ctx context.Context, filters ListFilters) ([]*MarketOverview, error) {
	markets, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	var overviews []*MarketOverview
	for _, market := range markets {
		overview := &MarketOverview{
			Market: market,
			// Complex calculations will be added here
			// This is placeholder for now - calculations should be moved from handlers
		}
		overviews = append(overviews, overview)
	}

	return overviews, nil
}

// GetMarketDetails returns detailed market information with calculations
func (s *Service) GetMarketDetails(ctx context.Context, marketID int64) (*MarketOverview, error) {
	market, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return nil, err
	}

	// Complex calculation logic will be moved here from marketdetailshandler.go
	overview := &MarketOverview{
		Market: market,
		// Calculations will be added here
	}

	return overview, nil
}

// SearchMarkets searches for markets by query
func (s *Service) SearchMarkets(ctx context.Context, query string, filters SearchFilters) ([]*Market, error) {
	return s.repo.Search(ctx, query, filters)
}

// ResolveMarket resolves a market with a given outcome
func (s *Service) ResolveMarket(ctx context.Context, marketID int64, resolution string) error {
	// Check market exists
	_, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return ErrMarketNotFound
	}

	return s.repo.ResolveMarket(ctx, marketID, resolution)
}

// ListActiveMarkets returns markets that are not resolved and active
func (s *Service) ListActiveMarkets(ctx context.Context, limit int) ([]*Market, error) {
	filters := ListFilters{
		Status: "active",
		Limit:  limit,
	}
	return s.repo.List(ctx, filters)
}

// ListClosedMarkets returns markets that are closed but not resolved
func (s *Service) ListClosedMarkets(ctx context.Context, limit int) ([]*Market, error) {
	filters := ListFilters{
		Status: "closed",
		Limit:  limit,
	}
	return s.repo.List(ctx, filters)
}

// ListResolvedMarkets returns markets that have been resolved
func (s *Service) ListResolvedMarkets(ctx context.Context, limit int) ([]*Market, error) {
	filters := ListFilters{
		Status: "resolved",
		Limit:  limit,
	}
	return s.repo.List(ctx, filters)
}

// validateQuestionTitle validates the market question title
func (s *Service) validateQuestionTitle(title string) error {
	if len(title) > MaxQuestionTitleLength || len(title) < 1 {
		return ErrInvalidQuestionLength
	}
	return nil
}

// validateDescription validates the market description
func (s *Service) validateDescription(description string) error {
	if len(description) > MaxDescriptionLength {
		return ErrInvalidDescriptionLength
	}
	return nil
}

// validateCustomLabels validates the custom yes/no labels
func (s *Service) validateCustomLabels(yesLabel, noLabel string) error {
	// Validate yes label
	if yesLabel != "" {
		yesLabel = strings.TrimSpace(yesLabel)
		if len(yesLabel) < MinLabelLength || len(yesLabel) > MaxLabelLength {
			return ErrInvalidLabel
		}
	}

	// Validate no label
	if noLabel != "" {
		noLabel = strings.TrimSpace(noLabel)
		if len(noLabel) < MinLabelLength || len(noLabel) > MaxLabelLength {
			return ErrInvalidLabel
		}
	}

	return nil
}

// ValidateQuestionTitle validates the market question title
func (s *Service) ValidateQuestionTitle(title string) error {
	if len(title) > MaxQuestionTitleLength || len(title) < 1 {
		return ErrInvalidQuestionLength
	}
	return nil
}

// ValidateDescription validates the market description
func (s *Service) ValidateDescription(description string) error {
	if len(description) > MaxDescriptionLength {
		return ErrInvalidDescriptionLength
	}
	return nil
}

// ValidateLabels validates the custom yes/no labels
func (s *Service) ValidateLabels(yesLabel, noLabel string) error {
	// Validate yes label
	if yesLabel != "" {
		yesLabel = strings.TrimSpace(yesLabel)
		if len(yesLabel) < MinLabelLength || len(yesLabel) > MaxLabelLength {
			return ErrInvalidLabel
		}
	}

	// Validate no label
	if noLabel != "" {
		noLabel = strings.TrimSpace(noLabel)
		if len(noLabel) < MinLabelLength || len(noLabel) > MaxLabelLength {
			return ErrInvalidLabel
		}
	}

	return nil
}

// validateMarketResolutionTime validates that the market resolution time meets business logic requirements (private)
func (s *Service) ValidateMarketResolutionTime(resolutionTime time.Time) error {
	now := s.clock.Now()
	minimumDuration := time.Duration(s.config.MinimumFutureHours * float64(time.Hour))
	minimumFutureTime := now.Add(minimumDuration)

	if resolutionTime.Before(minimumFutureTime) || resolutionTime.Equal(minimumFutureTime) {
		return fmt.Errorf("market resolution time must be at least %.1f hours in the future", s.config.MinimumFutureHours)
	}
	return nil
}
