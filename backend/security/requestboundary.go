package security

import (
	"bufio"
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"socialpredict/logger"
)

const requestIDHeader = "X-Request-Id"

type requestIDContextKey struct{}

// RequestBoundaryMiddleware attaches stable request correlation and completion logging at the HTTP boundary.
func RequestBoundaryMiddleware() func(http.Handler) http.Handler {
	return newRequestBoundaryMiddleware(logger.Standard(), time.Now, defaultRequestID)
}

// RequestIDFromContext returns the boundary request identifier when middleware has attached one.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func newRequestBoundaryMiddleware(runtimeLogger *logger.RuntimeLogger, now func() time.Time, newRequestID func() string) func(http.Handler) http.Handler {
	if runtimeLogger == nil {
		runtimeLogger = logger.Standard()
	}
	if now == nil {
		now = time.Now
	}
	if newRequestID == nil {
		newRequestID = defaultRequestID
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := now()
			requestID := incomingRequestID(r, newRequestID)

			w.Header().Set(requestIDHeader, requestID)
			r = r.WithContext(context.WithValue(r.Context(), requestIDContextKey{}, requestID))

			recorder := &requestBoundaryWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(recorder, r)

			fields := []logger.Field{
				logger.Operation("RequestBoundary"),
				logger.RequestID(requestID),
				logger.Method(requestMethod(r)),
				logger.Path(requestPath(r)),
				logger.StatusCode(recorder.statusCode),
				logger.DurationMS(now().Sub(startedAt)),
			}
			if address := strings.TrimSpace(getClientIP(r)); address != "" {
				fields = append(fields, logger.Address(address))
			}
			fields = append(fields, logger.TraceContextFromTraceparent(r.Header.Get("Traceparent"))...)

			switch {
			case recorder.statusCode >= http.StatusInternalServerError:
				runtimeLogger.Error("middleware", "request failed", nil, fields...)
			case recorder.statusCode >= http.StatusBadRequest:
				runtimeLogger.Warn("middleware", "request completed with client error", fields...)
			default:
				runtimeLogger.Info("middleware", "request completed", fields...)
			}
		})
	}
}

func defaultRequestID() string {
	var raw [16]byte
	if _, err := crand.Read(raw[:]); err == nil {
		return hex.EncodeToString(raw[:])
	}

	return strconv.FormatInt(time.Now().UnixNano(), 16)
}

func incomingRequestID(r *http.Request, newRequestID func() string) string {
	if r != nil {
		if requestID := strings.TrimSpace(r.Header.Get(requestIDHeader)); requestID != "" {
			return requestID
		}
	}

	return newRequestID()
}

func requestMethod(r *http.Request) string {
	if r == nil {
		return ""
	}
	return strings.TrimSpace(r.Method)
}

func requestPath(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}

	path := strings.TrimSpace(r.URL.EscapedPath())
	if path == "" {
		return "/"
	}
	return path
}

type requestBoundaryWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *requestBoundaryWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *requestBoundaryWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *requestBoundaryWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

func (w *requestBoundaryWriter) Flush() {
	flusher, ok := w.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}
	flusher.Flush()
}

func (w *requestBoundaryWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func (w *requestBoundaryWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return pusher.Push(target, opts)
}

func (w *requestBoundaryWriter) ReadFrom(reader io.Reader) (int64, error) {
	if readerFrom, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return readerFrom.ReadFrom(reader)
	}
	return io.Copy(w.ResponseWriter, reader)
}

func (w *requestBoundaryWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
