package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gat516/kube-platform/config"
)

// newHealthServer creates a minimal *Server for testing the health handler.
// It avoids calling New() — and therefore newMetrics() — so it does not
// register Prometheus metrics. This prevents a duplicate-registration panic
// when this test binary also runs TestServerRoutes (which calls New() via
// newTestServer).
func newHealthServer(version string) *Server {
	return &Server{cfg: &config.Config{Version: version}}
}

func TestHandleHealth(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		wantStatusCode int
		wantStatus     string
	}{
		{
			name:           "GET returns 200 with ok status",
			method:         http.MethodGet,
			wantStatusCode: http.StatusOK,
			wantStatus:     "ok",
		},
		{
			name:           "POST returns 405",
			method:         http.MethodPost,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE returns 405",
			method:         http.MethodDelete,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	s := newHealthServer("dev")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/health", nil)
			rec := httptest.NewRecorder()

			s.handleHealth(rec, req)

			if rec.Code != tc.wantStatusCode {
				t.Errorf("status code: got %d, want %d", rec.Code, tc.wantStatusCode)
			}

			if tc.wantStatus != "" {
				var body healthResponse
				if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if body.Status != tc.wantStatus {
					t.Errorf("body.status: got %q, want %q", body.Status, tc.wantStatus)
				}
			}
		})
	}
}

func TestHandleHealthContentType(t *testing.T) {
	s := newHealthServer("dev")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.handleHealth(rec, req)

	got := rec.Header().Get("Content-Type")
	want := "application/json"
	if got != want {
		t.Errorf("Content-Type: got %q, want %q", got, want)
	}
}

func TestHandleHealthVersion(t *testing.T) {
	s := newHealthServer("abc1234")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.handleHealth(rec, req)

	var body healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body.Version != "abc1234" {
		t.Errorf("body.version: got %q, want %q", body.Version, "abc1234")
	}
}
