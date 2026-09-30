package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gat516/k8s-platform/config"
)

// These response types are the public allowlist. Never forward GitHub's raw
// response, commit messages, actor identities, logs, or credentials to clients.
type publicRun struct {
	ID               int64      `json:"id"`
	Number           int        `json:"number"`
	Commit           string     `json:"commit"`
	URL              string     `json:"url"`
	Status           string     `json:"status"`
	Conclusion       string     `json:"conclusion"`
	StartedAt        *time.Time `json:"started_at"`
	DurationSeconds  *int64     `json:"duration_seconds"`
	Checks           string     `json:"checks"`
	Build            string     `json:"build"`
	Deploy           string     `json:"deploy"`
	Verify           string     `json:"verify"`
	Rollback         string     `json:"rollback"`
	DetailsAvailable bool       `json:"details_available"`
}

type pipelineView struct {
	State          string      `json:"state"`
	Message        string      `json:"message,omitempty"`
	FetchedAt      *time.Time  `json:"fetched_at"`
	RefreshSeconds int         `json:"refresh_seconds"`
	Runs           []publicRun `json:"runs"`
}

type serviceView struct {
	State     string     `json:"state"`
	CheckedAt *time.Time `json:"checked_at"`
	Version   string     `json:"version,omitempty"`
	Message   string     `json:"message,omitempty"`
}

type dashboardView struct {
	Application string       `json:"application"`
	Pipeline    pipelineView `json:"pipeline"`
	Service     serviceView  `json:"service"`
}

type dashboardSource struct {
	cfg        *config.Config
	client     *http.Client
	githubBase string
	mu         sync.Mutex
	cached     pipelineView
	nextFetch  time.Time
	service    serviceView
	nextProbe  time.Time
}

func newDashboardSource(cfg *config.Config) *dashboardSource {
	return &dashboardSource{
		cfg:        cfg,
		client:     &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }},
		githubBase: "https://api.github.com",
		cached:     pipelineView{State: "unavailable", Runs: []publicRun{}},
	}
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if err := json.NewEncoder(w).Encode(s.dashboard.snapshot(r.Context())); err != nil {
		log.Printf("dashboard response: %v", err)
	}
}

func (d *dashboardSource) snapshot(ctx context.Context) dashboardView {
	// One refresh at a time, including across concurrent visitors. Public requests
	// cannot select the upstream repository, workflow, URL, or credentials.
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now().UTC()
	if now.After(d.nextFetch) {
		interval := 60 * time.Second
		if d.cfg.ActionsToken == "" {
			interval = 10 * time.Minute
		}
		if d.cfg.ActionsRepository == "" {
			d.cached = pipelineView{State: "not_configured", Message: "Connect a GitHub Actions workflow to see results.", Runs: []publicRun{}}
		} else {
			refreshCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			view, err := d.fetchPipeline(refreshCtx)
			cancel()
			if err != nil {
				d.cached.State = "unavailable"
				if d.cached.FetchedAt != nil {
					d.cached.State = "stale"
				}
				d.cached.Message = "GitHub Actions is unavailable. Check the repository, workflow, and server access."
			} else {
				d.cached = view
			}
		}
		d.cached.RefreshSeconds = int(interval.Seconds())
		d.nextFetch = time.Now().Add(interval)
	}
	if now.After(d.nextProbe) {
		d.service = d.probeService(ctx)
		d.nextProbe = time.Now().Add(30 * time.Second)
	}
	return dashboardView{Application: d.cfg.Application, Pipeline: d.cached, Service: d.service}
}

type githubRun struct {
	ID         int64      `json:"id"`
	Number     int        `json:"run_number"`
	Attempt    int        `json:"run_attempt"`
	SHA        string     `json:"head_sha"`
	Status     string     `json:"status"`
	Conclusion string     `json:"conclusion"`
	StartedAt  *time.Time `json:"run_started_at"`
}
type githubStep struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}
type githubJob struct {
	Name        string       `json:"name"`
	Status      string       `json:"status"`
	Conclusion  string       `json:"conclusion"`
	StartedAt   *time.Time   `json:"started_at"`
	CompletedAt *time.Time   `json:"completed_at"`
	Steps       []githubStep `json:"steps"`
}

func (d *dashboardSource) githubGet(ctx context.Context, path string, into any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.githubBase+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if d.cfg.ActionsToken != "" {
		req.Header.Set("Authorization", "Bearer "+d.cfg.ActionsToken)
	}
	res, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub status %d", res.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(res.Body, 2<<20)).Decode(into)
}

func (d *dashboardSource) fetchPipeline(ctx context.Context) (pipelineView, error) {
	var response struct {
		Runs []githubRun `json:"workflow_runs"`
	}
	path := "/repos/" + d.cfg.ActionsRepository + "/actions/workflows/" + url.PathEscape(d.cfg.ActionsWorkflow) + "/runs?per_page=5&event=push&branch=" + url.QueryEscape(d.cfg.ActionsBranch)
	if err := d.githubGet(ctx, path, &response); err != nil {
		return pipelineView{}, err
	}
	view := pipelineView{State: "available", Runs: []publicRun{}}
	for i, run := range response.Runs {
		if i >= 5 {
			break
		}
		out := publicRun{ID: run.ID, Number: run.Number, Commit: run.SHA, Status: run.Status, Conclusion: run.Conclusion, StartedAt: run.StartedAt,
			URL:    fmt.Sprintf("https://github.com/%s/actions/runs/%d", d.cfg.ActionsRepository, run.ID),
			Checks: "unknown", Build: "unknown", Deploy: "unknown", Verify: "unknown", Rollback: "unknown"}
		var jobs struct {
			Total int         `json:"total_count"`
			Jobs  []githubJob `json:"jobs"`
		}
		attempt := run.Attempt
		if attempt < 1 {
			attempt = 1
		}
		jobsPath := fmt.Sprintf("/repos/%s/actions/runs/%d/attempts/%d/jobs?per_page=100", d.cfg.ActionsRepository, run.ID, attempt)
		if err := d.githubGet(ctx, jobsPath, &jobs); err == nil && jobs.Total <= 100 && jobs.Total == len(jobs.Jobs) {
			out.DetailsAvailable = true
			summarizeJobs(&out, jobs.Jobs)
		}
		view.Runs = append(view.Runs, out)
	}
	now := time.Now().UTC()
	view.FetchedAt = &now
	return view, nil
}

func actionState(status, conclusion string) string {
	if status != "completed" {
		if status == "in_progress" {
			return "running"
		}
		return "pending"
	}
	switch conclusion {
	case "success":
		return "passed"
	case "failure", "timed_out", "action_required", "startup_failure":
		return "failed"
	case "cancelled":
		return "cancelled"
	case "skipped", "neutral":
		return "skipped"
	default:
		return "unknown"
	}
}

func summarizeJobs(out *publicRun, jobs []githubJob) {
	out.Checks, out.Build, out.Deploy, out.Verify, out.Rollback = "unknown", "not_configured", "not_configured", "not_configured", "not_configured"
	var start, end *time.Time
	complete := true
	var checks []string
	for _, job := range jobs {
		if job.Status != "completed" {
			complete = false
		}
		if job.StartedAt != nil && (start == nil || job.StartedAt.Before(*start)) {
			start = job.StartedAt
		}
		if job.CompletedAt != nil && (end == nil || job.CompletedAt.After(*end)) {
			end = job.CompletedAt
		}
		state := actionState(job.Status, job.Conclusion)
		name := strings.ToLower(job.Name)
		switch {
		case name == "build":
			out.Build = state
		case name == "deploy":
			out.Deploy = state
			for _, step := range job.Steps {
				switch step.Name {
				case "Verify release":
					out.Verify = actionState(step.Status, step.Conclusion)
				case "Restore and verify previous release":
					out.Rollback = actionState(step.Status, step.Conclusion)
				case "Rollback on failure":
					out.Rollback = "unverified"
					if step.Conclusion == "skipped" {
						out.Rollback = "skipped"
					}
				}
			}
		default:
			checks = append(checks, state)
		}
	}
	if len(checks) > 0 {
		out.Checks = "passed"
		// A failure always wins over pending jobs; skipped checks are never passes.
		for _, priority := range []string{"skipped", "unknown", "pending", "running", "cancelled", "failed"} {
			for _, state := range checks {
				if state == priority {
					out.Checks = state
				}
			}
		}
	}
	// Use job execution timestamps, not GitHub's mutable updated_at field.
	if complete && out.Status == "completed" && start != nil && end != nil && !end.Before(*start) {
		seconds := int64(end.Sub(*start).Seconds())
		out.DurationSeconds = &seconds
	}
}

func (d *dashboardSource) probeService(ctx context.Context) serviceView {
	if d.cfg.ServiceHealthURL == "" {
		return serviceView{State: "not_configured", Message: "No service health endpoint configured."}
	}
	now := time.Now().UTC()
	view := serviceView{State: "unknown", CheckedAt: &now}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.cfg.ServiceHealthURL, nil)
	if err != nil {
		return view
	}
	// Do not attach the GitHub credential to the monitored service request.
	res, err := d.client.Do(req)
	if err != nil {
		view.Message = "The latest health check could not reach the service."
		return view
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		view.State = "unhealthy"
		return view
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, 4096))
	if err != nil {
		return view
	}
	var body struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	if strings.EqualFold(strings.TrimSpace(string(raw)), "ok") {
		view.State = "healthy"
	} else if json.Unmarshal(raw, &body) == nil && body.Status == "ok" {
		view.State = "healthy"
	} else {
		view.Message = "The endpoint did not return a recognized health response."
		return view
	}
	if len(body.Version) <= 80 {
		view.Version = body.Version
	}
	return view
}
