package runtime

import "testing"

func TestLoadConfigService(t *testing.T) {
	service, err := LoadConfigService()
	if err != nil {
		t.Fatalf("LoadConfigService returned error: %v", err)
	}
	if service == nil {
		t.Fatalf("expected config service, got nil")
	}
	if service.Current() == nil {
		t.Fatalf("expected current config, got nil")
	}
}
