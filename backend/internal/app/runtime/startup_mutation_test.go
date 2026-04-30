package runtime

import (
	"os"
	"testing"
)

func TestLoadStartupMutationModeDefaultsToNonWriter(t *testing.T) {
	unsetEnvForTest(t, StartupWriterEnv)
	unsetEnvForTest(t, legacyStartupWriterEnv)

	mode, err := LoadStartupMutationModeFromEnv()
	if err != nil {
		t.Fatalf("LoadStartupMutationModeFromEnv returned error: %v", err)
	}
	if mode.Writer {
		t.Fatalf("expected default startup mutation mode to be non-writer")
	}
	if mode.Source != "default" {
		t.Fatalf("expected default startup mutation source, got %q", mode.Source)
	}
}

func TestLoadStartupMutationModePrimaryFlagTakesPrecedence(t *testing.T) {
	t.Setenv(StartupWriterEnv, "0")
	t.Setenv(legacyStartupWriterEnv, "true")

	mode, err := LoadStartupMutationModeFromEnv()
	if err != nil {
		t.Fatalf("LoadStartupMutationModeFromEnv returned error: %v", err)
	}
	if mode.Writer {
		t.Fatalf("expected explicit false startup writer flag to disable writer mode")
	}
	if mode.Source != StartupWriterEnv {
		t.Fatalf("expected primary startup writer env source, got %q", mode.Source)
	}
}

func TestLoadStartupMutationModeUsesLegacyWhenPrimaryMissing(t *testing.T) {
	t.Setenv(legacyStartupWriterEnv, "true")

	mode, err := LoadStartupMutationModeFromEnv()
	if err != nil {
		t.Fatalf("LoadStartupMutationModeFromEnv returned error: %v", err)
	}
	if !mode.Writer {
		t.Fatalf("expected legacy startup writer flag to enable writer mode")
	}
	if mode.Source != legacyStartupWriterEnv {
		t.Fatalf("expected legacy startup writer env source, got %q", mode.Source)
	}
}

func TestLoadStartupMutationModeRejectsInvalidFlag(t *testing.T) {
	t.Setenv(StartupWriterEnv, "sometimes")

	if _, err := LoadStartupMutationModeFromEnv(); err == nil {
		t.Fatalf("expected invalid startup writer flag to fail closed")
	}
}

func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()

	original, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv(%s): %v", key, err)
	}
	t.Cleanup(func() {
		if existed {
			if err := os.Setenv(key, original); err != nil {
				t.Fatalf("Setenv(%s): %v", key, err)
			}
			return
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%s): %v", key, err)
		}
	})
}
