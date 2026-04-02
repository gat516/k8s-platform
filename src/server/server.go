// Package server provides the HTTP server, routing, and request handlers
// for the kube-platform API. It is the primary interface between the
// Next.js frontend and the Kubernetes cluster.
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gat516/kube-platform/config"
)

// Server wraps the HTTP server and its dependencies.
type Server struct {
	httpServer *http.Server
	cfg        *config.Config
	metrics    *metrics
}

// New creates a Server configured with the provided Config.
// Routes are registered on creation; call ListenAndServe to start accepting connections.
func New(cfg *config.Config) *Server {
	m := newMetrics()
	s := &Server{cfg: cfg, metrics: m}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.Handle("/metrics", metricsHandler())

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      s.requestLogger(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// ListenAndServe starts the HTTP server and blocks until it returns an error.
// It returns http.ErrServerClosed on a clean shutdown.
func (s *Server) ListenAndServe() error {
	log.Printf("server starting on port %d (env: %s)", s.cfg.Port, s.cfg.Environment)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server, waiting up to the deadline in ctx
// for in-flight requests to complete before closing connections.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// ServeHTTP implements http.Handler, allowing the Server to be used directly
// in tests via httptest without starting a real TCP listener.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.httpServer.Handler.ServeHTTP(w, r)
}

// requestLogger wraps a handler and emits a structured log line for every request,
// including method, path, status, and elapsed duration.
func (s *Server) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		status := strconv.Itoa(rw.statusCode)

		s.metrics.httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		s.metrics.httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())

		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code written
// by a handler, since http.ResponseWriter does not expose it after the fact.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before delegating to the underlying writer.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
