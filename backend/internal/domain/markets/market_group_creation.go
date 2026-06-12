package markets

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CreateMarketGroup creates a multiple-choice binary parent and normal binary
// child markets. The parent charges one proposal cost; child proposal costs are
// zero because they are implementation details of the group.
func (s *Service) CreateMarketGroup(ctx context.Context, req MarketGroupCreateRequest, creatorUsername string) (*MarketGroup, error) {
	groupRepo, err := s.marketGroupRepository()
	if err != nil {
		return nil, err
	}
	if err := s.validateMarketGroupCreateRequest(req); err != nil {
		return nil, err
	}

	tagSlugs, err := s.validateCreateTagSlugs(ctx, req.TagSlugs)
	if err != nil {
		return nil, err
	}
	if err := s.userService.ValidateUserExists(ctx, creatorUsername); err != nil {
		return nil, ErrUserNotFound
	}
	if err := s.creationPolicy.ValidateResolutionTime(s.clock.Now(), req.ResolutionDateTime, s.config.MinimumFutureHours); err != nil {
		return nil, err
	}

	now := s.clock.Now()
	lifecycleTemplate := s.creationPolicy.BuildMarketEntity(now, MarketCreateRequest{
		QuestionTitle:      req.QuestionTitle,
		Description:        req.Description,
		OutcomeType:        "BINARY",
		ResolutionDateTime: req.ResolutionDateTime,
	}, creatorUsername, labelPair{yes: "YES", no: "NO"})
	if err := s.applyCreationLifecycle(ctx, lifecycleTemplate, creatorUsername); err != nil {
		return nil, err
	}
	if err := s.creationPolicy.EnsureCreateMarketBalance(ctx, s.userService, creatorUsername, s.config.CreateMarketCost, s.config.MaximumDebtAllowed); err != nil {
		return nil, err
	}

	group := &MarketGroup{
		QuestionTitle:      strings.TrimSpace(req.QuestionTitle),
		Description:        strings.TrimSpace(req.Description),
		GroupType:          MarketGroupTypeMultipleChoiceBinary,
		ProbabilityPolicy:  MarketGroupProbabilityPolicyIndependentBinary,
		ResolutionPolicy:   MarketGroupResolutionPolicyIndependentChildren,
		LifecycleStatus:    lifecycleTemplate.LifecycleStatus,
		ProposalCost:       s.config.CreateMarketCost,
		CreatorUsername:    creatorUsername,
		StewardUsername:    creatorUsername,
		ApprovedBy:         lifecycleTemplate.ApprovedBy,
		ApprovedAt:         copyDomainTimePtr(lifecycleTemplate.ApprovedAt),
		ResolutionDateTime: req.ResolutionDateTime,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	members := make([]MarketGroupMember, 0, len(req.AnswerLabels))
	for index, rawLabel := range req.AnswerLabels {
		label := strings.TrimSpace(rawLabel)
		child := s.creationPolicy.BuildMarketEntity(now, MarketCreateRequest{
			QuestionTitle:      buildMarketGroupChildTitle(req.QuestionTitle, label),
			Description:        buildMarketGroupChildDescription(req.Description, label),
			OutcomeType:        "BINARY",
			ResolutionDateTime: req.ResolutionDateTime,
			YesLabel:           "YES",
			NoLabel:            "NO",
		}, creatorUsername, labelPair{yes: "YES", no: "NO"})
		child.LifecycleStatus = lifecycleTemplate.LifecycleStatus
		child.Status = lifecycleTemplate.Status
		child.ApprovedBy = lifecycleTemplate.ApprovedBy
		child.ApprovedAt = copyDomainTimePtr(lifecycleTemplate.ApprovedAt)
		child.ProposalCost = 0

		if err := s.repo.Create(ctx, child); err != nil {
			return nil, err
		}
		if err := s.assignTagsToMarket(ctx, child, tagSlugs, creatorUsername); err != nil {
			return nil, err
		}

		members = append(members, MarketGroupMember{
			MarketID:     child.ID,
			AnswerLabel:  label,
			DisplayOrder: index,
		})
	}

	if err := groupRepo.CreateMarketGroup(ctx, group, members); err != nil {
		return nil, err
	}
	return group, nil
}

// GetMarketGroupOverview returns parent group display data with child-market
// overviews. Buy/sell/resolve flows must still target child market IDs.
func (s *Service) GetMarketGroupOverview(ctx context.Context, groupID int64) (*MarketGroupOverview, error) {
	if groupID <= 0 {
		return nil, ErrInvalidInput
	}
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

	answers := make([]MarketGroupAnswerOverview, 0, len(group.Members))
	for _, member := range OrderedMarketGroupMembers(group.Members) {
		overview, err := s.GetMarketDetails(ctx, member.MarketID)
		if err != nil {
			return nil, err
		}
		answers = append(answers, MarketGroupAnswerOverview{
			Member:   member,
			Overview: overview,
		})
	}

	return &MarketGroupOverview{
		Group:   group,
		Creator: s.buildCreatorSummary(ctx, group.CreatorUsername),
		Answers: answers,
	}, nil
}

func (s *Service) validateMarketGroupCreateRequest(req MarketGroupCreateRequest) error {
	if len(strings.TrimSpace(req.QuestionTitle)) == 0 || len(req.QuestionTitle) > MaxQuestionTitleLength {
		return ErrInvalidQuestionLength
	}
	if len(req.Description) > MaxDescriptionLength {
		return ErrInvalidDescriptionLength
	}
	members := make([]MarketGroupMember, 0, len(req.AnswerLabels))
	for index, answer := range req.AnswerLabels {
		members = append(members, MarketGroupMember{
			MarketID:     int64(index + 1),
			AnswerLabel:  answer,
			DisplayOrder: index,
		})
	}
	return ValidateMarketGroupMembers(members)
}

func (s *Service) marketGroupRepository() (MarketGroupRepository, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidState
	}
	repo, ok := s.repo.(MarketGroupRepository)
	if !ok {
		return nil, ErrInvalidState
	}
	return repo, nil
}

func buildMarketGroupChildTitle(parentTitle, answerLabel string) string {
	parentTitle = strings.TrimSpace(parentTitle)
	answerLabel = strings.TrimSpace(answerLabel)
	candidate := fmt.Sprintf("%s - %s", parentTitle, answerLabel)
	if runeLen(candidate) <= MaxQuestionTitleLength {
		return candidate
	}
	if runeLen(answerLabel) >= MaxQuestionTitleLength {
		return truncateRunes(answerLabel, MaxQuestionTitleLength)
	}
	availableParent := MaxQuestionTitleLength - runeLen(answerLabel) - runeLen(" - ")
	if availableParent < 1 {
		return answerLabel
	}
	return fmt.Sprintf("%s - %s", strings.TrimSpace(truncateRunes(parentTitle, availableParent)), answerLabel)
}

func buildMarketGroupChildDescription(parentDescription, answerLabel string) string {
	answerLine := fmt.Sprintf("Answer choice: %s", strings.TrimSpace(answerLabel))
	note := "This is a binary child market in a multiple-choice binary market group. Each answer is traded independently as its own YES/NO market."
	description := strings.TrimSpace(parentDescription)
	if description == "" {
		return truncateRunes(answerLine+"\n\n"+note, MaxDescriptionLength)
	}
	suffix := "\n\n" + answerLine + "\n\n" + note
	if runeLen(description)+runeLen(suffix) <= MaxDescriptionLength {
		return description + suffix
	}
	availableDescription := MaxDescriptionLength - runeLen(suffix)
	if availableDescription <= 0 {
		return truncateRunes(strings.TrimSpace(answerLine+"\n\n"+note), MaxDescriptionLength)
	}
	return strings.TrimSpace(truncateRunes(description, availableDescription)) + suffix
}

func copyDomainTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func runeLen(value string) int {
	return len([]rune(value))
}

func truncateRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
