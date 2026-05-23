package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	RequestIDHeader           = "X-Request-Id"
	clientClosedRequestStatus = 499

	requestIDContextKey contextKey = "request_id"
	traceIDContextKey   contextKey = "trace_id"
	spanIDContextKey    contextKey = "span_id"
)

type contextKey string

// RequestIDFromContext retrieves the current request id from context when present.
func RequestIDFromContext(ctx context.Context) string {
	return contextValue(ctx, requestIDContextKey)
}

// TraceIDFromContext retrieves the current trace id from context when present.
func TraceIDFromContext(ctx context.Context) string {
	return contextValue(ctx, traceIDContextKey)
}

// SpanIDFromContext retrieves the current span id from context when present.
func SpanIDFromContext(ctx context.Context) string {
	return contextValue(ctx, spanIDContextKey)
}

func contextValue(ctx context.Context, key contextKey) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(key).(string)
	return value
}

// ContextWithRequestCorrelation attaches the request-boundary correlation fields
// used by runtime logging and legacy handler diagnostics.
func ContextWithRequestCorrelation(ctx context.Context, requestID, traceparent string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	requestID = strings.TrimSpace(requestID)
	if requestID != "" {
		ctx = context.WithValue(ctx, requestIDContextKey, requestID)
	}

	traceID, spanID := parseTraceparent(traceparent)
	if traceID != "" {
		ctx = context.WithValue(ctx, traceIDContextKey, traceID)
	}
	if spanID != "" {
		ctx = context.WithValue(ctx, spanIDContextKey, spanID)
	}

	return ctx
}

// RequestLoggingMiddleware logs one completion line per request and propagates
// stable request/correlation identifiers at the runtime boundary.
//
// Deprecated: production HTTP wiring uses security.RequestBoundaryMiddleware.
// Keep this helper for compatibility tests only; do not add it to server wiring.
func RequestLoggingMiddleware(next http.Handler) http.Handler {
	if next == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "handler unavailable", http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := strings.TrimSpace(r.Header.Get(RequestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}

		traceID, spanID := parseTraceparent(r.Header.Get("traceparent"))
		ctx := ContextWithRequestCorrelation(r.Context(), requestID, r.Header.Get("traceparent"))

		r = r.WithContext(ctx)

		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		recorder.Header().Set(RequestIDHeader, requestID)

		next.ServeHTTP(recorder, r)

		statusCode := recorder.statusCode
		canceled := isRequestCanceled(r.Context().Err())
		if canceled && !recorder.wroteHeader {
			statusCode = clientClosedRequestStatus
		}

		duration := time.Since(start).Milliseconds()
		message := fmt.Sprintf(
			"method=%s path=%s status=%d duration_ms=%d request_id=%s trace_id=%s span_id=%s remote_addr=%s",
			r.Method,
			r.URL.Path,
			statusCode,
			duration,
			requestID,
			orDash(traceID),
			orDash(spanID),
			orDash(remoteAddress(r)),
		)

		switch {
		case canceled && !recorder.wroteHeader:
			LogWarn("HTTPRequest", "Canceled", message)
		case statusCode >= http.StatusInternalServerError:
			LogError("HTTPRequest", "Complete", errors.New(message))
		case statusCode >= http.StatusBadRequest:
			LogWarn("HTTPRequest", "Complete", message)
		default:
			LogInfo("HTTPRequest", "Complete", message)
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (w *statusRecorder) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusRecorder) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
	}
	return w.ResponseWriter.Write(b)
}

func isRequestCanceled(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func newRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

func parseTraceparent(header string) (traceID string, spanID string) {
	parts := strings.Split(strings.TrimSpace(header), "-")
	if len(parts) != 4 {
		return "", ""
	}

	if len(parts[1]) == 32 {
		traceID = parts[1]
	}
	if len(parts[2]) == 16 {
		spanID = parts[2]
	}

	return traceID, spanID
}

func orDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func remoteAddress(r *http.Request) string {
	if r == nil {
		return ""
	}
	if forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		return forwardedFor
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-Ip")); realIP != "" {
		return realIP
	}
	return r.RemoteAddr
}
