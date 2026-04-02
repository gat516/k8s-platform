// Package config loads and validates runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration for the API server.
type Config struct {
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
}

// Load reads configuration from environment variables and returns a validated Config.
// Missing optional variables fall back to documented defaults.
// Returns an error if any required variable is absent or invalid.
func Load() (*Config, error) {
	port, err := parsePort(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	return &Config{
		Port:               port,
		AllowedGitHubUsers: parseAllowedUsers(os.Getenv("ALLOWED_GITHUB_USERS")),
		Environment:        getEnv("ENVIRONMENT", "local"),
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
