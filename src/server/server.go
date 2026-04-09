// Package server provides the HTTP server, routing, and request handlers
// for the k8s-platform API.
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gat516/k8s-platform/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Server wraps the HTTP server and its dependencies.
type Server struct {
	httpServer *http.Server
	cfg        *config.Config
	metrics    *metrics
	// k8sClient is nil when running outside a cluster; handlers fall back to
	// mock/empty data in that case.
	k8sClient *kubernetes.Clientset
}

// New creates a Server configured with the provided Config.
// Routes are registered on creation; call ListenAndServe to start accepting connections.
func New(cfg *config.Config) *Server {
	m := newMetrics()
	s := &Server{cfg: cfg, metrics: m}

	// Attempt to build an in-cluster Kubernetes client. If this binary is running
	// outside a cluster (e.g. local dev) the error is logged and k8sClient stays
	// nil — each handler must tolerate a nil client.
	k8sCfg, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("k8s in-cluster config unavailable (running outside cluster?): %v — API endpoints will return empty/mock data", err)
	} else {
		cs, err := kubernetes.NewForConfig(k8sCfg)
		if err != nil {
			log.Printf("failed to create k8s client: %v — API endpoints will return empty/mock data", err)
		} else {
			s.k8sClient = cs
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.Handle("/metrics", metricsHandler())
	mux.HandleFunc("/api/v1/cluster", s.handleCluster)
	mux.HandleFunc("/api/v1/services", s.handleServices)
	mux.HandleFunc("/api/v1/resources", s.handleResources)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      s.requestLogger(s.corsMiddleware(mux)),
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

// corsMiddleware sets Access-Control-Allow-Origin for allowed browser origins
// (configured via CORS_ORIGINS env var) so the Vite dashboard can fetch the API.
// It also handles OPTIONS preflight requests required by browsers for cross-origin
// requests with non-simple headers.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); s.cfg.CORSOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
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
