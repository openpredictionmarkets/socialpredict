package mcpserver

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	appruntime "socialpredict/internal/app/runtime"
	"socialpredict/logger"
)

func BuildHTTPHandler(rt *Runtime, readiness *appruntime.Readiness, probe appruntime.ServingProbe) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("live"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := probe.Ready(ctx); err != nil {
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("not ready"))
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	server := rt.MCPServer()
	mux.Handle("/mcp", mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{Stateless: true}))
	return mux
}

func StartHTTP(handler http.Handler, readiness *appruntime.Readiness, shutdownConfig appruntime.ShutdownConfig) {
	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "8081"
	}
	address := ":" + port
	httpServer := &http.Server{Addr: address, Handler: handler}

	errs := make(chan error, 1)
	go func() {
		errs <- httpServer.ListenAndServe()
	}()
	logger.Info("mcpserver", "MCP HTTP server listening", logger.Operation("Start"), logger.Address(address))

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case err := <-errs:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("mcpserver", "http server exited unexpectedly", err, logger.Operation("ListenAndServe"), logger.Address(address))
		}
	case <-signals:
		if readiness != nil {
			readiness.MarkNotReady()
		}
		shutdownConfig = appruntime.NormalizeShutdownConfig(shutdownConfig)
		time.Sleep(shutdownConfig.ReadinessDrainWindow)
		ctx, cancel := context.WithTimeout(context.Background(), shutdownConfig.ShutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Fatal("mcpserver", "graceful shutdown failed", err, logger.Operation("Shutdown"), logger.Address(address))
		}
		if err := <-errs; err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("mcpserver", "http server exited unexpectedly during shutdown", err, logger.Operation("ListenAndServe"), logger.Address(address))
		}
	}
}
