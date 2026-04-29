package runtime

import "sync/atomic"

// Readiness tracks whether the process should receive traffic.
type Readiness struct {
	ready atomic.Bool
}

func NewReadiness() *Readiness {
	return &Readiness{}
}

func (r *Readiness) MarkReady() {
	if r == nil {
		return
	}
	r.ready.Store(true)
}

func (r *Readiness) MarkNotReady() {
	if r == nil {
		return
	}
	r.ready.Store(false)
}

func (r *Readiness) Ready() bool {
	if r == nil {
		return false
	}
	return r.ready.Load()
}
