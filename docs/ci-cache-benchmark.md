# CI cache experiment

The September 29, 2026 dashboard runs reported a failed Go cache restore: the
workflow searched the repository root for `go.sum`, but this project keeps it in
`src/`. The scan job also rebuilt its scanner executables on every run. Docker
layers were not exported to a cache shared by later runners.

## Changes under test

- Point the Go test cache at `src/go.sum`.
- Give the scan job its own module/build cache key, incorporating
  `scripts/ci-tools.versions`, so the test job cannot save an incomplete scanner
  cache under the same key.
- Pin govulncheck to v1.8.0 and gosec to v2.29.0, the versions downloaded by the
  observed baseline. Cache those binaries by OS, architecture, Go module file,
  and tool-version file. Both scanners still run on every build; vulnerability
  results are not cached.
- Test a Buildx builder with GitHub Actions layer-cache import/export, using
  `mode=max` to preserve intermediate build stages.

## Method

The experiment is isolated on branch `codex/ci-cache-benchmark`, commit
`c41f226`. It does not change `main` and has no deployment job. The application
source and Dockerfile come from `d78a2aa` in every variant.

[Benchmark workflow and attempts](https://github.com/gat516/k8s-platform/actions/runs/36669744982)

Each attempt runs three variants on fresh GitHub-hosted `ubuntu-latest` runners:

1. **baseline:** no Go/scanner or persistent Docker cache, matching the effective
   behavior of the original broken Go cache configuration.
2. **go-cache:** working Go caches and cached, pinned scanner binaries.
3. **full:** the same Go/scanner caches plus persistent Docker layer caching.

Each variant retains the original dependency graph: tests, frontend build, and
security scans run in parallel; image build/push waits for all three. All
variants run the same checks and publish only `benchmark-*` image tags. The
benchmark uses pinned scanner versions in every variant to avoid measuring a
tool-version change. npm caching is the same in all variants.

Attempt 1 populates the new caches. Attempts 2 through 6 are the five measured
comparisons. Each attempt changes the image's embedded VERSION while keeping
application source constant. This forces the Go link/build layer to execute;
the result therefore describes repeated builds with a new version, not arbitrary
source or dependency changes. Cold-cache results are reported separately.

**Pipeline duration** is the time from the first job's start to the last job's
completion within a variant. It includes waits between jobs and excludes the
initial runner queue. **Runner seconds** sums the four job durations and is not
added to pipeline duration. These are CI measurements, not deployment time,
production recovery time, or billed cost.

Machine-readable evidence is in `benchmarks/ci-cache/`. Only complete attempts
with all twelve jobs passing qualify; skipped or failed checks cannot appear as
a speed improvement. Preserve all measured attempts, including slow ones.

Re-fetch an attempt with the read-only collector:

```bash
python3 scripts/collect_ci_benchmark.py \
  --run-id 36669744982 --attempt 2 \
  --output /tmp/ci-cache-attempt-2.json
```

The benchmark workflows live only on the experiment branch. The normal release
workflow keeps its deployment gate and serialized releases.
