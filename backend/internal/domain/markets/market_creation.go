package markets

import (
	"context"
	"time"
)

type labelPair struct {
	yes string
	no  string
}

// CreateMarket creates a new market with validation.
func (s *Service) CreateMarket(ctx context.Context, req MarketCreateRequest, creatorUsername string) (*Market, error) {
	if err := s.creationPolicy.ValidateCreateRequest(req); err != nil {
		return nil, err
	}

	labels := s.creationPolicy.NormalizeLabels(req.YesLabel, req.NoLabel)

	if err := s.userService.ValidateUserExists(ctx, creatorUsername); err != nil {
		return nil, ErrUserNotFound
	}

	if err := s.creationPolicy.ValidateResolutionTime(s.clock.Now(), req.ResolutionDateTime, s.config.MinimumFutureHours); err != nil {
		return nil, err
	}

	if err := s.creationPolicy.EnsureCreateMarketBalance(ctx, s.userService, creatorUsername, s.config.CreateMarketCost, s.config.MaximumDebtAllowed); err != nil {
		return nil, err
	}

	now := s.clock.Now()
	market := s.creationPolicy.BuildMarketEntity(now, req, creatorUsername, labels)

	if err := s.repo.Create(ctx, market); err != nil {
		return nil, err
	}

	return market, nil
}

// SetCustomLabels updates the custom labels for a market.
func (s *Service) SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error {
	if err := s.creationPolicy.ValidateCustomLabels(yesLabel, noLabel); err != nil {
		return err
	}

	if _, err := s.repo.GetByID(ctx, marketID); err != nil {
		return ErrMarketNotFound
	}

	return s.repo.UpdateLabels(ctx, marketID, yesLabel, noLabel)
}

// GetMarket retrieves a market by ID.
func (s *Service) GetMarket(ctx context.Context, id int64) (*Market, error) {
	return s.repo.GetByID(ctx, id)
}

// ValidateQuestionTitle validates the market question title.
func (s *Service) ValidateQuestionTitle(title string) error {
	if len(title) > MaxQuestionTitleLength || len(title) < 1 {
		return ErrInvalidQuestionLength
	}
	return nil
}

// ValidateDescription validates the market description.
func (s *Service) ValidateDescription(description string) error {
	if len(description) > MaxDescriptionLength {
		return ErrInvalidDescriptionLength
	}
	return nil
}

// ValidateLabels validates the custom yes/no labels.
func (s *Service) ValidateLabels(yesLabel, noLabel string) error {
	return s.creationPolicy.ValidateCustomLabels(yesLabel, noLabel)
}

// ValidateMarketResolutionTime validates that the market resolution time meets business logic requirements.
func (s *Service) ValidateMarketResolutionTime(resolutionTime time.Time) error {
	return s.creationPolicy.ValidateResolutionTime(s.clock.Now(), resolutionTime, s.config.MinimumFutureHours)
}
