package security

import (
	"bytes"
	"encoding/json"
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
	var seenLoggerRequestID string
	middleware := newRequestBoundaryMiddleware(logger.New(&buffer), now, func() string {
		return "req-123"
	})

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenRequestID = RequestIDFromContext(r.Context())
		seenLoggerRequestID = logger.RequestIDFromContext(r.Context())
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
	if seenLoggerRequestID != "req-123" {
		t.Fatalf("expected logger request ID in context, got %q", seenLoggerRequestID)
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

func TestRequestBoundaryMiddlewareRecoversPanicWithSanitizedCorrelatedFailure(t *testing.T) {
	var buffer bytes.Buffer

	timestamps := []time.Time{
		time.Unix(1_710_000_000, 0).UTC(),
		time.Unix(1_710_000_000, 0).UTC().Add(25 * time.Millisecond),
	}
	nowCalls := 0
	now := func() time.Time {
		current := timestamps[nowCalls]
		nowCalls++
		return current
	}

	middleware := newRequestBoundaryMiddleware(logger.New(&buffer), now, func() string {
		return "req-generated"
	})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("password=swordfish")
	}))

	req := httptest.NewRequest(http.MethodPost, "/v0/panic?token=abc123", nil)
	req.Header.Set(logger.RequestIDHeader, "req-client")
	req.Header.Set("Traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if got := rec.Header().Get(logger.RequestIDHeader); got != "req-client" {
		t.Fatalf("expected preserved request ID header, got %q", got)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected JSON content type, got %q", got)
	}

	var response struct {
		OK     bool   `json:"ok"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode panic response: %v", err)
	}
	if response.OK || response.Reason != "INTERNAL_ERROR" {
		t.Fatalf("expected sanitized internal error response, got %+v", response)
	}
	if strings.Contains(rec.Body.String(), "swordfish") {
		t.Fatalf("expected panic detail to stay out of client body, got %q", rec.Body.String())
	}

	output := buffer.String()
	for _, want := range []string{
		"level=ERROR",
		`component="middleware"`,
		`msg="request panic recovered"`,
		`operation="RequestBoundary"`,
		`request_id="req-client"`,
		`trace_id="4bf92f3577b34da6a3ce929d0e0e4736"`,
		`span_id="00f067aa0ba902b7"`,
		`status_code="500"`,
		`duration_ms="25"`,
		`error_type="exception.panic"`,
		`exception_recorded="true"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got %q", want, output)
		}
	}
	for _, forbidden := range []string{"swordfish", "abc123"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("expected sensitive value %q to stay out of logs, got %q", forbidden, output)
		}
	}
}

func TestRequestBoundaryMiddlewareClassifiesRuntimeFailureStatus(t *testing.T) {
	var buffer bytes.Buffer

	timestamps := []time.Time{
		time.Unix(1_710_000_000, 0).UTC(),
		time.Unix(1_710_000_000, 0).UTC().Add(10 * time.Millisecond),
	}
	nowCalls := 0
	now := func() time.Time {
		current := timestamps[nowCalls]
		nowCalls++
		return current
	}

	middleware := newRequestBoundaryMiddleware(logger.New(&buffer), now, func() string {
		return "req-429"
	})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		WriteRateLimited(w, RuntimeReasonLoginRateLimited)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v0/login", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", rec.Code)
	}

	var response struct {
		OK     bool   `json:"ok"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode rate limit response: %v", err)
	}
	if response.OK || response.Reason != RuntimeReasonLoginRateLimited {
		t.Fatalf("expected login rate limit reason, got %+v", response)
	}

	output := buffer.String()
	for _, want := range []string{
		"level=WARN",
		`status_code="429"`,
		`error_type="http.rate_limited"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got %q", want, output)
		}
	}
	if strings.Contains(output, RuntimeReasonLoginRateLimited) {
		t.Fatalf("expected public reason to stay out of telemetry fields, got %q", output)
	}
}
