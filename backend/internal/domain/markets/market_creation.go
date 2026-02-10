package markets

import (
	"context"
	"errors"
	"strings"
	"time"

	dwallet "socialpredict/internal/domain/wallet"
)

type labelPair struct {
	yes string
	no  string
}

// CreateMarket creates a new market with validation.
func (s *Service) CreateMarket(ctx context.Context, req MarketCreateRequest, creatorUsername string) (*Market, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	labels := normalizeLabels(req.YesLabel, req.NoLabel)

	if err := s.creatorProfileService.ValidateUserExists(ctx, creatorUsername); err != nil {
		return nil, ErrUserNotFound
	}

	if err := s.ValidateMarketResolutionTime(req.ResolutionDateTime); err != nil {
		return nil, err
	}

	if err := s.ensureCreateMarketBalance(ctx, creatorUsername); err != nil {
		return nil, err
	}

	market := s.buildMarketEntity(req, creatorUsername, labels)

	if err := s.repo.Create(ctx, market); err != nil {
		return nil, err
	}

	return market, nil
}

// SetCustomLabels updates the custom labels for a market.
func (s *Service) SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error {
	if err := s.validateCustomLabels(yesLabel, noLabel); err != nil {
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

func (s *Service) validateCreateRequest(req MarketCreateRequest) error {
	if err := s.validateQuestionTitle(req.QuestionTitle); err != nil {
		return err
	}
	if err := s.validateDescription(req.Description); err != nil {
		return err
	}
	return s.validateCustomLabels(req.YesLabel, req.NoLabel)
}

func normalizeLabels(yesLabel string, noLabel string) labelPair {
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

func (s *Service) ensureCreateMarketBalance(ctx context.Context, creatorUsername string) error {
	if err := s.walletService.Debit(ctx, creatorUsername, s.config.CreateMarketCost, s.config.MaximumDebtAllowed, dwallet.TxFee); err != nil {
		if errors.Is(err, dwallet.ErrInsufficientBalance) {
			return ErrInsufficientBalance
		}
		return err
	}
	return nil
}

func (s *Service) buildMarketEntity(req MarketCreateRequest, creatorUsername string, labels labelPair) *Market {
	now := s.clock.Now()
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

func (s *Service) validateQuestionTitle(title string) error {
	if len(title) > MaxQuestionTitleLength || len(title) < 1 {
		return ErrInvalidQuestionLength
	}
	return nil
}

func (s *Service) validateDescription(description string) error {
	if len(description) > MaxDescriptionLength {
		return ErrInvalidDescriptionLength
	}
	return nil
}

func (s *Service) validateCustomLabels(yesLabel, noLabel string) error {
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

// ValidateQuestionTitle validates the market question title.
func (s *Service) ValidateQuestionTitle(title string) error {
	return s.validateQuestionTitle(title)
}

// ValidateDescription validates the market description.
func (s *Service) ValidateDescription(description string) error {
	return s.validateDescription(description)
}

// ValidateLabels validates the custom yes/no labels.
func (s *Service) ValidateLabels(yesLabel, noLabel string) error {
	return s.validateCustomLabels(yesLabel, noLabel)
}

// ValidateMarketResolutionTime validates that the market resolution time meets business logic requirements.
func (s *Service) ValidateMarketResolutionTime(resolutionTime time.Time) error {
	now := s.clock.Now()
	minimumDuration := time.Duration(s.config.MinimumFutureHours * float64(time.Hour))
	minimumFutureTime := now.Add(minimumDuration)

	if resolutionTime.Before(minimumFutureTime) || resolutionTime.Equal(minimumFutureTime) {
		return ErrInvalidResolutionTime
	}
	return nil
}
