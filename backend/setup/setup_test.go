package setup

import (
	"sync"
	"testing"
)

type staticSource []byte

func (s staticSource) Bytes() ([]byte, error) {
	return append([]byte(nil), s...), nil
}

func resetLegacyLoadState() {
	economicConfig = nil
	legacyLoadState = struct {
		once sync.Once
		err  error
	}{}
}

func TestEmbeddedSourceDoesNotInitializeLegacySingleton(t *testing.T) {
	resetLegacyLoadState()

	data, err := (EmbeddedSource{}).Bytes()
	if err != nil {
		t.Fatalf("unexpected embedded source error: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("expected embedded setup source bytes")
	}
	if economicConfig != nil {
		t.Fatalf("expected embedded source access to avoid legacy singleton initialization")
	}
}

func TestLoadEconomicConfigFromSource(t *testing.T) {
	cfg, err := LoadEconomicConfigFromSource(staticSource(`
economics:
  user:
    maximumDebtAllowed: 123
frontend:
  charts:
    sigFigs: 5
`))
	if err != nil {
		t.Fatalf("unexpected error loading explicit source: %v", err)
	}
	if cfg.Economics.User.MaximumDebtAllowed != 123 {
		t.Fatalf("maximum debt = %d, want 123", cfg.Economics.User.MaximumDebtAllowed)
	}
	if cfg.Frontend.Charts.SigFigs != 5 {
		t.Fatalf("sig figs = %d, want 5", cfg.Frontend.Charts.SigFigs)
	}
}

func TestLoadEconomicsConfigSingleton(t *testing.T) {
	resetLegacyLoadState()

	cfg1, err := LoadEconomicsConfig()
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	if cfg1 == nil {
		t.Fatalf("expected non-nil config")
	}

	cfg2, err := LoadEconomicsConfig()
	if err != nil {
		t.Fatalf("unexpected error on second load: %v", err)
	}

	if economicConfig == nil {
		t.Fatalf("expected cached startup config to be initialized")
	}
	if cfg1 == cfg2 {
		t.Fatalf("expected LoadEconomicsConfig to return defensive copies")
	}
	if cfg1 == economicConfig || cfg2 == economicConfig {
		t.Fatalf("expected callers to receive detached copies of the cached startup snapshot")
	}
	if cfg1.Economics.User.MaximumDebtAllowed <= 0 {
		t.Fatalf("expected maximum debt to be populated, got %d", cfg1.Economics.User.MaximumDebtAllowed)
	}
	cfg1.Economics.User.MaximumDebtAllowed = 999

	cfg3, err := LoadEconomicsConfig()
	if err != nil {
		t.Fatalf("unexpected error on third load: %v", err)
	}
	if cfg3.Economics.User.MaximumDebtAllowed == 999 {
		t.Fatalf("mutating a returned config copy must not alter the cached startup snapshot")
	}

	if EconomicsConfig() == economicConfig {
		t.Fatalf("EconomicsConfig should return a defensive copy")
	}
}

func TestChartSigFigsUsesStartupSnapshot(t *testing.T) {
	resetLegacyLoadState()

	cfg, err := LoadEconomicsConfig()
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	cfg.Frontend.Charts.SigFigs = 99
	if got := ChartSigFigs(); got != 2 {
		t.Fatalf("expected embedded startup sig figs 2, got %d", got)
	}
	reloaded, err := LoadEconomicsConfig()
	if err != nil {
		t.Fatalf("unexpected error reloading config: %v", err)
	}
	if reloaded.Frontend.Charts.SigFigs == 99 {
		t.Fatalf("mutating a returned config copy must not affect subsequent reads")
	}
}
