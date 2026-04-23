package config

import "socialpredict/setup"

const (
	minChartSigFigs     = 2
	maxChartSigFigs     = 9
	defaultChartSigFigs = 4
)

type AppConfig = setup.EconomicConfig
type Economics = setup.Economics
type Frontend = setup.Frontend

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
		return NewStaticService(nil), nil
	}

	cfg, err := loader.Load()
	if err != nil {
		return nil, err
	}

	return NewStaticService(cfg), nil
}

func NewStaticService(cfg *AppConfig) *RuntimeService {
	if cfg == nil {
		cfg = &AppConfig{}
	}

	return &RuntimeService{current: cfg}
}

func (s *RuntimeService) Current() *AppConfig {
	if s == nil || s.current == nil {
		return &AppConfig{}
	}
	return s.current
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

func ClampChartSigFigs(frontend Frontend) int {
	sigFigs := frontend.Charts.SigFigs
	if sigFigs == 0 {
		sigFigs = defaultChartSigFigs
	}

	if sigFigs < minChartSigFigs {
		return minChartSigFigs
	}
	if sigFigs > maxChartSigFigs {
		return maxChartSigFigs
	}
	return sigFigs
}
