package markets

import (
	"context"
	"fmt"
	"strings"
	"time"

	users "socialpredict/internal/domain/users"
)

const (
	MarketGroupAnswerAdditionStatusPending  = "pending"
	MarketGroupAnswerAdditionStatusApproved = "approved"
	MarketGroupAnswerAdditionStatusRejected = "rejected"

	MarketGroupAnswerAdditionApprovedByAuto = "auto-approval"
)

type MarketGroupAnswerAddition struct {
	ID              int64
	GroupID         int64
	MarketID        int64
	GroupTitle      string
	AnswerLabel     string
	Status          string
	ProposedBy      string
	ReviewedBy      string
	ReviewedAt      *time.Time
	RejectionReason string
	AdditionCost    int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	MarketGroup     *MarketGroup
}

type MarketGroupAnswerAdditionRequest struct {
	AnswerLabel string
}

type MarketGroupAnswerAdditionFilters struct {
	GroupID    int64
	Status     string
	ProposedBy string
	Limit      int
	Offset     int
}

type MarketGroupAnswerAdditionRepository interface {
	CreateMarketGroupAnswerAddition(ctx context.Context, addition MarketGroupAnswerAddition) (*MarketGroupAnswerAddition, error)
	GetMarketGroupAnswerAddition(ctx context.Context, id int64) (*MarketGroupAnswerAddition, error)
	ListMarketGroupAnswerAdditions(ctx context.Context, filters MarketGroupAnswerAdditionFilters) ([]MarketGroupAnswerAddition, error)
	ReviewMarketGroupAnswerAddition(ctx context.Context, id int64, status string, marketID int64, actorUsername string, reason string, reviewedAt time.Time) (*MarketGroupAnswerAddition, error)
	AddMarketGroupMember(ctx context.Context, groupID int64, member MarketGroupMember) (*MarketGroupMember, error)
}

func NormalizeMarketGroupAnswerAdditionStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", MarketGroupAnswerAdditionStatusPending:
		return MarketGroupAnswerAdditionStatusPending
	case MarketGroupAnswerAdditionStatusApproved:
		return MarketGroupAnswerAdditionStatusApproved
	case MarketGroupAnswerAdditionStatusRejected:
		return MarketGroupAnswerAdditionStatusRejected
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func (s *Service) ProposeMarketGroupAnswerAddition(ctx context.Context, groupID int64, actorUsername string, req MarketGroupAnswerAdditionRequest) (*MarketGroupAnswerAddition, error) {
	actorUsername = strings.TrimSpace(actorUsername)
	if groupID <= 0 || actorUsername == "" {
		return nil, ErrInvalidInput
	}
	label, err := normalizeMarketGroupAnswerAdditionLabel(req.AnswerLabel)
	if err != nil {
		return nil, err
	}
	if err := s.ensureActiveModerator(ctx, actorUsername); err != nil {
		return nil, err
	}
	group, err := s.validateGroupAllowsAnswerAddition(ctx, groupID, label, 0)
	if err != nil {
		return nil, err
	}
	repo, err := s.marketGroupAnswerAdditionRepository()
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()
	addition, err := repo.CreateMarketGroupAnswerAddition(ctx, MarketGroupAnswerAddition{
		GroupID:      group.ID,
		GroupTitle:   group.QuestionTitle,
		AnswerLabel:  label,
		Status:       MarketGroupAnswerAdditionStatusPending,
		ProposedBy:   actorUsername,
		AdditionCost: s.marketGroupAddAnswerCost(),
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return nil, err
	}

	if s.autoApproveMarketGroupAnswersEnabled(ctx) {
		return s.ApproveMarketGroupAnswerAddition(ctx, addition.ID, MarketGroupAnswerAdditionApprovedByAuto, true)
	}
	return addition, nil
}

func (s *Service) ListMarketGroupAnswerAdditions(ctx context.Context, filters MarketGroupAnswerAdditionFilters) ([]MarketGroupAnswerAddition, error) {
	filters.Status = NormalizeMarketGroupAnswerAdditionStatus(filters.Status)
	filters.ProposedBy = strings.TrimSpace(filters.ProposedBy)
	if filters.Status != MarketGroupAnswerAdditionStatusPending &&
		filters.Status != MarketGroupAnswerAdditionStatusApproved &&
		filters.Status != MarketGroupAnswerAdditionStatusRejected {
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
	repo, err := s.marketGroupAnswerAdditionRepository()
	if err != nil {
		return nil, err
	}
	return repo.ListMarketGroupAnswerAdditions(ctx, filters)
}

func (s *Service) ApproveMarketGroupAnswerAddition(ctx context.Context, additionID int64, actorUsername string, confirmed bool) (*MarketGroupAnswerAddition, error) {
	actorUsername = strings.TrimSpace(actorUsername)
	if additionID <= 0 || actorUsername == "" || !confirmed {
		return nil, ErrInvalidInput
	}
	repo, err := s.marketGroupAnswerAdditionRepository()
	if err != nil {
		return nil, err
	}
	addition, err := repo.GetMarketGroupAnswerAddition(ctx, additionID)
	if err != nil {
		return nil, err
	}
	if addition == nil {
		return nil, ErrMarketGroupNotFound
	}
	if NormalizeMarketGroupAnswerAdditionStatus(addition.Status) != MarketGroupAnswerAdditionStatusPending {
		return nil, ErrInvalidState
	}

	group, err := s.validateGroupAllowsAnswerAddition(ctx, addition.GroupID, addition.AnswerLabel, addition.ID)
	if err != nil {
		return nil, err
	}
	cost := addition.AdditionCost
	if cost < 0 {
		cost = 0
	}
	if cost > 0 {
		if s.userService == nil {
			return nil, ErrUnauthorized
		}
		if err := s.userService.ValidateUserBalance(ctx, addition.ProposedBy, cost, s.config.MaximumDebtAllowed); err != nil {
			return nil, ErrInsufficientBalance
		}
		if err := s.userService.ApplyTransaction(ctx, addition.ProposedBy, cost, users.TransactionFee); err != nil {
			return nil, err
		}
	}

	now := s.clock.Now()
	child := s.buildApprovedMarketGroupAnswerChild(group, addition.AnswerLabel, now, actorUsername)
	if err := s.repo.Create(ctx, child); err != nil {
		return nil, err
	}
	if err := s.assignTagsToMarket(ctx, child, s.groupTagSlugs(ctx, group), addition.ProposedBy); err != nil {
		return nil, err
	}
	member, err := repo.AddMarketGroupMember(ctx, group.ID, MarketGroupMember{
		MarketID:     child.ID,
		AnswerLabel:  addition.AnswerLabel,
		DisplayOrder: len(group.Members),
	})
	if err != nil {
		return nil, err
	}
	reviewed, err := repo.ReviewMarketGroupAnswerAddition(ctx, additionID, MarketGroupAnswerAdditionStatusApproved, child.ID, actorUsername, "", now)
	if err != nil {
		return nil, err
	}
	group.Members = append(group.Members, *member)
	if err := s.createAnswerAdditionAmendments(ctx, group, addition, now, actorUsername); err != nil {
		return nil, err
	}
	return reviewed, nil
}

func (s *Service) RejectMarketGroupAnswerAddition(ctx context.Context, additionID int64, actorUsername string, reason string) (*MarketGroupAnswerAddition, error) {
	actorUsername = strings.TrimSpace(actorUsername)
	reason = strings.TrimSpace(reason)
	if additionID <= 0 || actorUsername == "" || reason == "" || len([]rune(reason)) > MaxDescriptionAmendmentReasonLength {
		return nil, ErrInvalidInput
	}
	repo, err := s.marketGroupAnswerAdditionRepository()
	if err != nil {
		return nil, err
	}
	return repo.ReviewMarketGroupAnswerAddition(ctx, additionID, MarketGroupAnswerAdditionStatusRejected, 0, actorUsername, reason, s.clock.Now())
}

func (s *Service) validateGroupAllowsAnswerAddition(ctx context.Context, groupID int64, label string, ignoredAdditionID int64) (*MarketGroup, error) {
	groupRepo, err := s.marketGroupRepository()
	if err != nil {
		return nil, err
	}
	group, err := groupRepo.GetMarketGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrMarketGroupNotFound
	}
	if NormalizeLifecycleStatus(group.LifecycleStatus) != MarketLifecyclePublished {
		return nil, ErrInvalidState
	}
	if !group.ResolutionDateTime.IsZero() && !group.ResolutionDateTime.After(s.clock.Now()) {
		return nil, ErrInvalidState
	}
	if len(group.Members)+1 > s.multipleChoiceBinaryHardAnswerSafetyCap() {
		return nil, ErrInvalidInput
	}
	if answerLabelExists(group.Members, label) {
		return nil, ErrInvalidInput
	}
	if pending, err := s.pendingAnswerLabelExists(ctx, groupID, label, ignoredAdditionID); err != nil {
		return nil, err
	} else if pending {
		return nil, ErrInvalidInput
	}
	return group, nil
}

func (s *Service) pendingAnswerLabelExists(ctx context.Context, groupID int64, label string, ignoredAdditionID int64) (bool, error) {
	repo, err := s.marketGroupAnswerAdditionRepository()
	if err != nil {
		return false, err
	}
	items, err := repo.ListMarketGroupAnswerAdditions(ctx, MarketGroupAnswerAdditionFilters{
		GroupID: groupID,
		Status:  MarketGroupAnswerAdditionStatusPending,
		Limit:   200,
	})
	if err != nil {
		return false, err
	}
	for _, item := range items {
		if ignoredAdditionID > 0 && item.ID == ignoredAdditionID {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(item.AnswerLabel), strings.TrimSpace(label)) {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) buildApprovedMarketGroupAnswerChild(group *MarketGroup, answerLabel string, now time.Time, approvedBy string) *Market {
	child := s.creationPolicy.BuildMarketEntity(now, MarketCreateRequest{
		QuestionTitle:      buildMarketGroupChildTitle(group.QuestionTitle, answerLabel),
		Description:        buildMarketGroupChildDescription(group.Description, answerLabel),
		OutcomeType:        "BINARY",
		ResolutionDateTime: group.ResolutionDateTime,
		YesLabel:           "YES",
		NoLabel:            "NO",
	}, group.CreatorUsername, labelPair{yes: "YES", no: "NO"})
	child.LifecycleStatus = MarketLifecyclePublished
	child.Status = MarketStatusActive
	child.ApprovedBy = approvedBy
	child.ApprovedAt = cloneDescriptionAmendmentTime(now)
	child.ProposalCost = 0
	child.StewardUsername = group.CurrentStewardUsername()
	return child
}

func (s *Service) createAnswerAdditionAmendments(ctx context.Context, group *MarketGroup, addition *MarketGroupAnswerAddition, addedAt time.Time, approvedBy string) error {
	repo, err := s.descriptionAmendmentRepository()
	if err != nil {
		return err
	}
	body := buildAnswerAdditionAmendmentBody(addition.AnswerLabel, addition.ProposedBy, addedAt)
	reason := fmt.Sprintf("Answer option %q was added to grouped market #%d.", addition.AnswerLabel, group.ID)
	for _, member := range OrderedMarketGroupMembers(group.Members) {
		if member.MarketID <= 0 {
			continue
		}
		if _, err := repo.CreateMarketDescriptionAmendment(ctx, MarketDescriptionAmendment{
			MarketID:     member.MarketID,
			Body:         body,
			BodyFormat:   DescriptionAmendmentFormatMarkdownLite,
			Status:       DescriptionAmendmentStatusApproved,
			CreatedBy:    addition.ProposedBy,
			CreatedAt:    addedAt,
			UpdatedAt:    addedAt,
			ApprovedBy:   approvedBy,
			ApprovedAt:   cloneDescriptionAmendmentTime(addedAt),
			SubmitReason: reason,
		}); err != nil {
			return err
		}
	}
	return nil
}

func buildAnswerAdditionAmendmentBody(answerLabel string, proposedBy string, addedAt time.Time) string {
	return fmt.Sprintf(
		"Added answer option **%s** on %s by [@%s](/user/%s).",
		strings.TrimSpace(answerLabel),
		addedAt.UTC().Format(time.RFC3339),
		strings.TrimSpace(proposedBy),
		strings.TrimSpace(proposedBy),
	)
}

func (s *Service) groupTagSlugs(ctx context.Context, group *MarketGroup) []string {
	if group == nil || len(group.Members) == 0 {
		return []string{}
	}
	first, err := s.GetMarket(ctx, OrderedMarketGroupMembers(group.Members)[0].MarketID)
	if err != nil || first == nil || len(first.Tags) == 0 {
		return []string{}
	}
	slugs := make([]string, 0, len(first.Tags))
	for _, tag := range first.Tags {
		if tag.Slug != "" && tag.IsActive {
			slugs = append(slugs, tag.Slug)
		}
	}
	return slugs
}

func (s *Service) marketGroupAddAnswerCost() int64 {
	if s == nil || s.config.MultipleChoiceBinaryAddAnswerCost <= 0 {
		return 0
	}
	return s.config.MultipleChoiceBinaryAddAnswerCost
}

func (s *Service) autoApproveMarketGroupAnswersEnabled(ctx context.Context) bool {
	settings, err := s.GetMarketGovernanceSettings(ctx)
	if err != nil || settings == nil {
		return false
	}
	return settings.AutoApproveMarketGroupAnswers
}

func (s *Service) marketGroupAnswerAdditionRepository() (MarketGroupAnswerAdditionRepository, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	repo, ok := s.repo.(MarketGroupAnswerAdditionRepository)
	if !ok {
		return nil, ErrInvalidInput
	}
	return repo, nil
}

func normalizeMarketGroupAnswerAdditionLabel(value string) (string, error) {
	label := strings.TrimSpace(value)
	if label == "" || len([]rune(label)) > MaxAnswerLabelLength {
		return "", ErrInvalidInput
	}
	return label, nil
}

func answerLabelExists(members []MarketGroupMember, label string) bool {
	for _, member := range members {
		if strings.EqualFold(strings.TrimSpace(member.AnswerLabel), strings.TrimSpace(label)) {
			return true
		}
	}
	return false
}
