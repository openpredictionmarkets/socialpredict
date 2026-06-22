package markets

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	MaxMarketTagSlugLength          = 64
	MaxMarketTagDisplayNameLength   = 120
	MaxMarketTagDescriptionLength   = 500
	MaxMarketTagColorKeyLength      = 40
	MaxMarketTagsPerMarket          = 5
	MarketTagAssignmentSourceCreate = "moderator_create"
	MarketTagAssignmentSourceAdmin  = "admin_adjust"
)

var (
	marketTagSlugPattern     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	marketTagColorKeyPattern = regexp.MustCompile(`^[a-z0-9_-]*$`)
)

// MarketTag is an admin-managed taxonomy label that can be attached to markets.
type MarketTag struct {
	ID          int64
	Slug        string
	DisplayName string
	Description string
	ColorKey    string
	SortOrder   int
	IsActive    bool
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// MarketTagRequest captures create/update input for a tag.
type MarketTagRequest struct {
	Slug        string
	DisplayName string
	Description string
	ColorKey    string
	SortOrder   int
	IsActive    *bool
}

// MarketTagRepository persists market taxonomy tags and assignments.
type MarketTagRepository interface {
	ListMarketTags(ctx context.Context, includeInactive bool) ([]MarketTag, error)
	CreateMarketTag(ctx context.Context, tag MarketTag) (*MarketTag, error)
	UpdateMarketTag(ctx context.Context, slug string, update MarketTagRequest) (*MarketTag, error)
	SetMarketTags(ctx context.Context, marketID int64, tagSlugs []string, assignedBy string, source string, assignedAt time.Time) ([]MarketTag, error)
	SetMarketGroupTags(ctx context.Context, groupID int64, tagSlugs []string, assignedBy string, source string, assignedAt time.Time) ([]MarketTag, error)
}

func (s *Service) ListMarketTags(ctx context.Context, includeInactive bool) ([]MarketTag, error) {
	repo, err := s.marketTagRepository()
	if err != nil {
		return nil, err
	}
	return repo.ListMarketTags(ctx, includeInactive)
}

func (s *Service) CreateMarketTag(ctx context.Context, req MarketTagRequest, actorUsername string) (*MarketTag, error) {
	tag, err := normalizeMarketTagRequest(req, actorUsername)
	if err != nil {
		return nil, err
	}
	repo, err := s.marketTagRepository()
	if err != nil {
		return nil, err
	}
	return repo.CreateMarketTag(ctx, tag)
}

func (s *Service) UpdateMarketTag(ctx context.Context, slug string, req MarketTagRequest) (*MarketTag, error) {
	slug = NormalizeMarketTagSlug(slug)
	if slug == "" {
		return nil, ErrInvalidInput
	}
	if err := validateMarketTagRequest(req, false); err != nil {
		return nil, err
	}
	req.Slug = NormalizeMarketTagSlug(req.Slug)
	if req.Slug != "" && req.Slug != slug {
		return nil, ErrInvalidInput
	}
	req.Description = strings.TrimSpace(req.Description)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.ColorKey = strings.TrimSpace(strings.ToLower(req.ColorKey))

	repo, err := s.marketTagRepository()
	if err != nil {
		return nil, err
	}
	return repo.UpdateMarketTag(ctx, slug, req)
}

func (s *Service) UpdateMarketTags(ctx context.Context, marketID int64, tagSlugs []string, actorUsername string) (*Market, error) {
	if marketID <= 0 || strings.TrimSpace(actorUsername) == "" {
		return nil, ErrInvalidInput
	}

	market, err := s.GetMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}

	switch NormalizeLifecycleStatus(market.LifecycleStatus) {
	case MarketLifecycleProposed, MarketLifecyclePublished:
	default:
		return nil, ErrInvalidState
	}

	normalized, err := s.validateCreateTagSlugs(ctx, tagSlugs)
	if err != nil {
		return nil, err
	}
	repo, err := s.marketTagRepository()
	if err != nil {
		return nil, err
	}
	tags, err := repo.SetMarketTags(ctx, marketID, normalized, actorUsername, MarketTagAssignmentSourceAdmin, s.clock.Now())
	if err != nil {
		return nil, err
	}
	market.Tags = tags
	return market, nil
}

func (s *Service) UpdateMarketGroupTags(ctx context.Context, groupID int64, tagSlugs []string, actorUsername string) (*AdminMarketReviewRow, error) {
	if groupID <= 0 || strings.TrimSpace(actorUsername) == "" {
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
	switch NormalizeLifecycleStatus(group.LifecycleStatus) {
	case MarketLifecycleProposed, MarketLifecyclePublished:
	default:
		return nil, ErrInvalidState
	}

	normalized, err := s.validateCreateTagSlugs(ctx, tagSlugs)
	if err != nil {
		return nil, err
	}
	tagRepo, err := s.marketTagRepository()
	if err != nil {
		return nil, err
	}
	tags, err := tagRepo.SetMarketGroupTags(ctx, groupID, normalized, actorUsername, MarketTagAssignmentSourceAdmin, s.clock.Now())
	if err != nil {
		return nil, err
	}

	children := make([]*Market, 0, len(group.Members))
	for _, member := range OrderedMarketGroupMembers(group.Members) {
		child, err := s.GetMarket(ctx, member.MarketID)
		if err != nil {
			return nil, err
		}
		child.Tags = tags
		children = append(children, child)
	}
	return &AdminMarketReviewRow{
		RowKey:        "group:" + strconv.FormatInt(group.ID, 10),
		IsMarketGroup: true,
		Market:        firstMarket(children),
		Group:         group,
		Children:      children,
		Tags:          tags,
	}, nil
}

func (s *Service) assignTagsToMarket(ctx context.Context, market *Market, tagSlugs []string, assignedBy string) error {
	normalized, err := NormalizeMarketTagSlugs(tagSlugs)
	if err != nil {
		return err
	}
	if len(normalized) == 0 {
		market.Tags = []MarketTag{}
		return nil
	}
	repo, err := s.marketTagRepository()
	if err != nil {
		return err
	}
	tags, err := repo.SetMarketTags(ctx, market.ID, normalized, assignedBy, MarketTagAssignmentSourceCreate, s.clock.Now())
	if err != nil {
		return err
	}
	market.Tags = tags
	return nil
}

func firstMarket(markets []*Market) *Market {
	for _, market := range markets {
		if market != nil {
			return market
		}
	}
	return nil
}

func (s *Service) validateCreateTagSlugs(ctx context.Context, tagSlugs []string) ([]string, error) {
	normalized, err := NormalizeMarketTagSlugs(tagSlugs)
	if err != nil {
		return nil, err
	}
	if len(normalized) == 0 {
		return []string{}, nil
	}
	repo, err := s.marketTagRepository()
	if err != nil {
		return nil, err
	}
	activeTags, err := repo.ListMarketTags(ctx, false)
	if err != nil {
		return nil, err
	}
	activeBySlug := make(map[string]bool, len(activeTags))
	for _, tag := range activeTags {
		activeBySlug[tag.Slug] = tag.IsActive
	}
	for _, slug := range normalized {
		if !activeBySlug[slug] {
			return nil, ErrInvalidInput
		}
	}
	return normalized, nil
}

func (s *Service) marketTagRepository() (MarketTagRepository, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	repo, ok := s.repo.(MarketTagRepository)
	if !ok {
		return nil, ErrInvalidInput
	}
	return repo, nil
}

func normalizeMarketTagRequest(req MarketTagRequest, actorUsername string) (MarketTag, error) {
	if req.Slug == "" {
		req.Slug = slugFromDisplayName(req.DisplayName)
	}
	if err := validateMarketTagRequest(req, true); err != nil {
		return MarketTag{}, err
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	return MarketTag{
		Slug:        NormalizeMarketTagSlug(req.Slug),
		DisplayName: strings.TrimSpace(req.DisplayName),
		Description: strings.TrimSpace(req.Description),
		ColorKey:    strings.TrimSpace(strings.ToLower(req.ColorKey)),
		SortOrder:   req.SortOrder,
		IsActive:    active,
		CreatedBy:   strings.TrimSpace(actorUsername),
	}, nil
}

func validateMarketTagRequest(req MarketTagRequest, requireDisplayName bool) error {
	slug := NormalizeMarketTagSlug(req.Slug)
	displayName := strings.TrimSpace(req.DisplayName)
	description := strings.TrimSpace(req.Description)
	colorKey := strings.TrimSpace(strings.ToLower(req.ColorKey))

	if slug != "" && (len(slug) > MaxMarketTagSlugLength || !marketTagSlugPattern.MatchString(slug)) {
		return ErrInvalidInput
	}
	if requireDisplayName && displayName == "" {
		return ErrInvalidInput
	}
	if displayName != "" && len(displayName) > MaxMarketTagDisplayNameLength {
		return ErrInvalidInput
	}
	if len(description) > MaxMarketTagDescriptionLength {
		return ErrInvalidInput
	}
	if len(colorKey) > MaxMarketTagColorKeyLength || !marketTagColorKeyPattern.MatchString(colorKey) {
		return ErrInvalidInput
	}
	return nil
}

// NormalizeMarketTagSlug canonicalizes a tag slug for storage and comparisons.
func NormalizeMarketTagSlug(slug string) string {
	return strings.Trim(strings.ToLower(strings.TrimSpace(slug)), "-")
}

// NormalizeMarketTagSlugs canonicalizes, validates, de-duplicates, and sorts tag slugs.
func NormalizeMarketTagSlugs(slugs []string) ([]string, error) {
	if len(slugs) == 0 {
		return []string{}, nil
	}
	seen := make(map[string]bool, len(slugs))
	normalized := make([]string, 0, len(slugs))
	for _, raw := range slugs {
		slug := NormalizeMarketTagSlug(raw)
		if slug == "" {
			continue
		}
		if len(slug) > MaxMarketTagSlugLength || !marketTagSlugPattern.MatchString(slug) {
			return nil, ErrInvalidInput
		}
		if !seen[slug] {
			seen[slug] = true
			normalized = append(normalized, slug)
		}
	}
	if len(normalized) > MaxMarketTagsPerMarket {
		return nil, ErrInvalidInput
	}
	sort.Strings(normalized)
	return normalized, nil
}

func slugFromDisplayName(displayName string) string {
	displayName = strings.ToLower(strings.TrimSpace(displayName))
	var builder strings.Builder
	lastHyphen := false
	for _, r := range displayName {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastHyphen = false
			continue
		}
		if !lastHyphen {
			builder.WriteRune('-')
			lastHyphen = true
		}
	}
	return NormalizeMarketTagSlug(builder.String())
}
