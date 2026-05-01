package security

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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
	}, NewClientIdentityExtractor(true))

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

func TestRequestBoundaryMiddlewareIgnoresForwardedAddressWithoutProxyTrust(t *testing.T) {
	var buffer bytes.Buffer

	timestamps := []time.Time{
		time.Unix(1_710_000_000, 0).UTC(),
		time.Unix(1_710_000_000, 0).UTC().Add(time.Millisecond),
	}
	nowCalls := 0
	now := func() time.Time {
		current := timestamps[nowCalls]
		nowCalls++
		return current
	}

	middleware := newRequestBoundaryMiddleware(logger.New(&buffer), now, func() string {
		return "req-direct"
	}, NewClientIdentityExtractor(false))
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/observed", nil)
	req.RemoteAddr = "10.0.0.10:12345"
	req.Header.Set("X-Forwarded-For", "198.51.100.10")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := buffer.String()
	if !strings.Contains(output, `address="10.0.0.10"`) {
		t.Fatalf("expected untrusted proxy headers to be ignored, got %q", output)
	}
	if strings.Contains(output, `address="198.51.100.10"`) {
		t.Fatalf("expected forwarded address to stay out of client identity log when proxy trust is disabled, got %q", output)
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

func TestRequestBoundaryMiddlewareClassifiesSharedRuntimeFailures(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		path          string
		writeFailure  func(http.ResponseWriter)
		wantStatus    int
		wantReason    string
		wantLevel     string
		wantErrorType string
	}{
		{
			name:   "shared 405 method not allowed",
			method: http.MethodPost,
			path:   "/swagger",
			writeFailure: func(w http.ResponseWriter) {
				WriteMethodNotAllowed(w)
			},
			wantStatus:    http.StatusMethodNotAllowed,
			wantReason:    RuntimeReasonMethodNotAllowed,
			wantLevel:     "level=WARN",
			wantErrorType: RuntimeErrorTypeMethodNotAllowed,
		},
		{
			name:   "shared 429 login rate limited",
			method: http.MethodPost,
			path:   "/v0/login",
			writeFailure: func(w http.ResponseWriter) {
				WriteRateLimited(w, RuntimeReasonLoginRateLimited)
			},
			wantStatus:    http.StatusTooManyRequests,
			wantReason:    RuntimeReasonLoginRateLimited,
			wantLevel:     "level=WARN",
			wantErrorType: RuntimeErrorTypeRateLimited,
		},
		{
			name:   "shared sanitized 500 internal error",
			method: http.MethodGet,
			path:   "/v0/internal",
			writeFailure: func(w http.ResponseWriter) {
				WriteInternalServerError(w)
			},
			wantStatus:    http.StatusInternalServerError,
			wantReason:    RuntimeReasonInternalError,
			wantLevel:     "level=ERROR",
			wantErrorType: RuntimeErrorTypeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				return "req-runtime"
			})
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				tt.writeFailure(w)
			}))

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("expected JSON content type, got %q", got)
			}

			var response RuntimeFailureResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode runtime failure response: %v", err)
			}
			if response.OK || response.Reason != tt.wantReason {
				t.Fatalf("expected reason %q, got %+v", tt.wantReason, response)
			}

			output := buffer.String()
			for _, want := range []string{
				tt.wantLevel,
				`status_code="` + strconv.Itoa(tt.wantStatus) + `"`,
				`error_type="` + tt.wantErrorType + `"`,
			} {
				if !strings.Contains(output, want) {
					t.Fatalf("expected %q in output, got %q", want, output)
				}
			}
			if strings.Contains(output, tt.wantReason) {
				t.Fatalf("expected public reason to stay out of telemetry fields, got %q", output)
			}
		})
	}
}
