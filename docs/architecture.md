# CI/CD Dashboard architecture

## Two independent observations

The React page calls `GET /api/v1/dashboard` on the Go backend. That backend reads:

1. GitHub Actions: runs and jobs for one workflow and branch.
2. An application health endpoint: whether it responds, and its reported version.

These observations stay separate. A passing workflow might contain only tests.
A failed workflow might have successfully rolled back and left the service healthy.

The dashboard has no deploy buttons, database, Kubernetes credentials, or model
provider credentials. GitHub Actions is the intended owner of ongoing releases;
the initial installation was bootstrapped by an operator through AWS SSM. The
dashboard itself never performs release operations.

## Hosting

`status.qireadr.com` is the live dashboard hostname. The Docker image contains
both the compiled React assets and Go API. Cloudflare DNS points to qireadr’s
existing AWS/k3s host; its Traefik ingress routes the page, assets, exact dashboard
API path, and `/health` to one container. Metrics remain internal. No separate
frontend host or cross-origin API URL is needed.

The runtime service account has no cluster permissions or mounted token. Keep
qireadr’s private `book` resources and shared Traefik configuration unchanged.
See [domain setup](domain.md) for the verified initial deployment, DNS/TLS setup,
and remaining CI deployment work. Automatic releases are still gated.

## Public data contract

`application`: configured display name.

`pipeline`: source state (`available`, `stale`, `unavailable`, `not_configured`), last
successful fetch time, refresh interval, and up to five runs. Each run contains
its ID/number, commit SHA, GitHub URL, status/conclusion, start time, execution
seconds, and the five stage results. No commit messages, actors, raw job names,
logs, secrets, or upstream error bodies are forwarded.

`service`: `healthy`, `unhealthy`, `unknown`, or `not_configured`; check timestamp;
optional reported version. A timeout or unreachable endpoint is unknown. A
non-200 response is unhealthy. An HTML page with status 200 is not a health result.

Workflow execution time spans the earliest job start to the last job completion
for that run attempt. It excludes time queued before jobs start. Missing job data
means unknown duration. It is not labeled deployment time or uptime.

## Recognizing the stages

- Jobs named `build` and `deploy` map to those stages.
- Other jobs contribute to the Checks stage (for the configured simple workflow).
- The deploy step `Verify release` reports the functional/version verification.
- The deploy step `Restore and verify previous release` reports verified recovery.
- The old `Rollback on failure` step is labeled **Not verified** even if it succeeded.
- Missing stages are **Not configured**. Skipped stages are **Not run**.

These names are an explicit contract for this small project, not automatic support
for arbitrary workflow layouts. A job-list response with more than 100 jobs is
left unknown instead of deriving a result from incomplete data.

## Caching and failures

The server caches all visitors' reads together. GitHub refreshes every minute with
a token, or ten minutes without one. The latter caps an ordinary five-run refresh
at roughly 36 API calls per hour. Other users of the same IP can still consume
GitHub's shared unauthenticated allowance. Service checks cache for 30 seconds.

Failed refreshes preserve prior pipeline results with a **stale** label and the
original fetched timestamp. An initial failure shows unavailable. The browser
also marks displayed data stale when it cannot reach the dashboard API.

No credentials are sent to the health endpoint. Redirects are not followed.
Upstream targets are fixed by server configuration, never by request parameters.

## Release workflow

The repository workflow builds the whole dashboard. Once replacement deployment
access is configured, its gated release job follows this sequence:

1. Run Go/Python tests, frontend build, and the existing security scans.
2. Build and publish an image identified by the source commit.
3. Capture the prior Kubernetes deployment revision and version reported over HTTP.
4. Apply the new release and wait for the rollout.
5. Verify HTTP health and the expected source version.
6. If deployment or verification fails, undo to the captured revision, wait, and
   verify the prior version over HTTP. Failed verification keeps recovery failed.

A failed release remains a failed workflow even when recovery succeeds. Concurrent
pushes are serialized so a second run cannot overwrite the first run's recovery.
An initial installation has no previous version and cannot promise rollback.

The workflow's `DEPLOY_HEALTH_URL` repository variable targets the service being
deployed. It is separate from the dashboard's `SERVICE_HEALTH_URL`, which defaults
to the qireadr application being observed. Deployment requires an explicit
`https://status.qireadr.com/health` value and `DASHBOARD_DEPLOY_ENABLED=true`.
There is no retired-host fallback. Keep the gate disabled until a scoped AWS
SSM deployment path or appropriate private runner access is implemented and tested;
the retained public-runner/kubeconfig job cannot reach the AWS API as configured.

## qireadr integration boundary

Keep qireadr's tests, images, and deployment workflow in `book`. This repository
owns the read-only view and the small reusable release script. qireadr is not
currently deployed by its checks workflow. Its reader API health response now includes its deployed source version, and
`book/main` push runs populate the dashboard. Its production migration process is
forward-only and must not be bypassed or replaced by this application-only script.

Before enabling qireadr CD: choose one service, expose its release version in a
health response, establish the narrowly scoped deployment credentials and route,
and run the good-release / broken-release / verified-recovery exercise in a
disposable environment. Only then report measured deployment/recovery times.
