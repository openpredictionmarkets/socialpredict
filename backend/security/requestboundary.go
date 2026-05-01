package security

import (
	"bufio"
	"context"
	crand "crypto/rand"
	"encoding/hex"
	stderrors "errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"socialpredict/logger"
)

const (
	requestIDHeader = logger.RequestIDHeader
)

type requestIDContextKey struct{}

// RequestBoundaryMiddleware attaches stable request correlation and completion logging at the HTTP boundary.
func RequestBoundaryMiddleware() func(http.Handler) http.Handler {
	return RequestBoundaryMiddlewareWithProxyTrust(false)
}

// RequestBoundaryMiddlewareWithProxyTrust attaches request-boundary behavior with explicit forwarded-header trust.
func RequestBoundaryMiddlewareWithProxyTrust(trustProxyHeaders bool) func(http.Handler) http.Handler {
	return newRequestBoundaryMiddleware(logger.Standard(), time.Now, defaultRequestID, NewClientIdentityExtractor(trustProxyHeaders))
}

// RequestIDFromContext returns the boundary request identifier when middleware has attached one.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func newRequestBoundaryMiddleware(runtimeLogger *logger.RuntimeLogger, now func() time.Time, newRequestID func() string, extractors ...ClientIdentityExtractor) func(http.Handler) http.Handler {
	if runtimeLogger == nil {
		runtimeLogger = logger.Standard()
	}
	if now == nil {
		now = time.Now
	}
	if newRequestID == nil {
		newRequestID = defaultRequestID
	}
	clientIdentity := NewClientIdentityExtractor(false)
	if len(extractors) > 0 {
		clientIdentity = extractors[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := now()
			requestID := incomingRequestID(r, newRequestID)

			w.Header().Set(requestIDHeader, requestID)
			ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
			ctx = logger.ContextWithRequestCorrelation(ctx, requestID, r.Header.Get("Traceparent"))
			r = r.WithContext(ctx)

			recorder := &requestBoundaryWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			defer func() {
				if recovered := recover(); recovered != nil {
					if !recorder.wroteHeader {
						writeRecoveredPanicResponse(recorder)
					} else {
						recorder.statusCode = http.StatusInternalServerError
					}

					runtimeLogger.Error(
						"middleware",
						"request panic recovered",
						stderrors.New("panic recovered"),
						append(
							requestBoundaryFields(r, requestID, recorder.statusCode, now().Sub(startedAt), RuntimeErrorTypePanic, clientIdentity),
							logger.ExceptionRecorded(),
						)...,
					)
				}
			}()

			next.ServeHTTP(recorder, r)

			fields := requestBoundaryFields(r, requestID, recorder.statusCode, now().Sub(startedAt), runtimeFailureErrorType(recorder.statusCode), clientIdentity)

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

func requestBoundaryFields(r *http.Request, requestID string, statusCode int, duration time.Duration, errorType string, extractors ...ClientIdentityExtractor) []logger.Field {
	clientIdentity := NewClientIdentityExtractor(false)
	if len(extractors) > 0 {
		clientIdentity = extractors[0]
	}
	fields := []logger.Field{
		logger.Operation("RequestBoundary"),
		logger.RequestID(requestID),
		logger.Method(requestMethod(r)),
		logger.Path(requestPath(r)),
		logger.StatusCode(statusCode),
		logger.DurationMS(duration),
	}
	if address := strings.TrimSpace(clientIdentity.Extract(r)); address != "" {
		fields = append(fields, logger.Address(address))
	}
	fields = append(fields, logger.TraceContextFromTraceparent(r.Header.Get("Traceparent"))...)
	if strings.TrimSpace(errorType) != "" {
		fields = append(fields, logger.ErrorType(errorType))
	}

	return fields
}

func writeRecoveredPanicResponse(w http.ResponseWriter) {
	WriteInternalServerError(w)
}

func runtimeFailureErrorType(statusCode int) string {
	errorType, ok := RuntimeFailureErrorType(statusCode)
	if !ok {
		return ""
	}
	return errorType
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
	statusCode  int
	wroteHeader bool
}

func (w *requestBoundaryWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *requestBoundaryWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *requestBoundaryWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
	}
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
