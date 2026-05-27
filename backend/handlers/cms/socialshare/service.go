package socialshare

import (
	"errors"
	"net/url"
	"strings"

	"socialpredict/models"

	"gorm.io/gorm"
)

const (
	settingsSlug                = "default"
	DefaultSiteName             = "SocialPredict"
	DefaultDescription          = "Prediction markets for the social web"
	DefaultImageURL             = "/logo512.png"
	DefaultImageAlt             = "SocialPredict share card"
	MaxSiteNameLength           = 80
	MaxDefaultDescriptionLength = 220
	MaxDefaultImageURLLength    = 500
	MaxImageAltLength           = 160
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type UpdateInput struct {
	SiteName           string
	DefaultDescription string
	DefaultImageURL    string
	ImageAlt           string
	UpdatedBy          string
	Version            uint
}

func (s *Service) GetSettings() (*models.SocialShareSettings, error) {
	item, err := s.repo.GetBySlug(settingsSlug)
	if err == nil {
		return item, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultSettings(), nil
	}
	return nil, err
}

func (s *Service) UpdateSettings(in UpdateInput) (*models.SocialShareSettings, error) {
	item, err := s.repo.GetBySlug(settingsSlug)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		item = DefaultSettings()
	}
	if in.Version != 0 && item.ID != 0 && in.Version != item.Version {
		return nil, errors.New("version mismatch")
	}

	siteName, err := validateText("site name", in.SiteName, MaxSiteNameLength, true)
	if err != nil {
		return nil, err
	}
	description, err := validateText("default description", in.DefaultDescription, MaxDefaultDescriptionLength, true)
	if err != nil {
		return nil, err
	}
	imageURL, err := validateImageURL(in.DefaultImageURL)
	if err != nil {
		return nil, err
	}
	imageAlt, err := validateText("image alt", in.ImageAlt, MaxImageAltLength, false)
	if err != nil {
		return nil, err
	}
	if imageAlt == "" {
		imageAlt = DefaultImageAlt
	}

	item.SiteName = siteName
	item.DefaultDescription = description
	item.DefaultImageURL = imageURL
	item.ImageAlt = imageAlt
	item.UpdatedBy = strings.TrimSpace(in.UpdatedBy)
	if item.ID == 0 || item.Version == 0 {
		item.Version = 1
	} else {
		item.Version++
	}

	if err := s.repo.Save(item); err != nil {
		return nil, err
	}
	return item, nil
}

func DefaultSettings() *models.SocialShareSettings {
	return &models.SocialShareSettings{
		Slug:               settingsSlug,
		SiteName:           DefaultSiteName,
		DefaultDescription: DefaultDescription,
		DefaultImageURL:    DefaultImageURL,
		ImageAlt:           DefaultImageAlt,
		Version:            1,
	}
}

func validateText(label string, value string, maxRunes int, required bool) (string, error) {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if required && value == "" {
		return "", errors.New(label + " is required")
	}
	if len([]rune(value)) > maxRunes {
		return "", errors.New(label + " is too long")
	}
	return value, nil
}

func validateImageURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("default image URL is required")
	}
	if len([]rune(value)) > MaxDefaultImageURLLength {
		return "", errors.New("default image URL is too long")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", errors.New("default image URL is invalid")
	}
	if parsed.IsAbs() {
		scheme := strings.ToLower(parsed.Scheme)
		if scheme != "http" && scheme != "https" {
			return "", errors.New("default image URL must use http or https")
		}
		return parsed.String(), nil
	}
	if !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return "", errors.New("default image URL must be absolute or root-relative")
	}
	return value, nil
}
