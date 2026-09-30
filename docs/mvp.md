# CI/CD Dashboard: learning and verification checklist

The goal is to explain and demonstrate one complete release path. Use GitHub
Actions, Go, React, Docker, and qireadr’s existing AWS/k3s host. Dashboard publication and HTTPS were verified on September 29, 2026; replacement
CI deployment access is pending. The old deployment is retired.

## 1. Understand the dashboard (implemented)

- Follow `App.jsx` → `/api/v1/dashboard` → GitHub runs/jobs and the health endpoint.
- Explain why successful CI does not identify the running application version.
- Explain why stale, unknown, and failed are different states.
- Confirm that a browser response contains no upstream credential or raw logs.

## 2. Understand deployment (initial deployment verified; CI access pending)

- Follow `.github/workflows/ci.yaml`: tests → image → capture → deploy → verify.
- Explain how the image version matches the version returned by `/health`.
- Explain what deployment concurrency prevents.
- Run one successful release in a disposable test environment and save its timings.

## 3. Understand recovery (scripted and unit-tested; live drill pending)

- Deliberately deploy a version that never becomes healthy in that test environment.
- Verify that the workflow selects the captured previous revision.
- Confirm both rollout readiness and the old version's HTTP response.
- Show that failed recovery is reported as failure, not successful rollback.
- Record failure-to-verified-recovery time and the number of drills performed.

## 4. Connect one qireadr service (pending)

- Confirm the tracked repository, branch, and workflow settings.
- Supply Actions: read access only if needed for a private repository.
- Expose the chosen service's release version in its health response.
- Use a qireadr-owned deployment workflow with appropriate AWS access.
- Keep schema-changing releases outside this first automatic rollback path.

## Resume evidence to collect

- A successful release with a commit, image/version, and verification result.
- A failed release followed by verified recovery, with timestamps.
- The number of controlled failure drills and their observed recovery times.
- A before/after deployment-time comparison only if both were measured consistently.

The dashboard reports observations. It does not claim production uptime, zero
lost jobs, performance improvements, or tested worker recovery without evidence.
