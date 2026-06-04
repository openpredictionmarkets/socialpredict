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
	MinRecommendation    = 1
	MaxRecommendation    = 50
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
	SectionsEnabled            bool
	DefaultRecommendationLimit int
	CuratedRecommendationLimit int
	IsPublished                bool
	UpdatedBy                  string
	Version                    uint
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
	page.SectionsEnabled = in.SectionsEnabled
	page.DefaultRecommendationLimit = defaultLimit
	page.CuratedRecommendationLimit = curatedLimit
	page.IsPublished = in.IsPublished
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
		SectionsEnabled:            false,
		DefaultRecommendationLimit: 20,
		CuratedRecommendationLimit: 5,
		IsPublished:                true,
		Version:                    1,
	}
	if slug == PageSlugTopicTemplate {
		page.Title = "Topic Markets"
		page.Description = "Browse and search markets in this topic."
		page.PageType = PageTypeSecondary
		page.SearchScope = SearchScopeTag
		page.FeaturedMarketsEnabled = true
		page.SectionsEnabled = true
	}
	return page
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
