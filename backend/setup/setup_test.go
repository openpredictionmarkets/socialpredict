package setup

import "testing"

func TestLoadEconomicsConfigSingleton(t *testing.T) {
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

	if cfg1 != cfg2 {
		t.Fatalf("expected LoadEconomicsConfig to return singleton instance")
	}

	if cfg1.Economics.User.MaximumDebtAllowed <= 0 {
		t.Fatalf("expected maximum debt to be populated, got %d", cfg1.Economics.User.MaximumDebtAllowed)
	}

	if EconomicsConfig() != cfg1 {
		t.Fatalf("EconomicsConfig should return the singleton instance")
	}
}

func TestChartSigFigsClamping(t *testing.T) {
	cfg, err := LoadEconomicsConfig()
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	original := cfg.Frontend.Charts.SigFigs
	t.Cleanup(func() {
		cfg.Frontend.Charts.SigFigs = original
	})

	cfg.Frontend.Charts.SigFigs = 1
	if got := ChartSigFigs(); got != minChartSigFigs {
		t.Fatalf("expected minChartSigFigs (%d), got %d", minChartSigFigs, got)
	}

	cfg.Frontend.Charts.SigFigs = 99
	if got := ChartSigFigs(); got != maxChartSigFigs {
		t.Fatalf("expected maxChartSigFigs (%d), got %d", maxChartSigFigs, got)
	}

	cfg.Frontend.Charts.SigFigs = 6
	if got := ChartSigFigs(); got != 6 {
		t.Fatalf("expected unclamped value 6, got %d", got)
	}
}
