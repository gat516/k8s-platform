#!/usr/bin/env python3
"""Read one CI cache benchmark attempt and save its timings, without rerunning it."""
import argparse
from datetime import datetime
import json
from pathlib import Path
import subprocess


def duration(start, end):
    return (datetime.fromisoformat(end) - datetime.fromisoformat(start)).total_seconds()


def summarize(jobs):
    results = {}
    for variant in ("baseline", "go-cache", "full"):
        group = [j for j in jobs if j["name"].startswith(f"trial ({variant}) / ")]
        names = {j["name"].rsplit(" / ", 1)[1] for j in group}
        if len(group) != 4 or names != {"test", "web", "scan", "build"}:
            raise ValueError(f"{variant}: expected all four jobs")
        if any(j["conclusion"] != "success" for j in group):
            raise ValueError(f"{variant}: all jobs must succeed")
        result = {
            "pipeline_seconds": duration(
                min(j["started_at"] for j in group),
                max(j["completed_at"] for j in group),
            ),
            "runner_seconds": sum(duration(j["started_at"], j["completed_at"]) for j in group),
        }
        for job in group:
            kind = job["name"].rsplit(" / ", 1)[1]
            result[f"{kind}_seconds"] = duration(job["started_at"], job["completed_at"])
            for step in job["steps"]:
                if step["name"] == "Install security scanners":
                    result["scanner_install_skipped"] = step["conclusion"] == "skipped"
                if step["name"] == "Build and push":
                    result["build_push_seconds"] = duration(step["started_at"], step["completed_at"])
        results[variant] = result
    return results


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", default="gat516/k8s-platform")
    parser.add_argument("--run-id", type=int, required=True)
    parser.add_argument("--attempt", type=int, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()

    def get(suffix):
        return json.loads(subprocess.check_output([
            "gh", "api", f"repos/{args.repo}/actions/runs/{args.run_id}/attempts/{args.attempt}{suffix}",
        ], text=True))

    run = get("")
    raw = get("/jobs?per_page=100")
    if run["status"] != "completed" or run["conclusion"] != "success":
        raise ValueError("Only successful, complete attempts are measurements")
    if raw["total_count"] != len(raw["jobs"]):
        raise ValueError("Job response is incomplete")
    jobs = [{key: job.get(key) for key in (
        "id", "name", "conclusion", "started_at", "completed_at", "labels", "steps",
    )} for job in raw["jobs"]]
    result = {
        "repository": args.repo, "run_id": args.run_id, "attempt": args.attempt,
        "commit": run["head_sha"],
        "url": f"https://github.com/{args.repo}/actions/runs/{args.run_id}/attempts/{args.attempt}",
        "run_started_at": run["run_started_at"], "conclusion": run["conclusion"],
        "jobs": jobs, "variants": summarize(jobs),
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result["variants"], indent=2))


if __name__ == "__main__":
    main()
