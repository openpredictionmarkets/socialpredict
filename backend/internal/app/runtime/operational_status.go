package runtime

import (
	"net/http"
	"sync/atomic"
)

// OperationalMetrics is the app-owned process metrics seam for first-alert signals.
type OperationalMetrics struct {
	requestFailuresTotal atomic.Uint64
}

func NewOperationalMetrics() *OperationalMetrics {
	return &OperationalMetrics{}
}

func (m *OperationalMetrics) ObserveRequestStatus(statusCode int) {
	if m == nil {
		return
	}

	if statusCode >= http.StatusInternalServerError {
		m.RecordRequestFailure()
	}
}

func (m *OperationalMetrics) RecordRequestFailure() {
	if m == nil {
		return
	}

	m.requestFailuresTotal.Add(1)
}

func (m *OperationalMetrics) Snapshot() OperationalSnapshot {
	if m == nil {
		return OperationalSnapshot{}
	}

	return OperationalSnapshot{
		RequestFailuresTotal: m.requestFailuresTotal.Load(),
	}
}

type OperationalSnapshot struct {
	RequestFailuresTotal uint64 `json:"requestFailuresTotal"`
}

// Middleware records process-local request counts for operator alerts.
func (m *OperationalMetrics) Middleware(next http.Handler) http.Handler {
	if next == nil {
		next = http.NotFoundHandler()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &operationalStatusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)
		if isOperationalProbePath(r) {
			return
		}
		m.ObserveRequestStatus(recorder.statusCode)
	})
}

func isOperationalProbePath(r *http.Request) bool {
	if r == nil || r.URL == nil {
		return false
	}

	switch r.URL.Path {
	case "/health", "/readyz", "/ops/status":
		return true
	default:
		return false
	}
}

type operationalStatusRecorder struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (w *operationalStatusRecorder) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *operationalStatusRecorder) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
	}
	return w.ResponseWriter.Write(data)
}
