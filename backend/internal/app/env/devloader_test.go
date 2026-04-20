package env

import (
	"os"
	"testing"
)

func TestLoadDevFileLoadsFile(t *testing.T) {
	tempDir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(origWD)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := os.WriteFile(".env.dev", []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	origVal, had := os.LookupEnv("FOO")
	if had {
		t.Cleanup(func() {
			os.Setenv("FOO", origVal)
		})
	} else {
		t.Cleanup(func() {
			os.Unsetenv("FOO")
		})
	}

	if err := os.Unsetenv("FOO"); err != nil {
		t.Fatalf("unsetenv: %v", err)
	}

	if err := LoadDevFile(); err != nil {
		t.Fatalf("LoadDevFile returned error: %v", err)
	}

	if got := os.Getenv("FOO"); got != "bar" {
		t.Fatalf("expected env var to be loaded, got %q", got)
	}
}

func TestLoadDevFileMissingFile(t *testing.T) {
	tempDir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(origWD)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origVal, had := os.LookupEnv("FOO")
	if had {
		t.Cleanup(func() {
			os.Setenv("FOO", origVal)
		})
	} else {
		t.Cleanup(func() {
			os.Unsetenv("FOO")
		})
	}

	if err := os.Unsetenv("FOO"); err != nil {
		t.Fatalf("unsetenv: %v", err)
	}

	if err := LoadDevFile(); err != nil {
		t.Fatalf("LoadDevFile should not fail when file missing: %v", err)
	}

	if got := os.Getenv("FOO"); got != "" {
		t.Fatalf("expected env var to remain empty, got %q", got)
	}
}
