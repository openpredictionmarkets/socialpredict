package runtime

import (
	"bytes"
	"strings"
	"testing"

	"socialpredict/logger"
)

func TestReadinessDefaultsNotReady(t *testing.T) {
	readiness := NewReadiness()

	if readiness.Ready() {
		t.Fatalf("expected new readiness gate to start not ready")
	}
}

func TestReadinessTransitions(t *testing.T) {
	readiness := NewReadiness()

	readiness.MarkReady()
	if !readiness.Ready() {
		t.Fatalf("expected readiness gate to report ready after MarkReady")
	}

	readiness.MarkNotReady()
	if readiness.Ready() {
		t.Fatalf("expected readiness gate to report not ready after MarkNotReady")
	}
}

func TestReadinessTransitionsEmitStableOperationalEvents(t *testing.T) {
	var buffer bytes.Buffer
	readiness := NewReadinessWithLogger(logger.New(&buffer))

	readiness.MarkReady()
	readiness.MarkReady()
	readiness.MarkNotReady()
	readiness.MarkNotReady()

	output := buffer.String()
	for _, want := range []string{
		`component="runtime"`,
		`operation="ReadinessTransition"`,
		`event="readiness.open"`,
		`state="open"`,
		`event="readiness.closed"`,
		`state="closed"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got %q", want, output)
		}
	}
	if strings.Count(output, `event="readiness.open"`) != 1 {
		t.Fatalf("expected one readiness-open event, got %q", output)
	}
	if strings.Count(output, `event="readiness.closed"`) != 1 {
		t.Fatalf("expected one readiness-closed event, got %q", output)
	}
}
