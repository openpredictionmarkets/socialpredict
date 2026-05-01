package runtime

import "testing"

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
