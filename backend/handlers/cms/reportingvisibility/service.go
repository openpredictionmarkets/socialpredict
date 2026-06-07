package reportingvisibility

import (
	"errors"
	"strings"

	"socialpredict/models"

	"gorm.io/gorm"
)

const SettingsSlugDefault = "default"

type Repository interface {
	GetSettings() (*models.ReportingVisibilitySettings, error)
	SaveSettings(*models.ReportingVisibilitySettings) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type UpdateInput struct {
	SystemMetricsPublic     *bool
	GlobalLeaderboardPublic *bool
	Version                 uint
	UpdatedBy               string
}

func (s *Service) GetSettings() (*models.ReportingVisibilitySettings, error) {
	item, err := s.repo.GetSettings()
	if err == nil {
		return item, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultSettings(), nil
	}
	return nil, err
}

func (s *Service) UpdateSettings(in UpdateInput) (*models.ReportingVisibilitySettings, error) {
	item, err := s.GetSettings()
	if err != nil {
		return nil, err
	}
	if in.Version != 0 && item.ID != 0 && in.Version != item.Version {
		return nil, errors.New("version mismatch")
	}

	item.Slug = SettingsSlugDefault
	if in.SystemMetricsPublic != nil {
		item.SystemMetricsPublic = *in.SystemMetricsPublic
	}
	if in.GlobalLeaderboardPublic != nil {
		item.GlobalLeaderboardPublic = *in.GlobalLeaderboardPublic
	}
	item.UpdatedBy = strings.TrimSpace(in.UpdatedBy)
	if item.ID == 0 {
		item.Version = 1
	} else {
		item.Version++
	}

	if err := s.repo.SaveSettings(item); err != nil {
		return nil, err
	}
	return item, nil
}

func DefaultSettings() *models.ReportingVisibilitySettings {
	return &models.ReportingVisibilitySettings{
		Slug:                    SettingsSlugDefault,
		SystemMetricsPublic:     true,
		GlobalLeaderboardPublic: true,
		Version:                 1,
	}
}
