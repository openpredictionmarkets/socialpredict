package markets

import (
	"context"
	"math"
	"strings"
	"time"

	marketmath "socialpredict/internal/domain/math/market"
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/internal/domain/math/probabilities/wpam"
	users "socialpredict/internal/domain/users"
	"socialpredict/models"
)

type defaultCreationPolicy struct {
	config Config
}

func (p defaultCreationPolicy) ValidateCreateRequest(req MarketCreateRequest) error {
	if len(req.QuestionTitle) > MaxQuestionTitleLength || len(req.QuestionTitle) < 1 {
		return ErrInvalidQuestionLength
	}
	if len(req.Description) > MaxDescriptionLength {
		return ErrInvalidDescriptionLength
	}
	return p.ValidateCustomLabels(req.YesLabel, req.NoLabel)
}

func (p defaultCreationPolicy) ValidateCustomLabels(yesLabel, noLabel string) error {
	if yesLabel != "" {
		yesLabel = strings.TrimSpace(yesLabel)
		if len(yesLabel) < MinLabelLength || len(yesLabel) > MaxLabelLength {
			return ErrInvalidLabel
		}
	}

	if noLabel != "" {
		noLabel = strings.TrimSpace(noLabel)
		if len(noLabel) < MinLabelLength || len(noLabel) > MaxLabelLength {
			return ErrInvalidLabel
		}
	}

	return nil
}

func (p defaultCreationPolicy) NormalizeLabels(yesLabel string, noLabel string) labelPair {
	y := strings.TrimSpace(yesLabel)
	n := strings.TrimSpace(noLabel)
	if y == "" {
		y = "YES"
	}
	if n == "" {
		n = "NO"
	}
	return labelPair{yes: y, no: n}
}

func (p defaultCreationPolicy) ValidateResolutionTime(now time.Time, resolution time.Time, minimumFutureHours float64) error {
	minimumDuration := time.Duration(minimumFutureHours * float64(time.Hour))
	minimumFutureTime := now.Add(minimumDuration)

	if resolution.Before(minimumFutureTime) || resolution.Equal(minimumFutureTime) {
		return ErrInvalidResolutionTime
	}
	return nil
}

func (p defaultCreationPolicy) EnsureCreateMarketBalance(ctx context.Context, usersSvc UserService, creatorUsername string, cost int64, maxDebt int64) error {
	if err := usersSvc.ValidateUserBalance(ctx, creatorUsername, cost, maxDebt); err != nil {
		return ErrInsufficientBalance
	}
	return usersSvc.DeductBalance(ctx, creatorUsername, cost)
}

func (p defaultCreationPolicy) BuildMarketEntity(now time.Time, req MarketCreateRequest, creatorUsername string, labels labelPair) *Market {
	return &Market{
		QuestionTitle:      req.QuestionTitle,
		Description:        req.Description,
		OutcomeType:        req.OutcomeType,
		ResolutionDateTime: req.ResolutionDateTime,
		CreatorUsername:    creatorUsername,
		YesLabel:           labels.yes,
		NoLabel:            labels.no,
		Status:             "active",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

type defaultResolutionPolicy struct{}

func (defaultResolutionPolicy) NormalizeResolution(resolution string) (string, error) {
	outcome := strings.ToUpper(strings.TrimSpace(resolution))
	switch outcome {
	case "YES", "NO", "N/A":
		return outcome, nil
	default:
		return "", ErrInvalidInput
	}
}

func (defaultResolutionPolicy) ValidateResolutionRequest(market *Market, username string) error {
	if market.CreatorUsername != username {
		return ErrUnauthorized
	}

	if market.Status == "resolved" {
		return ErrInvalidState
	}

	return nil
}

// ResolutionRepository contains the persistence needs for market resolution flows.
type ResolutionRepository interface {
	ResolveMarket(ctx context.Context, id int64, resolution string) error
	ListBetsForMarket(ctx context.Context, marketID int64) ([]*Bet, error)
	CalculatePayoutPositions(ctx context.Context, marketID int64) ([]*PayoutPosition, error)
}

func (defaultResolutionPolicy) Resolve(ctx context.Context, repo ResolutionRepository, userService UserService, marketID int64, outcome string) error {
	if err := repo.ResolveMarket(ctx, marketID, outcome); err != nil {
		return err
	}

	if outcome == "N/A" {
		return refundMarketBets(ctx, repo, userService, marketID)
	}

	return payoutWinningPositions(ctx, repo, userService, marketID)
}

type defaultProbabilityEngine struct {
	calculator wpam.ProbabilityCalculator
}

// DefaultProbabilityEngine builds the WPAM-backed probability engine with a supplied calculator.
func DefaultProbabilityEngine(calculator wpam.ProbabilityCalculator) ProbabilityEngine {
	return defaultProbabilityEngine{calculator: calculator}
}

func (e defaultProbabilityEngine) ensureCalculator() wpam.ProbabilityCalculator {
	if e.calculator.Seeds().InitialSubsidization == 0 {
		return wpam.NewProbabilityCalculator(nil)
	}
	return e.calculator
}

func (e defaultProbabilityEngine) Calculate(createdAt time.Time, bets []models.Bet) []ProbabilityChange {
	calculator := e.ensureCalculator()
	changes := calculator.CalculateMarketProbabilitiesWPAM(createdAt, bets)
	points := make([]ProbabilityChange, len(changes))
	for i, change := range changes {
		points[i] = ProbabilityChange{
			Probability: change.Probability,
			Timestamp:   change.Timestamp,
		}
	}
	return points
}

func (e defaultProbabilityEngine) Project(createdAt time.Time, bets []models.Bet, newBet models.Bet) ProbabilityProjection {
	calculator := e.ensureCalculator()
	projection := calculator.ProjectNewProbabilityWPAM(createdAt, bets, newBet)
	return ProbabilityProjection{
		ProjectedProbability: projection.Probability,
	}
}

type defaultProbabilityValidator struct{}

func (defaultProbabilityValidator) ValidateRequest(req ProbabilityProjectionRequest) error {
	if req.MarketID <= 0 || req.MarketID > int64(math.MaxUint32) || strings.TrimSpace(req.Outcome) == "" || req.Amount <= 0 {
		return ErrInvalidInput
	}

	outcome := strings.ToUpper(strings.TrimSpace(req.Outcome))
	if outcome != "YES" && outcome != "NO" {
		return ErrInvalidInput
	}
	return nil
}

func (defaultProbabilityValidator) ValidateMarket(market *Market, now time.Time) error {
	if strings.EqualFold(market.Status, "resolved") {
		return ErrInvalidState
	}

	if now.After(market.ResolutionDateTime) {
		return ErrInvalidState
	}
	return nil
}

type defaultSearchPolicy struct{}

func (defaultSearchPolicy) ValidateQuery(query string) error {
	if strings.TrimSpace(query) == "" {
		return ErrInvalidInput
	}
	return nil
}

func (defaultSearchPolicy) NormalizeFilters(filters SearchFilters) SearchFilters {
	if filters.Limit <= 0 || filters.Limit > 50 {
		filters.Limit = 20
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	return filters
}

func (defaultSearchPolicy) ShouldFetchFallback(primary []*Market, status string) bool {
	return len(primary) <= 5 && status != "" && status != "all"
}

func (defaultSearchPolicy) NewSearchResults(query string, status string, primary []*Market) *SearchResults {
	return &SearchResults{
		PrimaryResults:  primary,
		FallbackResults: []*Market{},
		Query:           query,
		PrimaryStatus:   status,
		PrimaryCount:    len(primary),
		FallbackCount:   0,
		TotalCount:      len(primary),
		FallbackUsed:    false,
	}
}

func (defaultSearchPolicy) BuildFallbackFilters(primary SearchFilters) SearchFilters {
	return SearchFilters{
		Status: "",
		Limit:  primary.Limit * 2,
		Offset: 0,
	}
}

func (defaultSearchPolicy) SelectFallback(primary []*Market, all []*Market, limit int) []*Market {
	primaryIDs := make(map[int64]bool)
	for _, market := range primary {
		primaryIDs[market.ID] = true
	}

	var fallbackResults []*Market
	for _, market := range all {
		if primaryIDs[market.ID] {
			continue
		}
		fallbackResults = append(fallbackResults, market)
		if len(fallbackResults) >= limit {
			break
		}
	}
	return fallbackResults
}

type defaultMetricsCalculator struct{}

func (defaultMetricsCalculator) Volume(bets []models.Bet) int64 {
	return marketmath.GetMarketVolume(bets)
}

func (defaultMetricsCalculator) VolumeWithDust(bets []models.Bet) int64 {
	return marketmath.GetMarketVolumeWithDust(bets)
}

func (defaultMetricsCalculator) Dust(bets []models.Bet) int64 {
	return marketmath.GetMarketDust(bets)
}

type defaultLeaderboardCalculator struct{}

func (defaultLeaderboardCalculator) Calculate(snapshot positionsmath.MarketSnapshot, bets []models.Bet) ([]positionsmath.UserProfitability, error) {
	return positionsmath.CalculateMarketLeaderboard(snapshot, bets)
}

type defaultStatusPolicy struct{}

func (defaultStatusPolicy) ValidateStatus(status string) error {
	switch status {
	case "active", "closed", "resolved", "all":
		return nil
	default:
		return ErrInvalidInput
	}
}

func (defaultStatusPolicy) NormalizePage(p Page, defaultLimit, maxLimit int) Page {
	if p.Limit <= 0 {
		p.Limit = defaultLimit
	}
	if p.Limit > maxLimit {
		p.Limit = maxLimit
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p
}

func refundMarketBets(ctx context.Context, repo ResolutionRepository, userService UserService, marketID int64) error {
	bets, err := repo.ListBetsForMarket(ctx, marketID)
	if err != nil {
		return err
	}
	for _, bet := range bets {
		if err := userService.ApplyTransaction(ctx, bet.Username, bet.Amount, users.TransactionRefund); err != nil {
			return err
		}
	}
	return nil
}

func payoutWinningPositions(ctx context.Context, repo ResolutionRepository, userService UserService, marketID int64) error {
	positions, err := repo.CalculatePayoutPositions(ctx, marketID)
	if err != nil {
		return err
	}
	for _, pos := range positions {
		if pos.Value <= 0 {
			continue
		}
		if err := userService.ApplyTransaction(ctx, pos.Username, pos.Value, users.TransactionWin); err != nil {
			return err
		}
	}
	return nil
}
