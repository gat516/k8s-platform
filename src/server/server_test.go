package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gat516/k8s-platform/config"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	cfg := &config.Config{
		Port:               8080,
		Environment:        "local",
		AllowedGitHubUsers: map[string]struct{}{},
		Version:            "dev",
	}
	return New(cfg)
}

func TestServerRoutes(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
	}{
		{
			name:           "GET /health returns 200",
			method:         http.MethodGet,
			path:           "/health",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "GET /metrics returns 200",
			method:         http.MethodGet,
			path:           "/metrics",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "GET /unknown returns 404",
			method:         http.MethodGet,
			path:           "/unknown",
			wantStatusCode: http.StatusNotFound,
		},
	}

	srv := newTestServer(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			srv.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatusCode {
				t.Errorf("status code: got %d, want %d", rec.Code, tc.wantStatusCode)
			}
		})
	}
}

func TestResponseWriterCapturesStatusCode(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	rw.WriteHeader(http.StatusTeapot)

	if rw.statusCode != http.StatusTeapot {
		t.Errorf("statusCode: got %d, want %d", rw.statusCode, http.StatusTeapot)
	}
}
