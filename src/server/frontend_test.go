package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gat516/k8s-platform/config"
	"github.com/prometheus/client_golang/prometheus"
)

func TestDashboardFiles(t *testing.T) {
	// New registers metrics globally; keep this server separate from other tests.
	registerer := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	t.Cleanup(func() { prometheus.DefaultRegisterer = registerer })
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "assets"), 0700); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"index.html":           "<!doctype html><title>CI/CD Dashboard</title>",
		"assets/app-abc123.js": "console.log('dashboard')",
		"private.txt":          "must not be public",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	srv := New(&config.Config{DashboardStaticDir: dir, Version: "release123"})
	for _, tc := range []struct {
		method, path string
		status       int
		body, cache  string
	}{
		{http.MethodGet, "/", 200, "CI/CD Dashboard", "no-cache"},
		{http.MethodHead, "/", 200, "", "no-cache"},
		{http.MethodGet, "/assets/app-abc123.js", 200, "console.log", "immutable"},
		{http.MethodGet, "/assets/missing.js", 404, "", ""},
		{http.MethodGet, "/assets/", 404, "", ""},
		{http.MethodGet, "/private.txt", 404, "", ""},
		{http.MethodGet, "/api/v1/unknown", 404, "", ""},
		{http.MethodGet, "/api/v1/cluster", 404, "", ""},
		{http.MethodGet, "/health", 200, `"version":"release123"`, ""},
		{http.MethodGet, "/api/v1/dashboard", 200, `"pipeline"`, "no-store"},
		{http.MethodPost, "/", 405, "", ""},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
			if rec.Code != tc.status || !strings.Contains(rec.Body.String(), tc.body) {
				t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
			}
			if tc.method == http.MethodHead && rec.Body.Len() != 0 {
				t.Fatal("HEAD returned a body")
			}
			if !strings.Contains(rec.Header().Get("Cache-Control"), tc.cache) {
				t.Fatalf("Cache-Control=%q, want %q", rec.Header().Get("Cache-Control"), tc.cache)
			}
		})
	}
}

func TestDashboardFilesCannotReadOutsideAssets(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "assets"), 0700); err != nil {
		t.Fatal(err)
	}
	private := filepath.Join(dir, "private.txt")
	if err := os.WriteFile(private, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(private, filepath.Join(dir, "assets", "link.txt")); err != nil {
		t.Fatal(err)
	}
	handler := dashboardFiles(dir)
	for _, path := range []string{"/private.txt", "/assets/../private.txt", "/assets/%2e%2e/private.txt", "/assets/link.txt", "/assets/"} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != 404 || strings.Contains(rec.Body.String(), "private") {
			t.Fatalf("path %s: %d %q", path, rec.Code, rec.Body.String())
		}
	}
}
