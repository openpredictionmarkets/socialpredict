package logger

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

// captureStdout redirects os.Stdout for the duration of fn and returns what was written.
func captureStdout(fn func()) (string, error) {
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	defer func() {
		_ = w.Close()
		os.Stdout = orig
	}()

	fn()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// helper wrappers to simulate the expected call depth for runtime.Caller(3)
func userLike(l *CustomLogger) {
	callInfo(l)
}
func callInfo(l *CustomLogger) {
	l.Info("Test - Caller: message")
}

func TestLogInfo_Convenience_WritesExpectedParts(t *testing.T) {
	out, err := captureStdout(func() {
		LogInfo("ChangePassword", "ChangePassword", "ChangePassword handler called")
	})
	if err != nil {
		t.Fatalf("failed capturing stdout: %v", err)
	}

	if !strings.Contains(out, "INFO") {
		t.Fatalf("expected INFO level in output, got: %q", out)
	}
	if !strings.Contains(out, "ChangePassword - ChangePassword: ChangePassword handler called") {
		t.Fatalf("expected formatted message in output, got: %q", out)
	}
	// Should include the logger method name per implementation
	if !strings.Contains(out, "logger.(*CustomLogger).Info()") && !strings.Contains(out, "(*CustomLogger).Info()") {
		t.Fatalf("expected function name to include Info(), got: %q", out)
	}
	// Should include a file reference
	if !strings.Contains(out, ".go:") {
		t.Fatalf("expected file:line in output, got: %q", out)
	}
}

func TestLogError_Convenience_IncludesError(t *testing.T) {
	errExample := errors.New("password must be between 8 and 128 characters long")
	out, err := captureStdout(func() {
		LogError("ChangePassword", "ValidateNewPasswordStrength", errExample)
	})
	if err != nil {
		t.Fatalf("failed capturing stdout: %v", err)
	}

	if !strings.Contains(out, "ERROR") {
		t.Fatalf("expected ERROR level in output, got: %q", out)
	}
	if !strings.Contains(out, "ValidateNewPasswordStrength") {
		t.Fatalf("expected function context in output, got: %q", out)
	}
	if !strings.Contains(out, errExample.Error()) {
		t.Fatalf("expected error message in output, got: %q", out)
	}
	// Should include a file reference
	if !strings.Contains(out, ".go:") {
		t.Fatalf("expected file:line in output, got: %q", out)
	}
}

func TestCustomLogger_Info_CallerDepth_FileReported(t *testing.T) {
	// Create a temp file to write logs into (since NewCustomLogger expects *os.File)
	tmp, err := os.CreateTemp("", "simplelogger-test-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	l := NewCustomLogger(tmp, "", log.LstdFlags)

	// Call through two wrappers so that runtime.Caller(3) resolves to userLike (outside logger)
	userLike(l)

	// Ensure content is written
	_ = tmp.Sync()
	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatalf("failed to read temp log file: %v", err)
	}
	out := string(data)

	if !strings.Contains(out, "INFO") {
		t.Fatalf("expected INFO level in output, got: %q", out)
	}
	// Caller should reference this test file with a :line suffix
	if !strings.Contains(out, "simplelogging_test.go:") && !strings.Contains(out, "_test.go:") {
		t.Fatalf("expected file:line pointing to test file, got: %q", out)
	}
	if !strings.Contains(out, "Test - Caller: message") {
		t.Fatalf("expected emitted message, got: %q", out)
	}
}
