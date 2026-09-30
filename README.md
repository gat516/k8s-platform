# CI/CD Dashboard

A small Go/React dashboard for understanding how one application moves from a
commit to a running release. The first tracked application is **qireadr**.

The learning goal is simple: **test → build → deploy → verify → recover**.
GitHub Actions runs the pipeline; this dashboard explains its results.

## What is implemented

- A Go API fetches the latest five push runs of one configured GitHub Actions workflow.
- A React page shows checks, build, deploy, verification, and rollback outcomes,
  plus commit IDs, workflow execution times, and links to the actual runs.
- An independent HTTP probe shows whether the application responds and its version
  when the health endpoint supplies one. CI success is never treated as evidence
  that the application was deployed.
- Results are cached, unavailable sources are labeled, and last successful fetch
  timestamps remain visible when results become stale.
- The dashboard's own Actions workflow tests, scans, and builds a combined UI/API image.
  Deployment is gated until the replacement AWS access is configured.
  It records the prior deployment revision, verifies the new version over HTTP,
  and restores **and verifies** the prior revision if deployment or verification fails.
- The public API returns a fixed set of fields. It does not return commit messages,
  actor identities, logs, tokens, private health URLs, or cluster inventory.

## Release status

The default source is `gat516/book`, workflow `ci.yml`, branch `main`. The reviewed
application changes are now on `main`, and real push runs are visible. That workflow
runs checks and builds; it does not deploy qireadr. Deploy/verify/rollback therefore
show **Not configured**, rather than invented successful releases.

qireadr's health endpoint reports the running reader API version (`61cc88c` at
this release). Other services may use different versions. Do not infer their
running versions from a Git commit or a successful workflow run.

The dashboard is live at **https://status.qireadr.com** on qireadr's existing
AWS/k3s node, behind Cloudflare. Initial deployment and HTTPS were verified on
September 29, 2026. The current release is committed source `ce467b2`, with a
plain service summary, workflow stages and run history. GitHub is reachable and
reports the `book/main` push runs.
Automatic deployment remains gated until CI access is connected. A controlled
recovery drill is still needed before claiming measured recovery performance.

## Local development

Requirements: Go matching `src/go.mod`, Node 22+, npm, and Python 3 for release tests.
No Kubernetes cluster, database, or Prometheus installation is needed to run the UI.

Terminal 1, from the repository root:

```bash
cp .env.example .env
set -a
. ./.env
set +a
cd src
go run .
```

Terminal 2:

```bash
cd frontend/dashboard
npm ci
npm run dev
```

Open the Vite URL. Development `/api/` requests proxy to the local Go API on port
8080. Production serves the built page and API from one container and origin;
leave `VITE_API_URL` unset. Docker sets `DASHBOARD_STATIC_DIR=/dashboard`. To test
the same setup locally, build the frontend and set `DASHBOARD_STATIC_DIR` to the
absolute path of `frontend/dashboard/dist` when starting Go.

### Data source configuration

| Variable | Purpose | Default |
| --- | --- | --- |
| `APPLICATION_NAME` | Display name | `qireadr` |
| `ACTIONS_REPOSITORY` | Repository to observe | `gat516/book` |
| `ACTIONS_WORKFLOW` | Workflow filename or numeric ID | `ci.yml` |
| `ACTIONS_BRANCH` | Branch whose push runs are shown | `main` |
| `ACTIONS_READ_TOKEN` | Optional server-side token with Actions: read for that repository | unset |
| `SERVICE_HEALTH_URL` | HTTPS health endpoint | `https://qireadr.com/api/healthz` |

Public repositories can work without a token; GitHub results then refresh every
ten minutes to limit API use. With a token, they refresh every minute. Service
checks refresh every 30 seconds while the dashboard is being read. This is an
on-demand dashboard, not a continuous uptime recorder.

A private repository needs a fine-grained token limited to Actions: read. Supplying
that token intentionally publishes the selected workflow's sanitized metadata to
this public dashboard. Do not put it in a `VITE_` variable. In Kubernetes, the
optional Secret is `cicd-dashboard-github`, key `token`.

The monitored health endpoint must return HTTP 200 and either plain `ok` or JSON
`{"status":"ok","version":"abc1234"}`. `version` is optional for display, but
required by the deployment verification script.

## The deliberately small project scope

1. One application and one Actions workflow.
2. Versioned images, serialized deployments, health/version checks, and recovery
   to the captured prior revision.
3. One read-only page with real results and explicit unknown/stale states.

Start with application-only releases whose database schema and external resources
remain compatible. Kubernetes rollback restores the deployment's pod template;
it does not restore databases, Secrets, ConfigMaps, or other infrastructure.

See [the architecture](docs/architecture.md) for the data flow and
[the learning checklist](docs/mvp.md) for the implementation/verification boundary.
See [the domain setup](docs/domain.md) to publish the dashboard and API at `status.qireadr.com`.

## Verification

```bash
cd src && go test ./... && go vet ./...
```

From the repository root:

```bash
python3 -m unittest discover -s scripts -p 'test_*.py'
cd frontend/dashboard && npm run build
```

The tests cover redaction, cache behavior, stale results, failure states, separation
of CI from deployment, exact-version verification, and rollback failure reporting.
The release tests simulate Kubernetes responses; they do not prove recovery on a
live cluster. Record real deployment and recovery timings after a controlled drill.

See the [CI cache experiment](docs/ci-cache-benchmark.md) for the repeatable
comparison of Go/scanner caches and Docker layer reuse, including raw run timings.

## Naming and deployment compatibility

The product and frontend package are **CI/CD Dashboard** / `cicd-dashboard`.
Existing Go module, GitHub remote, image repository, Kubernetes resources, and local
checkout path retain `k8s-platform` for compatibility. These identifiers do not
imply that the retired deployment is still running. Historical cluster
handlers remain in source but are no longer registered on the public API.
