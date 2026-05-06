package runtime

import (
	"errors"
	"testing"

	configsvc "socialpredict/internal/service/config"
	"socialpredict/setup"
)

func TestLoadConfigServiceUsesEmbeddedSource(t *testing.T) {
	service, err := LoadConfigService(setup.EmbeddedSource{})
	if err != nil {
		t.Fatalf("LoadConfigService returned error: %v", err)
	}
	if service == nil {
		t.Fatalf("expected config service, got nil")
	}
	if got := service.Current().Economics.User.MaximumDebtAllowed; got != 500 {
		t.Fatalf("maximum debt allowed = %d, want 500", got)
	}
	if got := service.ChartSigFigs(); got != 2 {
		t.Fatalf("chart sig figs = %d, want 2", got)
	}
}

func TestLoadConfigServiceUsesExplicitSource(t *testing.T) {
	service, err := LoadConfigService(configsvc.SourceFunc(func() ([]byte, error) {
		return []byte(`
economics:
  user:
    initialAccountBalance: 400
    maximumDebtAllowed: 120
frontend:
  charts:
    sigFigs: 8
`), nil
	}))
	if err != nil {
		t.Fatalf("LoadConfigService returned error: %v", err)
	}
	if service == nil {
		t.Fatalf("expected config service, got nil")
	}
	current := service.Current()
	if current == nil {
		t.Fatalf("expected current config, got nil")
	}
	if current.Economics.User.InitialAccountBalance != 400 {
		t.Fatalf("initial account balance = %d, want 400", current.Economics.User.InitialAccountBalance)
	}
	if current.Frontend.Charts.SigFigs != 8 {
		t.Fatalf("sig figs = %d, want 8", current.Frontend.Charts.SigFigs)
	}
}

func TestLoadConfigServiceFreezesRuntimeSnapshot(t *testing.T) {
	service, err := LoadConfigService(configsvc.SourceFunc(func() ([]byte, error) {
		return []byte(`
economics:
  user:
    maximumDebtAllowed: 120
`), nil
	}))
	if err != nil {
		t.Fatalf("LoadConfigService returned error: %v", err)
	}

	current := service.Current()
	current.Economics.User.MaximumDebtAllowed = 999

	if got := service.Current().Economics.User.MaximumDebtAllowed; got != 120 {
		t.Fatalf("runtime snapshot mutated to %d, want frozen 120", got)
	}
}

func TestLoadConfigServicePropagatesSourceError(t *testing.T) {
	wantErr := errors.New("load failed")

	service, err := LoadConfigService(configsvc.SourceFunc(func() ([]byte, error) {
		return nil, wantErr
	}))
	if !errors.Is(err, wantErr) {
		t.Fatalf("LoadConfigService error = %v, want %v", err, wantErr)
	}
	if service != nil {
		t.Fatalf("expected nil service on source error")
	}
}
