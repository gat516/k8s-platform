package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gat516/k8s-platform/config"
)

func TestDashboardFiltersPrivateDataAndCaches(t *testing.T) {
	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer test-secret" {
			t.Error("missing upstream credential")
		}
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/jobs") {
			fmt.Fprint(w, `{"total_count":2,"jobs":[{"name":"test","status":"completed","conclusion":"success","started_at":"2026-01-01T00:00:00Z","completed_at":"2026-01-01T00:01:00Z"},{"name":"deploy","status":"completed","conclusion":"failure","started_at":"2026-01-01T00:01:00Z","completed_at":"2026-01-01T00:02:30Z","steps":[{"name":"Verify release","status":"completed","conclusion":"failure"},{"name":"Restore and verify previous release","status":"completed","conclusion":"success"}]}]}`)
		} else {
			if r.URL.Query().Get("branch") != "main" || r.URL.Query().Get("event") != "push" {
				t.Error("missing workflow filters")
			}
			fmt.Fprint(w, `{"workflow_runs":[{"id":42,"run_number":7,"run_attempt":2,"head_sha":"abcdef1","status":"completed","conclusion":"failure","head_commit":{"message":"PRIVATE COMMIT"},"actor":{"login":"PRIVATE ACTOR"}}]}`)
		}
	}))
	defer upstream.Close()
	healthCalls := 0
	health := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		healthCalls++
		if r.Header.Get("Authorization") != "" {
			t.Error("GitHub token sent to health endpoint")
		}
		fmt.Fprint(w, `{"status":"ok","version":"old1234"}`)
	}))
	defer health.Close()
	d := newDashboardSource(&config.Config{Application: "qireadr", ActionsRepository: "owner/repo", ActionsWorkflow: "ci.yml", ActionsBranch: "main", ActionsToken: "test-secret", ServiceHealthURL: health.URL})
	d.githubBase = upstream.URL
	view := d.snapshot(context.Background())
	d.snapshot(context.Background())
	if calls != 2 || healthCalls != 1 {
		t.Fatalf("cache failed: GitHub=%d health=%d", calls, healthCalls)
	}
	run := view.Pipeline.Runs[0]
	if run.Checks != "passed" || run.Verify != "failed" || run.Rollback != "passed" || run.Conclusion != "failure" {
		t.Fatalf("incorrect failed-release/recovered-service distinction: %+v", run)
	}
	if run.DurationSeconds == nil || *run.DurationSeconds != 150 {
		t.Fatalf("wrong duration: %+v", run)
	}
	if view.Service.Version != "old1234" {
		t.Fatal("running version must come from service, not latest commit")
	}
	encoded, _ := json.Marshal(view)
	for _, private := range []string{"test-secret", "PRIVATE COMMIT", "PRIVATE ACTOR", health.URL} {
		if strings.Contains(string(encoded), private) {
			t.Errorf("public response leaked %s", private)
		}
	}
}

func TestDashboardRetainsButLabelsStaleResults(t *testing.T) {
	fail := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail {
			http.Error(w, "private upstream error", 403)
			return
		}
		fmt.Fprint(w, `{"workflow_runs":[]}`)
	}))
	defer upstream.Close()
	d := newDashboardSource(&config.Config{ActionsRepository: "owner/repo", ActionsWorkflow: "ci.yml"})
	d.githubBase = upstream.URL
	first := d.snapshot(context.Background())
	fail = true
	d.nextFetch = time.Time{}
	second := d.snapshot(context.Background())
	if second.Pipeline.State != "stale" || !second.Pipeline.FetchedAt.Equal(*first.Pipeline.FetchedAt) {
		t.Fatalf("stale results not marked: %+v", second)
	}
	if strings.Contains(second.Pipeline.Message, "private upstream") {
		t.Fatal("leaked upstream body")
	}
}

func TestCIOnlyAndUnverifiedRollback(t *testing.T) {
	view := publicRun{Status: "completed", Deploy: "not_configured", Verify: "not_configured", Rollback: "not_configured"}
	summarizeJobs(&view, []githubJob{{Name: "go (reader-api)", Status: "completed", Conclusion: "success"}})
	if view.Deploy != "not_configured" || view.Checks != "passed" {
		t.Fatalf("CI was confused with CD: %+v", view)
	}
	summarizeJobs(&view, []githubJob{{Name: "deploy", Status: "completed", Conclusion: "failure", Steps: []githubStep{{Name: "Rollback on failure", Status: "completed", Conclusion: "success"}}}})
	if view.Rollback != "unverified" {
		t.Fatal("rollback command success must not imply recovery was verified")
	}
}

func TestChecksFailureWinsAndSkippedIsNotPass(t *testing.T) {
	view := publicRun{}
	summarizeJobs(&view, []githubJob{{Name: "test", Status: "completed", Conclusion: "failure"}, {Name: "scan", Status: "queued"}})
	if view.Checks != "failed" {
		t.Fatal("queued work hid failed check")
	}
	summarizeJobs(&view, []githubJob{{Name: "test", Status: "completed", Conclusion: "skipped"}})
	if view.Checks != "skipped" {
		t.Fatal("skipped checks became passing checks")
	}
}

func TestDashboardDoesNotExposeUpstreamDetailsWhenJobsUnavailable(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/jobs") {
			http.Error(w, "secret", 500)
			return
		}
		fmt.Fprint(w, `{"workflow_runs":[{"id":42,"status":"completed","conclusion":"success"}]}`)
	}))
	defer upstream.Close()
	d := newDashboardSource(&config.Config{ActionsRepository: "owner/repo", ActionsWorkflow: "ci.yml"})
	d.githubBase = upstream.URL
	view := d.snapshot(context.Background())
	if view.Pipeline.Runs[0].DetailsAvailable || view.Pipeline.Runs[0].DurationSeconds != nil {
		t.Fatal("missing job details must stay unknown")
	}
}

func TestHealthFailureAndMissingVersion(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   string
	}{{200, `ok`, "healthy"}, {503, `unavailable`, "unhealthy"}, {200, `<html>login page</html>`, "unknown"}, {200, `{"status":"error"}`, "unknown"}} {
		t.Run(tc.want, func(t *testing.T) {
			u := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); fmt.Fprint(w, tc.body) }))
			defer u.Close()
			d := newDashboardSource(&config.Config{ServiceHealthURL: u.URL})
			got := d.probeService(context.Background())
			if got.State != tc.want || got.Version != "" {
				t.Fatalf("wrong observation: %+v", got)
			}
		})
	}
}
