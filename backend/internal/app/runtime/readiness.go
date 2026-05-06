package runtime

import (
	"sync/atomic"

	"socialpredict/logger"
)

const (
	ReadinessEventOpen   = logger.EventReadinessOpen
	ReadinessEventClosed = logger.EventReadinessClosed

	readinessComponent   = "runtime"
	readinessOperation   = "ReadinessTransition"
	readinessStateOpen   = "open"
	readinessStateClosed = "closed"
)

// Readiness tracks whether the process should receive traffic.
type Readiness struct {
	ready  atomic.Bool
	logger *logger.RuntimeLogger
}

func NewReadiness() *Readiness {
	return NewReadinessWithLogger(logger.Standard())
}

// NewReadinessWithLogger creates a readiness gate with an explicit runtime logger.
func NewReadinessWithLogger(runtimeLogger *logger.RuntimeLogger) *Readiness {
	if runtimeLogger == nil {
		runtimeLogger = logger.Standard()
	}
	return &Readiness{logger: runtimeLogger}
}

func (r *Readiness) MarkReady() {
	if r == nil {
		return
	}
	if r.ready.Swap(true) {
		return
	}
	r.logTransition(ReadinessEventOpen, readinessStateOpen)
}

func (r *Readiness) MarkNotReady() {
	if r == nil {
		return
	}
	if !r.ready.Swap(false) {
		return
	}
	r.logTransition(ReadinessEventClosed, readinessStateClosed)
}

func (r *Readiness) Ready() bool {
	if r == nil {
		return false
	}
	return r.ready.Load()
}

func (r *Readiness) logTransition(event, state string) {
	runtimeLogger := r.logger
	if runtimeLogger == nil {
		runtimeLogger = logger.Standard()
	}
	runtimeLogger.Info(
		readinessComponent,
		"readiness state changed",
		logger.Event(event),
		logger.Operation(readinessOperation),
		logger.State(state),
	)
}
