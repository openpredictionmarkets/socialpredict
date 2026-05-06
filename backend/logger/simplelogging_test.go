package logger

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func useStandardForTest(t *testing.T) *bytes.Buffer {
	t.Helper()

	original := standard
	buffer := &bytes.Buffer{}
	standard = newRuntimeLogger(buffer, func(int) {})

	t.Cleanup(func() {
		standard = original
	})

	return buffer
}

func TestLogInfoCompatibilityWrapperUsesStableRuntimeFields(t *testing.T) {
	buffer := useStandardForTest(t)

	LogInfo("startup", "LoadConfigService", "configuration loaded")

	output := buffer.String()
	if !strings.Contains(output, "level=INFO") {
		t.Fatalf("expected INFO level in output, got %q", output)
	}
	if !strings.Contains(output, `msg="configuration loaded"`) {
		t.Fatalf("expected message field in output, got %q", output)
	}
	if !strings.Contains(output, `component="startup"`) {
		t.Fatalf("expected component field in output, got %q", output)
	}
	if !strings.Contains(output, `operation="LoadConfigService"`) {
		t.Fatalf("expected operation field in output, got %q", output)
	}
	if !strings.Contains(output, `source="simplelogging_test.go:`) {
		t.Fatalf("expected caller source in output, got %q", output)
	}
}

func TestErrorRuntimeLoggingRedactsSensitiveMessageAndFields(t *testing.T) {
	var buffer bytes.Buffer
	logger := newRuntimeLogger(&buffer, func(int) {})

	logger.Error(
		"auth",
		"password=swordfish token=abc123",
		errors.New("token=abc123"),
		String("api_key", "key-123"),
		String("username", "alice"),
	)

	output := buffer.String()
	if !strings.Contains(output, `component="auth"`) {
		t.Fatalf("expected component field in output, got %q", output)
	}
	if !strings.Contains(output, `msg="password=[REDACTED] token=[REDACTED]"`) {
		t.Fatalf("expected redacted message in output, got %q", output)
	}
	if !strings.Contains(output, `error="token=[REDACTED]"`) {
		t.Fatalf("expected redacted error field in output, got %q", output)
	}
	if !strings.Contains(output, `api_key="[REDACTED]"`) {
		t.Fatalf("expected redacted api_key field in output, got %q", output)
	}
	if !strings.Contains(output, `username="alice"`) {
		t.Fatalf("expected non-sensitive field in output, got %q", output)
	}
}

func TestTraceContextFromTraceparentUsesOTelFieldVocabulary(t *testing.T) {
	var buffer bytes.Buffer
	logger := newRuntimeLogger(&buffer, func(int) {})

	traceFields := TraceContextFromTraceparent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	logger.Info("middleware", "request received", append(traceFields, Operation("AuthMiddleware"))...)

	output := buffer.String()
	if !strings.Contains(output, `trace_id="4bf92f3577b34da6a3ce929d0e0e4736"`) {
		t.Fatalf("expected trace_id in output, got %q", output)
	}
	if !strings.Contains(output, `span_id="00f067aa0ba902b7"`) {
		t.Fatalf("expected span_id in output, got %q", output)
	}
	if !strings.Contains(output, `trace_flags="01"`) {
		t.Fatalf("expected trace_flags in output, got %q", output)
	}
	if fields := TraceContextFromTraceparent("malformed"); len(fields) != 0 {
		t.Fatalf("expected malformed traceparent to produce no fields, got %+v", fields)
	}
}

func TestEventAndStateFieldsUseStableRuntimeVocabulary(t *testing.T) {
	var buffer bytes.Buffer
	logger := newRuntimeLogger(&buffer, func(int) {})

	logger.Info("runtime", "readiness opened", Event(EventReadinessOpen), Operation("MarkReady"), State("open"))

	output := buffer.String()
	for _, want := range []string{
		`component="runtime"`,
		`event="readiness.open"`,
		`operation="MarkReady"`,
		`state="open"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got %q", want, output)
		}
	}
}

func TestFatalLogsAndUsesInjectedExit(t *testing.T) {
	var (
		buffer   bytes.Buffer
		exitCode = -1
	)
	logger := newRuntimeLogger(&buffer, func(code int) {
		exitCode = code
	})

	logger.Fatal(
		"startup",
		"database initialization failed",
		errors.New("password=swordfish"),
		Event(EventStartupIncompatibility),
		Operation("InitDB"),
		ErrorType(EventStartupIncompatibility),
	)

	output := buffer.String()
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(output, "level=FATAL") {
		t.Fatalf("expected FATAL level in output, got %q", output)
	}
	if !strings.Contains(output, `operation="InitDB"`) {
		t.Fatalf("expected operation field in output, got %q", output)
	}
	if !strings.Contains(output, `event="startup.incompatibility"`) {
		t.Fatalf("expected startup event field in output, got %q", output)
	}
	if !strings.Contains(output, `error_type="startup.incompatibility"`) {
		t.Fatalf("expected startup error_type field in output, got %q", output)
	}
	if !strings.Contains(output, `error="password=[REDACTED]"`) {
		t.Fatalf("expected redacted error field in output, got %q", output)
	}
}
