package security

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"socialpredict/logger"
)

func TestRequestBoundaryMiddlewareAddsContextHeaderAndStructuredLog(t *testing.T) {
	var buffer bytes.Buffer

	timestamps := []time.Time{
		time.Unix(1_710_000_000, 0).UTC(),
		time.Unix(1_710_000_000, 0).UTC().Add(1500 * time.Millisecond),
	}
	nowCalls := 0
	now := func() time.Time {
		current := timestamps[nowCalls]
		nowCalls++
		return current
	}

	var seenRequestID string
	middleware := newRequestBoundaryMiddleware(logger.New(&buffer), now, func() string {
		return "req-123"
	})

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenRequestID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/observed?token=abc123", nil)
	req.Header.Set("Traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	req.Header.Set("X-Forwarded-For", "198.51.100.10")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if seenRequestID != "req-123" {
		t.Fatalf("expected request ID in context, got %q", seenRequestID)
	}
	if got := rec.Header().Get(requestIDHeader); got != "req-123" {
		t.Fatalf("expected response request ID header req-123, got %q", got)
	}

	output := buffer.String()
	for _, want := range []string{
		"level=INFO",
		`component="middleware"`,
		`msg="request completed"`,
		`operation="RequestBoundary"`,
		`request_id="req-123"`,
		`trace_id="4bf92f3577b34da6a3ce929d0e0e4736"`,
		`span_id="00f067aa0ba902b7"`,
		`trace_flags="01"`,
		`method="GET"`,
		`path="/observed"`,
		`status_code="201"`,
		`duration_ms="1500"`,
		`address="198.51.100.10"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got %q", want, output)
		}
	}
	if strings.Contains(output, "abc123") {
		t.Fatalf("expected query secrets to stay out of request log output, got %q", output)
	}
}
