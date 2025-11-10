package markets

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	marketmath "socialpredict/internal/domain/math/market"
	"socialpredict/internal/domain/math/probabilities/wpam"
	users "socialpredict/internal/domain/users"
	"socialpredict/models"
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

// UserService defines the interface for user-related operations
type UserService interface {
	ValidateUserExists(ctx context.Context, username string) error
	ValidateUserBalance(ctx context.Context, username string, requiredAmount int64, maxDebt int64) error
	DeductBalance(ctx context.Context, username string, amount int64) error
	ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error
	GetPublicUser(ctx context.Context, username string) (*users.PublicUser, error)
}

// Config holds configuration for the markets service
type Config struct {
	MinimumFutureHours float64
	CreateMarketCost   int64
	MaximumDebtAllowed int64
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

// SearchResults represents the result of a market search with fallback
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

// ServiceInterface defines the interface for market service operations
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

// GetPublicMarket returns a public representation of a market.
func (s *Service) GetPublicMarket(ctx context.Context, marketID int64) (*PublicMarket, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.GetPublicMarket(ctx, marketID)
}

// MarketOverview represents enriched market data with calculations
type MarketOverview struct {
	Market             *Market
	Creator            *CreatorSummary
	ProbabilityChanges []ProbabilityPoint
	LastProbability    float64
	NumUsers           int
	TotalVolume        int64
	MarketDust         int64
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
			Market:  market,
			Creator: s.buildCreatorSummary(ctx, market.CreatorUsername),
			// Complex calculations will be added here
			// This is placeholder for now - calculations should be moved from handlers
		}
		overviews = append(overviews, overview)
	}

	return overviews, nil
}

// GetMarketDetails returns detailed market information with calculations
func (s *Service) GetMarketDetails(ctx context.Context, marketID int64) (*MarketOverview, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}

	market, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return nil, err
	}

	bets, err := s.repo.ListBetsForMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}

	modelBets := convertToModelBets(bets)
	probabilityChanges := wpam.CalculateMarketProbabilitiesWPAM(market.CreatedAt, modelBets)
	probabilityPoints := make([]ProbabilityPoint, len(probabilityChanges))
	for i, change := range probabilityChanges {
		probabilityPoints[i] = ProbabilityPoint{
			Probability: change.Probability,
			Timestamp:   change.Timestamp,
		}
	}

	lastProbability := 0.0
	if len(probabilityPoints) > 0 {
		lastProbability = probabilityPoints[len(probabilityPoints)-1].Probability
	}

	totalVolumeWithDust := marketmath.GetMarketVolumeWithDust(modelBets)
	marketDust := marketmath.GetMarketDust(modelBets)
	numUsers := countUniqueUsers(modelBets)

	return &MarketOverview{
		Market:             market,
		Creator:            s.buildCreatorSummary(ctx, market.CreatorUsername),
		ProbabilityChanges: probabilityPoints,
		LastProbability:    lastProbability,
		NumUsers:           numUsers,
		TotalVolume:        totalVolumeWithDust,
		MarketDust:         marketDust,
	}, nil
}

func (s *Service) buildCreatorSummary(ctx context.Context, username string) *CreatorSummary {
	summary := &CreatorSummary{Username: username}
	if s.userService == nil {
		return summary
	}
	user, err := s.userService.GetPublicUser(ctx, username)
	if err != nil || user == nil {
		return summary
	}
	summary.DisplayName = user.DisplayName
	summary.PersonalEmoji = user.PersonalEmoji
	return summary
}

func convertToModelBets(bets []*Bet) []models.Bet {
	if len(bets) == 0 {
		return []models.Bet{}
	}
	out := make([]models.Bet, len(bets))
	for i, bet := range bets {
		out[i] = models.Bet{
			Username: bet.Username,
			MarketID: bet.MarketID,
			Amount:   bet.Amount,
			PlacedAt: bet.PlacedAt,
			Outcome:  bet.Outcome,
		}
	}
	return out
}

func countUniqueUsers(bets []models.Bet) int {
	if len(bets) == 0 {
		return 0
	}
	seen := make(map[string]struct{})
	for _, bet := range bets {
		if bet.Username == "" {
			continue
		}
		if _, ok := seen[bet.Username]; !ok {
			seen[bet.Username] = struct{}{}
		}
	}
	return len(seen)
}

// SearchMarkets searches for markets by query with fallback logic
func (s *Service) SearchMarkets(ctx context.Context, query string, filters SearchFilters) (*SearchResults, error) {
	// Validate query
	if strings.TrimSpace(query) == "" {
		return nil, ErrInvalidInput
	}

	// Validate and set defaults
	if filters.Limit <= 0 || filters.Limit > 50 {
		filters.Limit = 20
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	// Primary search within specified status
	primaryResults, err := s.repo.Search(ctx, query, filters)
	if err != nil {
		return nil, err
	}

	searchResults := &SearchResults{
		PrimaryResults:  primaryResults,
		FallbackResults: []*Market{},
		Query:           query,
		PrimaryStatus:   filters.Status,
		PrimaryCount:    len(primaryResults),
		FallbackCount:   0,
		TotalCount:      len(primaryResults),
		FallbackUsed:    false,
	}

	// If we have 5 or fewer primary results and we're not already searching "all", search all markets
	if len(primaryResults) <= 5 && filters.Status != "" && filters.Status != "all" {
		// Search all markets for fallback
		allFilters := SearchFilters{
			Status: "", // Empty means search all
			Limit:  filters.Limit * 2,
			Offset: 0,
		}

		allResults, err := s.repo.Search(ctx, query, allFilters)
		if err != nil {
			return searchResults, nil // Return primary results even if fallback fails
		}

		// Filter out markets that are already in primary results
		primaryIDs := make(map[int64]bool)
		for _, market := range primaryResults {
			primaryIDs[market.ID] = true
		}

		var fallbackResults []*Market
		for _, market := range allResults {
			if !primaryIDs[market.ID] {
				fallbackResults = append(fallbackResults, market)
				if len(fallbackResults) >= filters.Limit {
					break
				}
			}
		}

		if len(fallbackResults) > 0 {
			searchResults.FallbackResults = fallbackResults
			searchResults.FallbackCount = len(fallbackResults)
			searchResults.TotalCount = searchResults.PrimaryCount + searchResults.FallbackCount
			searchResults.FallbackUsed = true
		}
	}

	return searchResults, nil
}

// ResolveMarket resolves a market with a given outcome
func (s *Service) ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error {
	outcome := strings.ToUpper(strings.TrimSpace(resolution))
	if outcome != "YES" && outcome != "NO" && outcome != "N/A" {
		return ErrInvalidInput
	}

	market, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return ErrMarketNotFound
	}

	if market.CreatorUsername != username {
		return ErrUnauthorized
	}

	if market.Status == "resolved" {
		return ErrInvalidState
	}

	if err := s.repo.ResolveMarket(ctx, marketID, outcome); err != nil {
		return err
	}

	switch outcome {
	case "N/A":
		bets, err := s.repo.ListBetsForMarket(ctx, marketID)
		if err != nil {
			return err
		}
		for _, bet := range bets {
			if err := s.userService.ApplyTransaction(ctx, bet.Username, bet.Amount, users.TransactionRefund); err != nil {
				return err
			}
		}
	default: // YES or NO
		positions, err := s.repo.CalculatePayoutPositions(ctx, marketID)
		if err != nil {
			return err
		}
		for _, pos := range positions {
			if pos.Value <= 0 {
				continue
			}
			if err := s.userService.ApplyTransaction(ctx, pos.Username, pos.Value, users.TransactionWin); err != nil {
				return err
			}
		}
	}

	return nil
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

// Page represents pagination parameters
type Page struct {
	Limit  int
	Offset int
}

// LeaderboardRow represents a single row in the market leaderboard
type LeaderboardRow struct {
	Username string
	Profit   float64
	Volume   int64
	Rank     int
}

// ProbabilityProjectionRequest represents a request for probability projection
type ProbabilityProjectionRequest struct {
	MarketID int64
	Amount   int64
	Outcome  string
}

// ProbabilityProjection represents the result of a probability projection
type ProbabilityProjection struct {
	CurrentProbability   float64
	ProjectedProbability float64
}

// ListByStatus returns markets filtered by status with pagination
func (s *Service) ListByStatus(ctx context.Context, status string, p Page) ([]*Market, error) {
	// Validate status
	switch status {
	case "active", "closed", "resolved", "all":
		// Valid status
	default:
		return nil, ErrInvalidInput
	}

	// Validate pagination
	if p.Limit <= 0 {
		p.Limit = 100
	}
	if p.Limit > 1000 {
		p.Limit = 1000
	}
	if p.Offset < 0 {
		p.Offset = 0
	}

	return s.repo.ListByStatus(ctx, status, p)
}

// GetMarketLeaderboard returns the leaderboard for a specific market
func (s *Service) GetMarketLeaderboard(ctx context.Context, marketID int64, p Page) ([]*LeaderboardRow, error) {
	// 1. Validate market exists
	_, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return nil, ErrMarketNotFound
	}

	// 2. Validate pagination
	if p.Limit <= 0 {
		p.Limit = 100
	}
	if p.Limit > 1000 {
		p.Limit = 1000
	}
	if p.Offset < 0 {
		p.Offset = 0
	}

	// 3. Call repository to get leaderboard data
	// This will be implemented in repository layer
	// For now, return empty slice - calculations will be moved here from handlers
	var leaderboard []*LeaderboardRow

	// TODO: Move leaderboard calculation logic from positionsmath.CalculateMarketLeaderboard here
	// This should involve:
	// - Getting all bets for the market
	// - Calculating profit/loss for each user
	// - Ranking users by profitability
	// - Applying pagination

	return leaderboard, nil
}

// ProjectProbability projects what the probability would be after a hypothetical bet
func (s *Service) ProjectProbability(ctx context.Context, req ProbabilityProjectionRequest) (*ProbabilityProjection, error) {
	// 1. Validate market exists
	if req.MarketID <= 0 || req.MarketID > int64(math.MaxUint32) || strings.TrimSpace(req.Outcome) == "" || req.Amount <= 0 {
		return nil, ErrInvalidInput
	}

	outcome := strings.ToUpper(strings.TrimSpace(req.Outcome))
	if outcome != "YES" && outcome != "NO" {
		return nil, ErrInvalidInput
	}

	market, err := s.repo.GetByID(ctx, req.MarketID)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(market.Status, "resolved") {
		return nil, ErrInvalidState
	}

	now := s.clock.Now()
	if now.After(market.ResolutionDateTime) {
		return nil, ErrInvalidState
	}

	bets, err := s.repo.ListBetsForMarket(ctx, req.MarketID)
	if err != nil {
		return nil, err
	}

	modelBets := convertToModelBets(bets)
	probabilityTrack := wpam.CalculateMarketProbabilitiesWPAM(market.CreatedAt, modelBets)

	currentProbability := 0.5
	if len(probabilityTrack) > 0 {
		currentProbability = probabilityTrack[len(probabilityTrack)-1].Probability
	}

	newBet := models.Bet{
		Username: "preview",
		MarketID: uint(req.MarketID),
		Amount:   req.Amount,
		Outcome:  outcome,
		PlacedAt: now,
	}

	projection := wpam.ProjectNewProbabilityWPAM(market.CreatedAt, modelBets, newBet)

	return &ProbabilityProjection{
		CurrentProbability:   currentProbability,
		ProjectedProbability: projection.Probability,
	}, nil
}

// CalculateMarketVolume returns the total traded volume for a market.
func (s *Service) CalculateMarketVolume(ctx context.Context, marketID int64) (int64, error) {
	if marketID <= 0 {
		return 0, ErrInvalidInput
	}

	if _, err := s.repo.GetByID(ctx, marketID); err != nil {
		return 0, err
	}

	bets, err := s.repo.ListBetsForMarket(ctx, marketID)
	if err != nil {
		return 0, err
	}

	modelBets := convertToModelBets(bets)
	return marketmath.GetMarketVolume(modelBets), nil
}

// BetDisplayInfo represents a bet with probability information
type BetDisplayInfo struct {
	Username    string    `json:"username"`
	Outcome     string    `json:"outcome"`
	Amount      int64     `json:"amount"`
	Probability float64   `json:"probability"`
	PlacedAt    time.Time `json:"placedAt"`
}

// GetMarketBets returns the bet history for a market with probabilities
func (s *Service) GetMarketBets(ctx context.Context, marketID int64) ([]*BetDisplayInfo, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}

	market, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return nil, err
	}

	bets, err := s.repo.ListBetsForMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if len(bets) == 0 {
		return []*BetDisplayInfo{}, nil
	}

	modelBets := convertToModelBets(bets)
	probabilityChanges := wpam.CalculateMarketProbabilitiesWPAM(market.CreatedAt, modelBets)
	if len(probabilityChanges) == 0 {
		probabilityChanges = []wpam.ProbabilityChange{{
			Probability: 0,
			Timestamp:   market.CreatedAt,
		}}
	}

	sort.Slice(probabilityChanges, func(i, j int) bool {
		return probabilityChanges[i].Timestamp.Before(probabilityChanges[j].Timestamp)
	})

	// Ensure bets are processed in chronological order
	sort.Slice(modelBets, func(i, j int) bool {
		return modelBets[i].PlacedAt.Before(modelBets[j].PlacedAt)
	})

	results := make([]*BetDisplayInfo, 0, len(modelBets))
	for _, bet := range modelBets {
		matchedProbability := probabilityChanges[0].Probability
		for _, change := range probabilityChanges {
			if change.Timestamp.After(bet.PlacedAt) {
				break
			}
			matchedProbability = change.Probability
		}

		results = append(results, &BetDisplayInfo{
			Username:    bet.Username,
			Outcome:     bet.Outcome,
			Amount:      bet.Amount,
			Probability: matchedProbability,
			PlacedAt:    bet.PlacedAt,
		})
	}

	return results, nil
}

// GetMarketPositions returns all user positions in a market
func (s *Service) GetMarketPositions(ctx context.Context, marketID int64) (MarketPositions, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}

	// Ensure market exists
	if _, err := s.repo.GetByID(ctx, marketID); err != nil {
		return nil, err
	}

	positions, err := s.repo.ListMarketPositions(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if positions == nil {
		return MarketPositions{}, nil
	}
	return positions, nil
}

// GetUserPositionInMarket returns a specific user's position in a market
func (s *Service) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*UserPosition, error) {
	// 1. Validate market exists
	_, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return nil, ErrMarketNotFound
	}

	if strings.TrimSpace(username) == "" {
		return nil, ErrInvalidInput
	}

	position, err := s.repo.GetUserPosition(ctx, marketID, username)
	if err != nil {
		return nil, err
	}
	return position, nil
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
