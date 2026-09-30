# CI/CD Dashboard

This project has pivoted from a general Kubernetes portfolio dashboard to one
small release dashboard. Read `README.md` and `docs/architecture.md` before changes.

## Scope

- One configured application, one GitHub Actions workflow, one read-only page.
- GitHub Actions owns testing, deployment, and rollback. The dashboard reports results.
- qireadr is the first monitored application; its application code and release
  workflow belong in `/home/cj/projects/book` (not in this repo).
- No fabricated pipeline runs, uptime, latency, success states, or release versions.
- Keep unknown, stale, failed, and verified recovery distinct.
- Keep the project understandable: avoid adding another orchestrator, database,
  event system, cluster management UI, or new cloud service without a concrete need.

## Boundaries

- Public response types in `src/server/dashboard.go` are the output allowlist.
- GitHub credentials stay server-side; never put them in a `VITE_` variable.
- No user input may choose upstream URLs or repositories.
- Legacy Kubernetes handler files are retained but are not registered as public routes.
- Existing Go module and Kubernetes/image identifiers remain compatible with the
  prior deployment. The product/frontend package is CI/CD Dashboard / cicd-dashboard.
- Preserve pre-existing uncommitted work. Do not touch private key files or kubeconfigs.

## Checks

From `src`: `go test ./...` and `go vet ./...`.
From the root: `python3 -m unittest discover -s scripts -p 'test_*.py'`.
From `frontend/dashboard`: `npm run build`.

Release tests use simulated Kubernetes responses. Do not call them a successful
live rollback drill or publish recovery metrics without a controlled real run.
