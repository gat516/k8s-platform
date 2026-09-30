// Package config loads and validates runtime configuration from environment variables.
package config

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Config holds all runtime configuration for the API server.
type Config struct {
	Application       string
	ActionsRepository string
	ActionsWorkflow   string
	ActionsBranch     string
	ActionsToken      string
	ServiceHealthURL  string
	// DashboardStaticDir is the production Vite build directory. Empty leaves
	// frontend serving to the local Vite development server.
	DashboardStaticDir string
	// Port is the TCP port the HTTP server listens on.
	Port int

	// AllowedGitHubUsers is the set of GitHub usernames permitted to access
	// authenticated platform endpoints. Populated from ALLOWED_GITHUB_USERS
	// (comma-separated). Empty means no authenticated routes are accessible.
	AllowedGitHubUsers map[string]struct{}

	// Environment identifies the deployment context ("local", "production").
	Environment string

	// Version is the binary version string, injected at build time via -ldflags.
	// Defaults to "dev" when built without ldflags (local development).
	Version string

	// CORSOrigins is the set of allowed origins for separately hosted browser
	// clients. Populated from CORS_ORIGINS (comma-separated).
	// Defaults to allowing localhost:3000 for local development.
	CORSOrigins map[string]bool

	// PrometheusURL is the base URL of the in-cluster Prometheus instance used
	// to query node-level resource metrics. Defaults to the standard
	// kube-prometheus-stack service address.
	PrometheusURL string
}

// Load reads configuration from environment variables and returns a validated Config.
// Missing optional variables fall back to documented defaults.
// Returns an error if any required variable is absent or invalid.
func Load() (*Config, error) {
	port, err := parsePort(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}
	repository := getEnv("ACTIONS_REPOSITORY", "gat516/book")
	if !regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`).MatchString(repository) {
		return nil, fmt.Errorf("ACTIONS_REPOSITORY must be owner/repository")
	}
	workflow := getEnv("ACTIONS_WORKFLOW", "ci.yml")
	if !regexp.MustCompile(`^[A-Za-z0-9_.-]+$`).MatchString(workflow) {
		return nil, fmt.Errorf("ACTIONS_WORKFLOW must be a workflow filename or ID")
	}
	healthURL := getEnv("SERVICE_HEALTH_URL", "https://qireadr.com/api/healthz")
	u, err := url.Parse(healthURL)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("SERVICE_HEALTH_URL must be an HTTPS URL without credentials, query, or fragment")
	}

	return &Config{
		Application:        getEnv("APPLICATION_NAME", "qireadr"),
		ActionsRepository:  repository,
		ActionsWorkflow:    workflow,
		ActionsBranch:      getEnv("ACTIONS_BRANCH", "main"),
		ActionsToken:       os.Getenv("ACTIONS_READ_TOKEN"),
		ServiceHealthURL:   healthURL,
		DashboardStaticDir: os.Getenv("DASHBOARD_STATIC_DIR"),
		Port:               port,
		AllowedGitHubUsers: parseAllowedUsers(os.Getenv("ALLOWED_GITHUB_USERS")),
		Environment:        getEnv("ENVIRONMENT", "local"),
		CORSOrigins:        parseCORSOrigins(getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173")),
		PrometheusURL:      getEnv("PROMETHEUS_URL", "http://prometheus-operated.monitoring.svc.cluster.local:9090"),
	}, nil
}

// IsUserAllowed reports whether the given GitHub username is on the allowlist.
func (c *Config) IsUserAllowed(username string) bool {
	_, ok := c.AllowedGitHubUsers[username]
	return ok
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parsePort(s string) (int, error) {
	port, err := strconv.Atoi(s)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("%q is not a valid port number", s)
	}
	return port, nil
}

func parseCORSOrigins(raw string) map[string]bool {
	origins := make(map[string]bool)
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			origins[o] = true
		}
	}
	return origins
}

func parseAllowedUsers(raw string) map[string]struct{} {
	allowed := make(map[string]struct{})
	for _, u := range strings.Split(raw, ",") {
		u = strings.TrimSpace(u)
		if u != "" {
			allowed[u] = struct{}{}
		}
	}
	return allowed
}
