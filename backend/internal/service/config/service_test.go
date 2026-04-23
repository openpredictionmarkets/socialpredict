package config

import (
	"errors"
	"testing"

	"socialpredict/setup"
)

func TestNewServiceLoadsCurrentConfig(t *testing.T) {
	cfg := &setup.EconomicConfig{
		Economics: setup.Economics{
			User: setup.User{MaximumDebtAllowed: 500},
		},
		Frontend: setup.Frontend{
			Charts: setup.FrontendCharts{SigFigs: 7},
		},
	}

	svc, err := NewService(LoaderFunc(func() (*AppConfig, error) {
		return cfg, nil
	}))
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	if got := svc.Current(); got != cfg {
		t.Fatalf("Current returned unexpected config pointer")
	}
	if got := svc.Economics().User.MaximumDebtAllowed; got != 500 {
		t.Fatalf("Economics returned wrong maximum debt: got %d", got)
	}
	if got := svc.ChartSigFigs(); got != 7 {
		t.Fatalf("ChartSigFigs returned %d, want 7", got)
	}
}

func TestNewServicePropagatesLoaderError(t *testing.T) {
	wantErr := errors.New("boom")

	svc, err := NewService(LoaderFunc(func() (*AppConfig, error) {
		return nil, wantErr
	}))
	if !errors.Is(err, wantErr) {
		t.Fatalf("NewService error = %v, want %v", err, wantErr)
	}
	if svc != nil {
		t.Fatalf("expected nil service on loader error")
	}
}

func TestClampChartSigFigs(t *testing.T) {
	tests := []struct {
		name     string
		frontend Frontend
		want     int
	}{
		{name: "defaults zero", frontend: Frontend{}, want: 4},
		{
			name: "clamps low",
			frontend: Frontend{
				Charts: setup.FrontendCharts{SigFigs: 1},
			},
			want: 2,
		},
		{
			name: "clamps high",
			frontend: Frontend{
				Charts: setup.FrontendCharts{SigFigs: 99},
			},
			want: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClampChartSigFigs(tt.frontend); got != tt.want {
				t.Fatalf("ClampChartSigFigs() = %d, want %d", got, tt.want)
			}
		})
	}
}
