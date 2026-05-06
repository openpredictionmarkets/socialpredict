package runtime

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOperationalMetricsMiddlewareCountsServerFailuresOnly(t *testing.T) {
	metrics := NewOperationalMetrics()

	handler := metrics.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/server-error":
			w.WriteHeader(http.StatusInternalServerError)
		case "/client-error":
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))

	for _, path := range []string{"/ok", "/client-error", "/server-error", "/server-error"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	snapshot := metrics.Snapshot()
	if snapshot.RequestFailuresTotal != 2 {
		t.Fatalf("requestFailuresTotal = %d, want 2", snapshot.RequestFailuresTotal)
	}
}

func TestOperationalMetricsMiddlewareSkipsOperatorProbeFailures(t *testing.T) {
	metrics := NewOperationalMetrics()

	handler := metrics.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))

	for _, path := range []string{"/health", "/readyz", "/ops/status"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	if snapshot := metrics.Snapshot(); snapshot.RequestFailuresTotal != 0 {
		t.Fatalf("requestFailuresTotal = %d, want 0 for operator probes", snapshot.RequestFailuresTotal)
	}
}

func TestOperationalSnapshotJSONUsesOperatorFacingCounterName(t *testing.T) {
	metrics := NewOperationalMetrics()
	metrics.RecordRequestFailure()

	payload, err := json.Marshal(metrics.Snapshot())
	if err != nil {
		t.Fatalf("marshal operational snapshot: %v", err)
	}

	if got := string(payload); got != `{"requestFailuresTotal":1}` {
		t.Fatalf("unexpected operational snapshot JSON: %s", got)
	}
}
