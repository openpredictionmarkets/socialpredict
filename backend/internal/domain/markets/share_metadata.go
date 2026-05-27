package markets

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
)

const (
	defaultShareSiteName       = "SocialPredict"
	defaultShareDescription    = "Prediction markets for the social web"
	defaultShareImagePath      = "/logo512.png"
	maxShareTitleLength        = 120
	maxShareDescriptionLength  = 220
	shareMarketCanonicalPrefix = "/markets"
)

// ShareMetadataConfig carries runtime-owned values needed for public share cards.
type ShareMetadataConfig struct {
	PublicBaseURL   string
	DefaultImageURL string
	SiteName        string
}

// ShareMetadata is the public read model consumed by link-preview rendering.
type ShareMetadata struct {
	MarketID     int64
	Title        string
	Description  string
	CanonicalURL string
	ImageURL     string
	PublicStatus string
	SiteName     string
	Creator      string
	MarketTitle  string
	Shareable    bool
}

// GetShareMetadata returns the public metadata approved for market link previews.
func (s *Service) GetShareMetadata(ctx context.Context, marketID int64, config ShareMetadataConfig) (*ShareMetadata, error) {
	market, err := s.GetPublicMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if market == nil {
		return nil, ErrMarketNotFound
	}

	status := PublicStatusFromLifecycle(market.LifecycleStatus, market.IsResolved, market.ResolutionDateTime, s.clock.Now())
	if !shareablePublicStatus(status) {
		return nil, ErrMarketNotFound
	}

	siteName := strings.TrimSpace(config.SiteName)
	if siteName == "" {
		siteName = defaultShareSiteName
	}

	marketTitle := truncateForShare(strings.TrimSpace(market.QuestionTitle), maxShareTitleLength)
	if marketTitle == "" {
		marketTitle = "Prediction market"
	}

	description := truncateForShare(strings.TrimSpace(market.Description), maxShareDescriptionLength)
	if description == "" {
		description = defaultShareDescription
	}

	return &ShareMetadata{
		MarketID:     market.ID,
		Title:        marketTitle + " | " + siteName,
		Description:  description,
		CanonicalURL: shareURL(config.PublicBaseURL, shareMarketCanonicalPrefix, market.ID),
		ImageURL:     shareImageURL(config.PublicBaseURL, config.DefaultImageURL),
		PublicStatus: status,
		SiteName:     siteName,
		Creator:      market.CreatorUsername,
		MarketTitle:  marketTitle,
		Shareable:    true,
	}, nil
}

func shareablePublicStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case MarketStatusActive, MarketStatusClosed, MarketStatusResolved:
		return true
	default:
		return false
	}
}

func truncateForShare(value string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(strings.Join(strings.Fields(value), " "))
	if len(runes) <= maxRunes {
		return string(runes)
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}

func shareURL(base string, elems ...any) string {
	parsed, err := url.Parse(normalizePublicBaseURL(base))
	if err != nil {
		parsed, _ = url.Parse("http://localhost")
	}

	pathParts := []string{parsed.Path}
	for _, elem := range elems {
		pathParts = append(pathParts, strings.Trim(strings.TrimSpace(toSharePathPart(elem)), "/"))
	}
	parsed.Path = path.Join(pathParts...)
	return parsed.String()
}

func shareImageURL(base string, image string) string {
	image = strings.TrimSpace(image)
	if image == "" {
		return shareURL(base, defaultShareImagePath)
	}
	if parsed, err := url.Parse(image); err == nil && parsed.IsAbs() {
		return parsed.String()
	}
	return shareURL(base, image)
}

func normalizePublicBaseURL(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		return "http://localhost"
	}
	if strings.HasPrefix(base, "http://") || strings.HasPrefix(base, "https://") {
		return base
	}
	return "https://" + base
}

func toSharePathPart(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case int64:
		return strconv.FormatInt(typed, 10)
	default:
		return strings.TrimSpace(strings.ReplaceAll(fmt.Sprint(value), " ", ""))
	}
}
