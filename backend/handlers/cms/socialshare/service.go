package socialshare

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"socialpredict/models"

	"gorm.io/gorm"
)

const (
	settingsSlug                      = "default"
	imageSlug                         = "default"
	DefaultSiteName                   = "SocialPredict"
	DefaultDescription                = "Prediction markets for the social web"
	DefaultImageURL                   = "/logo512.png"
	DefaultImageAlt                   = "SocialPredict share card"
	UploadedImageURL                  = "/v0/content/social-share/image"
	MaxSiteNameLength                 = 80
	MaxDefaultDescriptionLength       = 220
	MaxDefaultImageURLLength          = 500
	MaxImageAltLength                 = 160
	MaxUploadedImageBytes       int64 = 5 * 1024 * 1024
)

var allowedImageContentTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

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

type UploadImageInput struct {
	FileName  string
	Data      []byte
	ImageAlt  string
	UpdatedBy string
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

func (s *Service) GetImage() (*models.SocialShareImage, error) {
	return s.repo.GetImageBySlug(imageSlug)
}

func (s *Service) UpdateSettings(in UpdateInput) (*models.SocialShareSettings, error) {
	item, err := s.getSettingsForUpdate()
	if err != nil {
		return nil, err
	}
	if in.Version != 0 && item.ID != 0 && in.Version != item.Version {
		return nil, errors.New("version mismatch")
	}

	siteName, description, imageURL, imageAlt, err := validateSettingsInput(in.SiteName, in.DefaultDescription, in.DefaultImageURL, in.ImageAlt)
	if err != nil {
		return nil, err
	}

	item.SiteName = siteName
	item.DefaultDescription = description
	item.DefaultImageURL = imageURL
	item.ImageAlt = imageAlt
	item.UpdatedBy = strings.TrimSpace(in.UpdatedBy)
	bumpSettingsVersion(item)

	if err := s.repo.Save(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) UploadImage(in UploadImageInput) (*models.SocialShareSettings, error) {
	image, err := s.getImageForUpdate()
	if err != nil {
		return nil, err
	}
	contentType, err := validateUploadedImage(in.Data)
	if err != nil {
		return nil, err
	}

	image.Slug = imageSlug
	image.FileName = truncateFileName(in.FileName)
	image.ContentType = contentType
	image.SizeBytes = int64(len(in.Data))
	image.Data = append(image.Data[:0], in.Data...)
	image.UpdatedBy = strings.TrimSpace(in.UpdatedBy)
	if err := s.repo.SaveImage(image); err != nil {
		return nil, err
	}

	settings, err := s.getSettingsForUpdate()
	if err != nil {
		return nil, err
	}
	settings.DefaultImageURL = UploadedImageURL
	settings.UpdatedBy = strings.TrimSpace(in.UpdatedBy)
	if imageAlt, err := validateText("image alt", in.ImageAlt, MaxImageAltLength, false); err != nil {
		return nil, err
	} else if imageAlt != "" {
		settings.ImageAlt = imageAlt
	} else if strings.TrimSpace(settings.ImageAlt) == "" {
		settings.ImageAlt = DefaultImageAlt
	}
	bumpSettingsVersion(settings)
	if err := s.repo.Save(settings); err != nil {
		return nil, err
	}
	return settings, nil
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

func (s *Service) getSettingsForUpdate() (*models.SocialShareSettings, error) {
	item, err := s.repo.GetBySlug(settingsSlug)
	if err == nil {
		return item, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultSettings(), nil
	}
	return nil, err
}

func (s *Service) getImageForUpdate() (*models.SocialShareImage, error) {
	item, err := s.repo.GetImageBySlug(imageSlug)
	if err == nil {
		return item, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &models.SocialShareImage{Slug: imageSlug}, nil
	}
	return nil, err
}

func validateSettingsInput(siteNameValue string, descriptionValue string, imageURLValue string, imageAltValue string) (string, string, string, string, error) {
	siteName, err := validateText("site name", siteNameValue, MaxSiteNameLength, true)
	if err != nil {
		return "", "", "", "", err
	}
	description, err := validateText("default description", descriptionValue, MaxDefaultDescriptionLength, true)
	if err != nil {
		return "", "", "", "", err
	}
	imageURL, err := validateImageURL(imageURLValue)
	if err != nil {
		return "", "", "", "", err
	}
	imageAlt, err := validateText("image alt", imageAltValue, MaxImageAltLength, false)
	if err != nil {
		return "", "", "", "", err
	}
	if imageAlt == "" {
		imageAlt = DefaultImageAlt
	}
	return siteName, description, imageURL, imageAlt, nil
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

func validateUploadedImage(data []byte) (string, error) {
	if len(data) == 0 {
		return "", errors.New("image is required")
	}
	if int64(len(data)) > MaxUploadedImageBytes {
		return "", errors.New("image is too large")
	}
	contentType := http.DetectContentType(data)
	if _, ok := allowedImageContentTypes[contentType]; !ok {
		return "", errors.New("unsupported image content type")
	}
	return contentType, nil
}

func truncateFileName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "social-share-image"
	}
	runes := []rune(value)
	if len(runes) > 255 {
		return string(runes[:255])
	}
	return value
}

func bumpSettingsVersion(item *models.SocialShareSettings) {
	if item.ID == 0 || item.Version == 0 {
		item.Version = 1
		return
	}
	item.Version++
}
