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
	DescriptionAmendmentApprovedByAuto     = "auto-approval"
	MarketProposalApprovedByAuto           = "auto-approval"
	MaxDescriptionAmendmentLength          = 2000
	MaxDescriptionAmendmentReasonLength    = 500
)

var unsafeMarkdownLitePattern = regexp.MustCompile(`(?i)<[^>]+>|javascript:|data:`)

type MarketDescriptionAmendment struct {
	ID                         int64
	MarketID                   int64
	MarketTitle                string
	MarketDescription          string
	MarketGroup                *MarketGroup
	Version                    int
	Body                       string
	BodyFormat                 string
	Status                     string
	CreatedBy                  string
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
	ApprovedBy                 string
	ApprovedAt                 *time.Time
	RejectedBy                 string
	RejectedAt                 *time.Time
	RejectionReason            string
	SubmitReason               string
	PreviousApprovedAmendments []MarketDescriptionAmendment
}

type MarketDescriptionAmendmentRequest struct {
	Body         string
	BodyFormat   string
	SubmitReason string
}

type MarketDescriptionAmendmentFilters struct {
	MarketID  int64
	Status    string
	CreatedBy string
	Limit     int
	Offset    int
}

type MarketDescriptionAmendmentRepository interface {
	CreateMarketDescriptionAmendment(ctx context.Context, amendment MarketDescriptionAmendment) (*MarketDescriptionAmendment, error)
	ListMarketDescriptionAmendments(ctx context.Context, filters MarketDescriptionAmendmentFilters) ([]MarketDescriptionAmendment, error)
	ReviewMarketDescriptionAmendment(ctx context.Context, id int64, status string, actorUsername string, reason string, reviewedAt time.Time) (*MarketDescriptionAmendment, error)
	ReviewGroupedMarketDescriptionAmendments(ctx context.Context, ids []int64, status string, actorUsername string, reason string, reviewedAt time.Time) ([]MarketDescriptionAmendment, error)
}

type MarketGovernanceSettings struct {
	AutoApproveDescriptionAmendments        bool
	AutoApproveMarketProposals              bool
	AutoApproveMarketGroupAnswers           bool
	MarketGroupAnswerAdditionApprovalPolicy string
	Version                                 uint
	UpdatedBy                               string
	UpdatedAt                               time.Time
}

type MarketGovernanceSettingsUpdate struct {
	AutoApproveDescriptionAmendments        *bool
	AutoApproveMarketProposals              *bool
	AutoApproveMarketGroupAnswers           *bool
	MarketGroupAnswerAdditionApprovalPolicy *string
	Version                                 uint
	UpdatedBy                               string
}

type MarketGovernanceSettingsRepository interface {
	GetMarketGovernanceSettings(ctx context.Context) (*MarketGovernanceSettings, error)
	UpdateMarketGovernanceSettings(ctx context.Context, update MarketGovernanceSettingsUpdate) (*MarketGovernanceSettings, error)
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
	now := s.clock.Now()
	if !marketAllowsDescriptionAmendmentProposal(market, now) {
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
	amendmentStatus := DescriptionAmendmentStatusPending
	approvedBy := ""
	var approvedAt *time.Time
	settings, settingsErr := s.GetMarketGovernanceSettings(ctx)
	if settingsErr == nil && settings != nil && settings.AutoApproveDescriptionAmendments {
		amendmentStatus = DescriptionAmendmentStatusApproved
		approvedBy = DescriptionAmendmentApprovedByAuto
		approvedAt = cloneDescriptionAmendmentTime(now)
	}
	return repo.CreateMarketDescriptionAmendment(ctx, MarketDescriptionAmendment{
		MarketID:     marketID,
		Body:         body,
		BodyFormat:   format,
		Status:       amendmentStatus,
		CreatedBy:    actorUsername,
		CreatedAt:    now,
		UpdatedAt:    now,
		ApprovedBy:   approvedBy,
		ApprovedAt:   approvedAt,
		SubmitReason: reason,
	})
}

func (s *Service) ListMarketDescriptionAmendments(ctx context.Context, filters MarketDescriptionAmendmentFilters) ([]MarketDescriptionAmendment, error) {
	filters.Status = NormalizeDescriptionAmendmentStatus(filters.Status)
	filters.CreatedBy = strings.TrimSpace(filters.CreatedBy)
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
	items, err := repo.ListMarketDescriptionAmendments(ctx, filters)
	if err != nil {
		return nil, err
	}
	return s.hydrateDescriptionAmendmentContext(ctx, items), nil
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

func (s *Service) ReviewGroupedMarketDescriptionAmendments(ctx context.Context, amendmentIDs []int64, status string, actorUsername string, reason string) ([]MarketDescriptionAmendment, error) {
	actorUsername = strings.TrimSpace(actorUsername)
	status = NormalizeDescriptionAmendmentStatus(status)
	reason = strings.TrimSpace(reason)
	if !validGroupedDescriptionAmendmentIDs(amendmentIDs) || actorUsername == "" || reason == "" || len([]rune(reason)) > MaxDescriptionAmendmentReasonLength {
		return nil, ErrInvalidInput
	}
	if status != DescriptionAmendmentStatusApproved && status != DescriptionAmendmentStatusRejected {
		return nil, ErrInvalidInput
	}
	repo, err := s.descriptionAmendmentRepository()
	if err != nil {
		return nil, err
	}
	return repo.ReviewGroupedMarketDescriptionAmendments(ctx, amendmentIDs, status, actorUsername, reason, s.clock.Now())
}

func (s *Service) GetMarketGovernanceSettings(ctx context.Context) (*MarketGovernanceSettings, error) {
	repo, err := s.marketGovernanceSettingsRepository()
	if err != nil {
		return nil, err
	}
	return repo.GetMarketGovernanceSettings(ctx)
}

func (s *Service) UpdateMarketGovernanceSettings(ctx context.Context, update MarketGovernanceSettingsUpdate) (*MarketGovernanceSettings, error) {
	update.UpdatedBy = strings.TrimSpace(update.UpdatedBy)
	if update.UpdatedBy == "" ||
		(update.AutoApproveDescriptionAmendments == nil &&
			update.AutoApproveMarketProposals == nil &&
			update.AutoApproveMarketGroupAnswers == nil &&
			update.MarketGroupAnswerAdditionApprovalPolicy == nil) {
		return nil, ErrInvalidInput
	}
	if update.MarketGroupAnswerAdditionApprovalPolicy != nil {
		policy := NormalizeMarketGroupAnswerAdditionApprovalPolicy(*update.MarketGroupAnswerAdditionApprovalPolicy)
		if !IsValidMarketGroupAnswerAdditionApprovalPolicy(policy) {
			return nil, ErrInvalidInput
		}
		update.MarketGroupAnswerAdditionApprovalPolicy = &policy
	}
	repo, err := s.marketGovernanceSettingsRepository()
	if err != nil {
		return nil, err
	}
	return repo.UpdateMarketGovernanceSettings(ctx, update)
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

func validGroupedDescriptionAmendmentIDs(ids []int64) bool {
	if len(ids) == 0 {
		return false
	}
	seen := map[int64]bool{}
	for _, id := range ids {
		if id <= 0 || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}

func (s *Service) marketGovernanceSettingsRepository() (MarketGovernanceSettingsRepository, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	repo, ok := s.repo.(MarketGovernanceSettingsRepository)
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

func (s *Service) hydrateDescriptionAmendmentContext(ctx context.Context, items []MarketDescriptionAmendment) []MarketDescriptionAmendment {
	if len(items) == 0 {
		return items
	}

	marketCache := map[int64]*Market{}
	approvedCache := map[int64][]MarketDescriptionAmendment{}

	for index := range items {
		item := &items[index]
		market, ok := marketCache[item.MarketID]
		if !ok {
			var err error
			market, err = s.GetMarket(ctx, item.MarketID)
			if err != nil {
				market = nil
			}
			marketCache[item.MarketID] = market
		}
		if market != nil {
			item.MarketTitle = market.QuestionTitle
			item.MarketDescription = market.Description
		}
		if group, err := s.GetMarketGroupForMarket(ctx, item.MarketID); err == nil && group != nil {
			item.MarketGroup = group
		}

		approved, ok := approvedCache[item.MarketID]
		if !ok {
			approved = s.approvedDescriptionAmendments(ctx, item.MarketID)
			approvedCache[item.MarketID] = approved
		}
		item.PreviousApprovedAmendments = approvedAmendmentsBeforeVersion(approved, item.Version)
	}

	return items
}

func approvedAmendmentsBeforeVersion(items []MarketDescriptionAmendment, version int) []MarketDescriptionAmendment {
	if len(items) == 0 {
		return []MarketDescriptionAmendment{}
	}
	out := make([]MarketDescriptionAmendment, 0, len(items))
	for _, item := range items {
		if item.Status == DescriptionAmendmentStatusApproved && item.Version < version {
			out = append(out, item)
		}
	}
	return out
}

func cloneDescriptionAmendmentTime(value time.Time) *time.Time {
	cloned := value
	return &cloned
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

func marketAllowsDescriptionAmendmentProposal(market *Market, now time.Time) bool {
	if market == nil || market.IsResolved() {
		return false
	}
	if !market.ResolutionDateTime.IsZero() && !now.Before(market.ResolutionDateTime) {
		return false
	}
	switch NormalizeLifecycleStatus(market.LifecycleStatus) {
	case MarketLifecycleProposed, MarketLifecyclePublished:
		return true
	default:
		return false
	}
}
