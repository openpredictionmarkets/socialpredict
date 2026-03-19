package server

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// TestGracefulShutdown verifies that an http.Server can be started and then
// shut down cleanly via Shutdown(ctx) — the core mechanism added for issue #52.
// We test the http.Server lifecycle directly without involving the full router,
// which would require a database and full app initialisation.
func TestGracefulShutdown(t *testing.T) {
	srv := &http.Server{
		Addr:    "127.0.0.1:0", // OS-assigned port
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	started := make(chan struct{})

	go func() {
		close(started)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("unexpected server error: %v", err)
		}
	}()

	// Give the server a moment to start
	<-started
	time.Sleep(10 * time.Millisecond)

	// Shut down with a context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown returned unexpected error: %v", err)
	}
}

// TestGracefulShutdown_ContextTimeout verifies that Shutdown respects a deadline
// — if the context is already cancelled, Shutdown returns context.DeadlineExceeded
// or context.Canceled (not a fatal panic or hang).
func TestGracefulShutdown_ContextTimeout(t *testing.T) {
	srv := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	}

	go func() {
		srv.ListenAndServe() //nolint:errcheck
	}()
	time.Sleep(10 * time.Millisecond)

	// Immediately-expired context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond) // ensure deadline has passed

	err := srv.Shutdown(ctx)
	// Either nil (closed before timeout) or context error — both acceptable
	if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		t.Errorf("unexpected error from Shutdown with expired context: %v", err)
	}
}
