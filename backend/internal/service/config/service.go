package config

import "socialpredict/setup"

// Service exposes typed access to application configuration slices.
type Service interface {
	Current() *AppConfig
	Economics() Economics
	Frontend() Frontend
	ChartSigFigs() int
}

// Loader loads configuration from its source of truth.
type Loader interface {
	Load() (*AppConfig, error)
}

type LoaderFunc func() (*AppConfig, error)

func (f LoaderFunc) Load() (*AppConfig, error) {
	return f()
}

type RuntimeService struct {
	current *AppConfig
}

func NewService(loader Loader) (*RuntimeService, error) {
	if loader == nil {
		return NewStaticService((*AppConfig)(nil)), nil
	}

	cfg, err := loader.Load()
	if err != nil {
		return nil, err
	}

	return NewStaticService(cfg), nil
}

// NewStaticService preserves a static config snapshot, including legacy setup snapshots.
func NewStaticService(cfg any) *RuntimeService {
	return &RuntimeService{current: normalizeConfig(cfg)}
}

func (s *RuntimeService) Current() *AppConfig {
	if s == nil || s.current == nil {
		return &AppConfig{}
	}
	return s.current.Clone()
}

func (s *RuntimeService) Economics() Economics {
	return s.Current().Economics
}

func (s *RuntimeService) Frontend() Frontend {
	return s.Current().Frontend
}

func (s *RuntimeService) ChartSigFigs() int {
	return ClampChartSigFigs(s.Frontend())
}

func normalizeConfig(cfg any) *AppConfig {
	switch typed := cfg.(type) {
	case nil:
		return &AppConfig{}
	case *AppConfig:
		return typed.Clone()
	case AppConfig:
		return typed.Clone()
	case *setup.EconomicConfig:
		return FromSetup(typed)
	case setup.EconomicConfig:
		return FromSetup(&typed)
	default:
		return &AppConfig{}
	}
}
