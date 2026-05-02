package runtime

import (
	"testing"
	"time"
)

func TestLoadShutdownConfigDefaultsDrainWindow(t *testing.T) {
	t.Setenv(ReadinessDrainWindowEnv, "")
	t.Setenv(ShutdownTimeoutEnv, "")

	config, err := LoadShutdownConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadShutdownConfigFromEnv: %v", err)
	}
	if config.ReadinessDrainWindow != DefaultReadinessDrainWindow {
		t.Fatalf("ReadinessDrainWindow = %v, want %v", config.ReadinessDrainWindow, DefaultReadinessDrainWindow)
	}
	if config.ShutdownTimeout != DefaultShutdownTimeout {
		t.Fatalf("ShutdownTimeout = %v, want %v", config.ShutdownTimeout, DefaultShutdownTimeout)
	}
}

func TestLoadShutdownConfigUsesExplicitDrainWindow(t *testing.T) {
	t.Setenv(ReadinessDrainWindowEnv, "7")
	t.Setenv(ShutdownTimeoutEnv, "25")

	config, err := LoadShutdownConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadShutdownConfigFromEnv: %v", err)
	}
	if config.ReadinessDrainWindow != 7*time.Second {
		t.Fatalf("ReadinessDrainWindow = %v, want 7s", config.ReadinessDrainWindow)
	}
	if config.ShutdownTimeout != 25*time.Second {
		t.Fatalf("ShutdownTimeout = %v, want 25s", config.ShutdownTimeout)
	}
}

func TestLoadShutdownConfigRejectsInvalidReadinessDrainWindow(t *testing.T) {
	tests := []string{"0", "-1", "soon"}

	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			t.Setenv(ReadinessDrainWindowEnv, value)

			if _, err := LoadShutdownConfigFromEnv(); err == nil {
				t.Fatalf("expected invalid drain window %q to fail", value)
			}
		})
	}
}

func TestLoadShutdownConfigRejectsInvalidShutdownTimeout(t *testing.T) {
	tests := []string{"0", "-1", "eventually"}

	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			t.Setenv(ShutdownTimeoutEnv, value)

			if _, err := LoadShutdownConfigFromEnv(); err == nil {
				t.Fatalf("expected invalid shutdown timeout %q to fail", value)
			}
		})
	}
}

func TestNormalizeShutdownConfigDefaultsInvalidDrainWindow(t *testing.T) {
	config := NormalizeShutdownConfig(ShutdownConfig{})

	if config.ReadinessDrainWindow != DefaultReadinessDrainWindow {
		t.Fatalf("ReadinessDrainWindow = %v, want %v", config.ReadinessDrainWindow, DefaultReadinessDrainWindow)
	}
	if config.ShutdownTimeout != DefaultShutdownTimeout {
		t.Fatalf("ShutdownTimeout = %v, want %v", config.ShutdownTimeout, DefaultShutdownTimeout)
	}
}
