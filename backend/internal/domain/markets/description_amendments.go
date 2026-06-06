package markets

import (
	"context"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	users "socialpredict/internal/domain/users"
)

const (
	DescriptionAmendmentFormatMarkdownLite = "markdown_lite"
	DescriptionAmendmentStatusPending      = "pending"
	DescriptionAmendmentStatusApproved     = "approved"
	DescriptionAmendmentStatusRejected     = "rejected"
	MaxDescriptionAmendmentLength          = 2000
	MaxDescriptionAmendmentReasonLength    = 500
)

var unsafeMarkdownLitePattern = regexp.MustCompile(`(?i)<[^>]+>|javascript:|data:`)

type MarketDescriptionAmendment struct {
	ID              int64
	MarketID        int64
	Version         int
	Body            string
	BodyFormat      string
	Status          string
	CreatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ApprovedBy      string
	ApprovedAt      *time.Time
	RejectedBy      string
	RejectedAt      *time.Time
	RejectionReason string
	SubmitReason    string
}

type MarketDescriptionAmendmentRequest struct {
	Body         string
	BodyFormat   string
	SubmitReason string
}

type MarketDescriptionAmendmentFilters struct {
	MarketID int64
	Status   string
	Limit    int
	Offset   int
}

type MarketDescriptionAmendmentRepository interface {
	CreateMarketDescriptionAmendment(ctx context.Context, amendment MarketDescriptionAmendment) (*MarketDescriptionAmendment, error)
	ListMarketDescriptionAmendments(ctx context.Context, filters MarketDescriptionAmendmentFilters) ([]MarketDescriptionAmendment, error)
	ReviewMarketDescriptionAmendment(ctx context.Context, id int64, status string, actorUsername string, reason string, reviewedAt time.Time) (*MarketDescriptionAmendment, error)
}

func NormalizeDescriptionAmendmentStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case DescriptionAmendmentStatusPending, "":
		return DescriptionAmendmentStatusPending
	case DescriptionAmendmentStatusApproved:
		return DescriptionAmendmentStatusApproved
	case DescriptionAmendmentStatusRejected:
		return DescriptionAmendmentStatusRejected
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func NormalizeDescriptionAmendmentFormat(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", DescriptionAmendmentFormatMarkdownLite, "markdown":
		return DescriptionAmendmentFormatMarkdownLite
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func (s *Service) ProposeMarketDescriptionAmendment(ctx context.Context, marketID int64, actorUsername string, req MarketDescriptionAmendmentRequest) (*MarketDescriptionAmendment, error) {
	actorUsername = strings.TrimSpace(actorUsername)
	if marketID <= 0 || actorUsername == "" {
		return nil, ErrInvalidInput
	}
	body, format, reason, err := validateDescriptionAmendmentInput(req)
	if err != nil {
		return nil, err
	}

	market, err := s.GetMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if !marketAllowsDescriptionAmendmentProposal(market) {
		return nil, ErrInvalidState
	}
	if !market.StewardedBy(actorUsername) {
		return nil, ErrUnauthorized
	}
	if err := s.ensureActiveModerator(ctx, actorUsername); err != nil {
		return nil, err
	}

	repo, err := s.descriptionAmendmentRepository()
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	return repo.CreateMarketDescriptionAmendment(ctx, MarketDescriptionAmendment{
		MarketID:     marketID,
		Body:         body,
		BodyFormat:   format,
		Status:       DescriptionAmendmentStatusPending,
		CreatedBy:    actorUsername,
		CreatedAt:    now,
		UpdatedAt:    now,
		SubmitReason: reason,
	})
}

func (s *Service) ListMarketDescriptionAmendments(ctx context.Context, filters MarketDescriptionAmendmentFilters) ([]MarketDescriptionAmendment, error) {
	filters.Status = NormalizeDescriptionAmendmentStatus(filters.Status)
	if filters.Status != DescriptionAmendmentStatusPending && filters.Status != DescriptionAmendmentStatusApproved && filters.Status != DescriptionAmendmentStatusRejected {
		return nil, ErrInvalidInput
	}
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 200 {
		filters.Limit = 200
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	repo, err := s.descriptionAmendmentRepository()
	if err != nil {
		return nil, err
	}
	return repo.ListMarketDescriptionAmendments(ctx, filters)
}

func (s *Service) ReviewMarketDescriptionAmendment(ctx context.Context, amendmentID int64, status string, actorUsername string, reason string) (*MarketDescriptionAmendment, error) {
	actorUsername = strings.TrimSpace(actorUsername)
	status = NormalizeDescriptionAmendmentStatus(status)
	reason = strings.TrimSpace(reason)
	if amendmentID <= 0 || actorUsername == "" || reason == "" || len([]rune(reason)) > MaxDescriptionAmendmentReasonLength {
		return nil, ErrInvalidInput
	}
	if status != DescriptionAmendmentStatusApproved && status != DescriptionAmendmentStatusRejected {
		return nil, ErrInvalidInput
	}
	repo, err := s.descriptionAmendmentRepository()
	if err != nil {
		return nil, err
	}
	return repo.ReviewMarketDescriptionAmendment(ctx, amendmentID, status, actorUsername, reason, s.clock.Now())
}

func (s *Service) descriptionAmendmentRepository() (MarketDescriptionAmendmentRepository, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	repo, ok := s.repo.(MarketDescriptionAmendmentRepository)
	if !ok {
		return nil, ErrInvalidInput
	}
	return repo, nil
}

func (s *Service) approvedDescriptionAmendments(ctx context.Context, marketID int64) []MarketDescriptionAmendment {
	repo, err := s.descriptionAmendmentRepository()
	if err != nil {
		return []MarketDescriptionAmendment{}
	}
	items, err := repo.ListMarketDescriptionAmendments(ctx, MarketDescriptionAmendmentFilters{
		MarketID: marketID,
		Status:   DescriptionAmendmentStatusApproved,
		Limit:    100,
	})
	if err != nil {
		return []MarketDescriptionAmendment{}
	}
	return items
}

func (s *Service) ensureActiveModerator(ctx context.Context, username string) error {
	if s.userService == nil {
		return ErrUnauthorized
	}
	user, err := s.userService.GetPublicUser(ctx, username)
	if err != nil || user == nil {
		return ErrUserNotFound
	}
	if users.NormalizeUserType(user.UserType) != users.UserTypeModerator {
		return ErrUnauthorized
	}
	if users.NormalizeModeratorStatus(user.UserType, string(user.ModeratorStatus)) != users.ModeratorStatusActive {
		return ErrUnauthorized
	}
	return nil
}

func validateDescriptionAmendmentInput(req MarketDescriptionAmendmentRequest) (string, string, string, error) {
	body := strings.TrimSpace(req.Body)
	format := NormalizeDescriptionAmendmentFormat(req.BodyFormat)
	reason := strings.TrimSpace(req.SubmitReason)
	if body == "" || !utf8.ValidString(body) || len([]rune(body)) > MaxDescriptionAmendmentLength {
		return "", "", "", ErrInvalidInput
	}
	if format != DescriptionAmendmentFormatMarkdownLite {
		return "", "", "", ErrInvalidInput
	}
	if unsafeMarkdownLitePattern.MatchString(body) {
		return "", "", "", ErrInvalidInput
	}
	if len([]rune(reason)) > MaxDescriptionAmendmentReasonLength {
		return "", "", "", ErrInvalidInput
	}
	return body, format, reason, nil
}

func marketAllowsDescriptionAmendmentProposal(market *Market) bool {
	if market == nil || market.IsResolved() {
		return false
	}
	switch NormalizeLifecycleStatus(market.LifecycleStatus) {
	case MarketLifecycleProposed, MarketLifecyclePublished:
		return true
	default:
		return false
	}
}
