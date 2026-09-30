#!/usr/bin/env python3
"""One service, one previous revision, explicit HTTP/version verification.

Used by GitHub Actions. This changes an application release, never a database.
"""
import argparse
import json
import re
import subprocess
import time
import urllib.request


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def health(url):
    # Identify the release checker explicitly; Cloudflare can reject urllib's
    # generic Python signature before the request reaches the health endpoint.
    request = urllib.request.Request(url, headers={
        "User-Agent": "cicd-dashboard-release/1.0",
        "Accept": "application/json",
        "Cache-Control": "no-cache",
    })
    with urllib.request.build_opener(NoRedirect).open(request, timeout=5) as response:
        if response.status != 200:
            raise RuntimeError("health check did not return 200")
        body = json.loads(response.read(4096))
    version = body.get("version", "")
    if body.get("status") != "ok" or not re.fullmatch(r"[a-zA-Z0-9._-]{1,80}", version):
        raise RuntimeError("health endpoint must report status=ok and a release version")
    return version


def kubectl(namespace, *args):
    return subprocess.check_output(["kubectl", "-n", namespace, *args], text=True)


def capture(namespace, deployment, url):
    # --ignore-not-found distinguishes an initial install from a broken kubeconfig.
    raw = kubectl(namespace, "get", "deployment", deployment, "--ignore-not-found", "-o", "json")
    if not raw.strip():
        return {"previous_revision": "", "previous_version": ""}
    obj = json.loads(raw)
    revision = obj["metadata"].get("annotations", {}).get("deployment.kubernetes.io/revision", "")
    if not revision.isdigit():
        raise RuntimeError("existing deployment has no rollback revision")
    kubectl(namespace, "rollout", "status", f"deployment/{deployment}", "--timeout=120s")
    return {"previous_revision": revision, "previous_version": health(url)}


def verify(url, version, timeout=60):
    start = time.monotonic()
    while True:
        try:
            if health(url) == version:
                return round(time.monotonic() - start, 2)
        except (OSError, ValueError, RuntimeError):
            pass
        if time.monotonic() - start >= timeout:
            raise RuntimeError("expected release did not pass HTTP/version verification")
        time.sleep(2)


def rollback(namespace, deployment, revision, version, url):
    if not revision.isdigit() or not version:
        raise RuntimeError("a captured previous revision and version are required")
    start = time.monotonic()
    kubectl(namespace, "rollout", "undo", f"deployment/{deployment}", f"--to-revision={revision}")
    kubectl(namespace, "rollout", "status", f"deployment/{deployment}", "--timeout=120s")
    verify(url, version)
    return round(time.monotonic() - start, 2)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=["capture", "verify", "rollback"])
    parser.add_argument("--namespace", default="default")
    parser.add_argument("--deployment", default="k8s-platform")
    parser.add_argument("--health-url", required=True)
    parser.add_argument("--version")
    parser.add_argument("--revision")
    args = parser.parse_args()
    if not args.health_url.startswith("https://"):
        parser.error("health URL must use HTTPS")
    if args.command == "capture":
        for key, value in capture(args.namespace, args.deployment, args.health_url).items():
            print(f"{key}={value}")
    elif args.command == "verify":
        if not args.version:
            parser.error("--version is required")
        print(json.dumps({"verified": True, "verification_seconds": verify(args.health_url, args.version)}))
    else:
        if not args.version or not args.revision:
            parser.error("--version and --revision are required")
        print(json.dumps({"recovery_verified": True, "recovery_seconds": rollback(args.namespace, args.deployment, args.revision, args.version, args.health_url)}))


if __name__ == "__main__":
    main()
