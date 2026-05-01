package logger

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContextWithRequestCorrelation(t *testing.T) {
	ctx := ContextWithRequestCorrelation(
		nil,
		" req-123 ",
		"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
	)

	if got := RequestIDFromContext(ctx); got != "req-123" {
		t.Fatalf("request id = %q, want req-123", got)
	}
	if got := TraceIDFromContext(ctx); got != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("trace id = %q", got)
	}
	if got := SpanIDFromContext(ctx); got != "00f067aa0ba902b7" {
		t.Fatalf("span id = %q", got)
	}
	if got := RequestIDFromContext(nil); got != "" {
		t.Fatalf("nil context request id = %q, want empty", got)
	}
}

func TestRequestLoggingMiddlewarePropagatesCorrelationAndLogsSuccess(t *testing.T) {
	buffer := useStandardForTest(t)

	handler := RequestLoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := RequestIDFromContext(r.Context()); got != "req-client" {
			t.Fatalf("request id in context = %q", got)
		}
		if got := TraceIDFromContext(r.Context()); got != "4bf92f3577b34da6a3ce929d0e0e4736" {
			t.Fatalf("trace id in context = %q", got)
		}
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v0/test?token=secret", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set(RequestIDHeader, "req-client")
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if got := rec.Header().Get(RequestIDHeader); got != "req-client" {
		t.Fatalf("response request id = %q", got)
	}
	output := buffer.String()
	for _, want := range []string{
		"level=INFO",
		`component="HTTPRequest"`,
		`operation="Complete"`,
		`request_id=req-client`,
		`trace_id=4bf92f3577b34da6a3ce929d0e0e4736`,
		`span_id=00f067aa0ba902b7`,
		`status=201`,
		`remote_addr=10.0.0.1:1234`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output %q", want, output)
		}
	}
	if strings.Contains(output, "secret") {
		t.Fatalf("query secret leaked into log output: %q", output)
	}
}

func TestRequestLoggingMiddlewareLogsSeverityByStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantLevel  string
	}{
		{name: "client error", statusCode: http.StatusNotFound, wantLevel: "level=WARN"},
		{name: "server error", statusCode: http.StatusInternalServerError, wantLevel: "level=ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := useStandardForTest(t)
			handler := RequestLoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))

			req := httptest.NewRequest(http.MethodGet, "/status", nil)
			req.Header.Set(RequestIDHeader, "req-status")
			handler.ServeHTTP(httptest.NewRecorder(), req)

			output := buffer.String()
			if !strings.Contains(output, tt.wantLevel) {
				t.Fatalf("expected %q in output %q", tt.wantLevel, output)
			}
		})
	}
}

func TestRequestLoggingMiddlewareNilHandlerAndCanceledRequest(t *testing.T) {
	t.Run("nil handler writes 500", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RequestLoggingMiddleware(nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nil", nil))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rec.Code)
		}
	})

	t.Run("canceled request logs 499 when no header was written", func(t *testing.T) {
		buffer := useStandardForTest(t)
		handler := RequestLoggingMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		req := httptest.NewRequest(http.MethodGet, "/canceled", nil).WithContext(ctx)
		req.Header.Set(RequestIDHeader, "req-canceled")

		handler.ServeHTTP(httptest.NewRecorder(), req)

		output := buffer.String()
		if !strings.Contains(output, "level=WARN") || !strings.Contains(output, "status=499") {
			t.Fatalf("expected canceled request warning with 499, got %q", output)
		}
	})
}

func TestRemoteAddressSelection(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	if got := remoteAddress(req); got != "10.0.0.1:1234" {
		t.Fatalf("remoteAddress = %q", got)
	}

	req.Header.Set("X-Real-Ip", "198.51.100.11")
	if got := remoteAddress(req); got != "198.51.100.11" {
		t.Fatalf("remoteAddress with real ip = %q", got)
	}

	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	if got := remoteAddress(req); got != "203.0.113.10" {
		t.Fatalf("remoteAddress with forwarded for = %q", got)
	}

	if got := remoteAddress(nil); got != "" {
		t.Fatalf("nil remoteAddress = %q", got)
	}
}
