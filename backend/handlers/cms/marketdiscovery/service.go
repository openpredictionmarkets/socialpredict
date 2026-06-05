package marketdiscovery

import (
	"errors"
	"strings"

	"socialpredict/models"

	"gorm.io/gorm"
)

const (
	PageSlugMarkets       = "markets"
	PageSlugTopicTemplate = "topic-template"

	PageTypeTop       = "top"
	PageTypeSecondary = "secondary"

	SearchScopeAll = "all"
	SearchScopeTag = "tag"

	MaxTitleLength       = 160
	MaxDescriptionLength = 500
	MaxSlugLength        = 96
	MaxPinsPerPage       = 48
	MinRecommendation    = 1
	MaxRecommendation    = 50

	ScopeTypePage        = "page"
	PinTypeMarket        = "market"
	PinTypeDiscoveryPage = "discovery_page"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type UpdateInput struct {
	Title                      string
	Description                string
	PageType                   string
	PrimaryTagSlug             string
	SearchScope                string
	FeaturedTopicsEnabled      bool
	FeaturedMarketsEnabled     bool
	DefaultRecommendationLimit int
	CuratedRecommendationLimit int
	UpdatedBy                  string
	Version                    uint
}

type PinInput struct {
	PinType        string
	MarketID       int64
	TargetPageSlug string
	Label          string
	SortOrder      int
}

type PageComposition struct {
	Page *models.MarketDiscoveryPage
	Pins []models.MarketDiscoveryPin
}

func (s *Service) GetPage(slug string) (*models.MarketDiscoveryPage, error) {
	slug = normalizeSlug(slug)
	if slug == "" {
		return nil, errors.New("page slug is required")
	}
	page, err := s.repo.GetPageBySlug(slug)
	if err == nil {
		return page, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultPage(slug), nil
	}
	return nil, err
}

func (s *Service) GetComposition(slug string) (*PageComposition, error) {
	page, err := s.GetPage(slug)
	if err != nil {
		return nil, err
	}
	if page.ID == 0 {
		return &PageComposition{Page: page, Pins: []models.MarketDiscoveryPin{}}, nil
	}
	pins, err := s.repo.ListPins(page.ID)
	if err != nil {
		return nil, err
	}
	return &PageComposition{Page: page, Pins: pins}, nil
}

func (s *Service) UpdatePage(slug string, in UpdateInput) (*models.MarketDiscoveryPage, error) {
	slug = normalizeSlug(slug)
	if slug == "" {
		return nil, errors.New("page slug is required")
	}
	page, err := s.GetPage(slug)
	if err != nil {
		return nil, err
	}
	if in.Version != 0 && page.ID != 0 && in.Version != page.Version {
		return nil, errors.New("version mismatch")
	}

	title, description, pageType, primaryTagSlug, searchScope, defaultLimit, curatedLimit, err := validateInput(slug, in)
	if err != nil {
		return nil, err
	}

	page.Slug = slug
	page.Title = title
	page.Description = description
	page.PageType = pageType
	page.PrimaryTagSlug = primaryTagSlug
	page.SearchScope = searchScope
	page.FeaturedTopicsEnabled = in.FeaturedTopicsEnabled
	page.FeaturedMarketsEnabled = in.FeaturedMarketsEnabled
	page.DefaultRecommendationLimit = defaultLimit
	page.CuratedRecommendationLimit = curatedLimit
	page.UpdatedBy = strings.TrimSpace(in.UpdatedBy)
	if page.ID == 0 {
		page.Version = 1
	} else {
		page.Version = page.Version + 1
	}

	if err := s.repo.SavePage(page); err != nil {
		return nil, err
	}
	return page, nil
}

func (s *Service) ReplacePins(slug string, inputs []PinInput, actorUsername string) (*PageComposition, error) {
	if len(inputs) > MaxPinsPerPage {
		return nil, errors.New("too many pins")
	}
	page, err := s.ensurePersistedPage(slug)
	if err != nil {
		return nil, err
	}
	pins := make([]models.MarketDiscoveryPin, 0, len(inputs))
	for index, input := range inputs {
		pin, err := pinFromInput(input, index, actorUsername)
		if err != nil {
			return nil, err
		}
		pins = append(pins, pin)
	}
	if err := s.repo.ReplacePins(page.ID, pins); err != nil {
		return nil, err
	}
	return s.GetComposition(page.Slug)
}

func (s *Service) ensurePersistedPage(slug string) (*models.MarketDiscoveryPage, error) {
	page, err := s.GetPage(slug)
	if err != nil {
		return nil, err
	}
	if page.ID != 0 {
		return page, nil
	}
	if err := s.repo.SavePage(page); err != nil {
		return nil, err
	}
	return page, nil
}

func DefaultPage(slug string) *models.MarketDiscoveryPage {
	slug = normalizeSlug(slug)
	page := &models.MarketDiscoveryPage{
		Slug:                       slug,
		Title:                      "Markets",
		Description:                "Browse and search prediction markets.",
		PageType:                   PageTypeTop,
		SearchScope:                SearchScopeAll,
		FeaturedTopicsEnabled:      false,
		FeaturedMarketsEnabled:     false,
		DefaultRecommendationLimit: 20,
		CuratedRecommendationLimit: 5,
		Version:                    1,
	}
	if slug == PageSlugTopicTemplate {
		page.Title = "Topic Markets"
		page.Description = "Browse and search markets in this topic."
		page.PageType = PageTypeSecondary
		page.SearchScope = SearchScopeTag
		page.FeaturedMarketsEnabled = true
	} else if slug != PageSlugMarkets {
		page.Title = titleFromSlug(slug)
		page.Description = "Browse and search markets in this topic."
		page.PageType = PageTypeSecondary
		page.PrimaryTagSlug = slug
		page.SearchScope = SearchScopeTag
		page.FeaturedMarketsEnabled = true
	}
	return page
}

func pinFromInput(input PinInput, index int, actorUsername string) (models.MarketDiscoveryPin, error) {
	pinType := strings.ToLower(strings.TrimSpace(input.PinType))
	if pinType == "" {
		pinType = PinTypeMarket
	}
	if pinType != PinTypeMarket && pinType != PinTypeDiscoveryPage {
		return models.MarketDiscoveryPin{}, errors.New("invalid pin type")
	}
	label := strings.Join(strings.Fields(strings.TrimSpace(input.Label)), " ")
	if len([]rune(label)) > MaxTitleLength {
		return models.MarketDiscoveryPin{}, errors.New("pin label is too long")
	}
	sortOrder := input.SortOrder
	if sortOrder == 0 {
		sortOrder = index + 1
	}
	pin := models.MarketDiscoveryPin{
		ScopeType: ScopeTypePage,
		PinType:   pinType,
		Label:     label,
		SortOrder: sortOrder,
		CreatedBy: strings.TrimSpace(actorUsername),
	}
	switch pinType {
	case PinTypeMarket:
		if input.MarketID <= 0 {
			return models.MarketDiscoveryPin{}, errors.New("market pin requires market id")
		}
		pin.MarketID = &input.MarketID
	case PinTypeDiscoveryPage:
		targetPageSlug := normalizeSlug(input.TargetPageSlug)
		if targetPageSlug == "" {
			return models.MarketDiscoveryPin{}, errors.New("discovery page pin requires target page slug")
		}
		pin.TargetPageSlug = targetPageSlug
	}
	return pin, nil
}

func validateInput(slug string, in UpdateInput) (string, string, string, string, string, int, int, error) {
	title := strings.Join(strings.Fields(strings.TrimSpace(in.Title)), " ")
	if title == "" {
		title = DefaultPage(slug).Title
	}
	if len([]rune(title)) > MaxTitleLength {
		return "", "", "", "", "", 0, 0, errors.New("title is too long")
	}
	description := strings.Join(strings.Fields(strings.TrimSpace(in.Description)), " ")
	if len([]rune(description)) > MaxDescriptionLength {
		return "", "", "", "", "", 0, 0, errors.New("description is too long")
	}
	pageType := strings.ToLower(strings.TrimSpace(in.PageType))
	if pageType == "" {
		pageType = DefaultPage(slug).PageType
	}
	if pageType != PageTypeTop && pageType != PageTypeSecondary {
		return "", "", "", "", "", 0, 0, errors.New("invalid page type")
	}
	searchScope := strings.ToLower(strings.TrimSpace(in.SearchScope))
	if searchScope == "" {
		searchScope = DefaultPage(slug).SearchScope
	}
	if searchScope != SearchScopeAll && searchScope != SearchScopeTag {
		return "", "", "", "", "", 0, 0, errors.New("invalid search scope")
	}
	defaultLimit := clampRecommendationLimit(in.DefaultRecommendationLimit, DefaultPage(slug).DefaultRecommendationLimit)
	curatedLimit := clampRecommendationLimit(in.CuratedRecommendationLimit, DefaultPage(slug).CuratedRecommendationLimit)
	return title, description, pageType, normalizeSlug(in.PrimaryTagSlug), searchScope, defaultLimit, curatedLimit, nil
}

func clampRecommendationLimit(value int, fallback int) int {
	if value == 0 {
		value = fallback
	}
	if value < MinRecommendation {
		return MinRecommendation
	}
	if value > MaxRecommendation {
		return MaxRecommendation
	}
	return value
}

func normalizeSlug(value string) string {
	return strings.Trim(strings.ToLower(strings.TrimSpace(value)), "-")
}

func slugFromTitle(title string) string {
	title = strings.ToLower(strings.TrimSpace(title))
	var builder strings.Builder
	lastHyphen := false
	for _, r := range title {
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
	return normalizeSlug(builder.String())
}

func titleFromSlug(slug string) string {
	parts := strings.Fields(strings.ReplaceAll(normalizeSlug(slug), "-", " "))
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	if len(parts) == 0 {
		return "Topic Markets"
	}
	return strings.Join(parts, " ")
}
