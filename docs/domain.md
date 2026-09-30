# Publish at status.qireadr.com

The dashboard and its Go API run together on qireadr’s existing AWS/k3s host.
Cloudflare provides DNS and HTTPS proxying.

```text
status.qireadr.com → Cloudflare → existing AWS Traefik → dashboard container
  /                       React page
  /assets/*               built JavaScript and CSS
  /api/v1/dashboard       sanitized GitHub Actions + qireadr health data
  /health                 dashboard release version for deployment verification
```

The Docker image includes the Vite build and serves it from the Go server. Leave
`VITE_API_URL` unset: the browser uses its own origin. No separate frontend service,
API hostname, Vercel project, database, or AWS stack is needed.

## Current deployment (September 29, 2026)

**Live:** https://status.qireadr.com, release `worktree-d0533b2b5190`. The
source-snapshot tag identifies the reviewed, still-uncommitted build inputs.
The initial release was applied through AWS SSM to instance
`i-0d907def09bd3092f`, with Elastic IP `3.221.223.180` confirmed by AWS.

- The dashboard pod is Ready with no restarts. Its observed memory use was 9Mi.
- Public HTML and both JS/CSS assets match the local production build exactly.
- `/health` reports the expected release version, and the release verifier passes.
- `/api/v1/dashboard` returns uncached JSON. qireadr reports healthy; GitHub is
  reachable and currently returns zero matching push runs for the configured source.
- `/metrics`, legacy APIs, trainer routes, and unknown paths return 404 publicly.
- The origin certificate validates for `status.qireadr.com` using TLS 1.3 and
  expires December 28, 2026. Public HTTPS also works through Cloudflare.
- DNS is already **Proxied**. Certificate issuance succeeded in that mode; keep
  the current record and existing HTTP ACME handling. No proxy toggle is needed.
- qireadr’s existing application deployments remained Ready after the rollout.

The full Docker build, local container smoke check (read-only filesystem, 64Mi
memory limit), and live HTTP checks passed. No browser was connected for visual
inspection. Automatic deployment and a controlled rollback drill remain pending.

The image was imported into the existing node using a bootstrap manifest with
`imagePullPolicy: IfNotPresent`. It has not been pushed to GHCR. Preserve
`/root/status-dashboard/releases/worktree-d0533b2b5190/` on the node: it contains
the image archive, source snapshot/hashes, and the exact applied `release.yaml`.
A replacement node needs that image imported again until registry publication is
configured. Local artifacts and verification results are also in
`/tmp/status-dashboard-release/`; `/tmp` is temporary storage.

The bootstrap used a temporary, private S3 object under `deployments/status/` in
the existing data bucket. That exact uploaded object version was deleted after
verification. No AWS resources, IAM policies, or registry credentials were added.

qireadr’s [hosting status](../../../book/docs/HOSTED_IMPLEMENTATION.md) records the
existing AWS origin as `3.221.223.180`, instance `i-0d907def09bd3092f` in `us-east-1`.
Confirm those resources before applying. Its [Cloudflare runbook](../../../book/deploy/cloudflare/README.md)
documents the existing `letsencrypt` Traefik resolver and Full (strict) TLS.
Management uses AWS SSM; the Kubernetes API is not open to public GitHub runners.

The dashboard’s resources are named `k8s-platform` in `default`; qireadr’s private
application stays in `book`. Check the node has room for the dashboard’s requests
(50m CPU, 32Mi memory) and 64Mi memory limit. The dashboard service account has no
cluster role and mounts no Kubernetes token. An optional Actions-read token belongs
only in Secret `cicd-dashboard-github`, key `token`, in `default`.

## 1. Build and prepare the release

From the repository root, after committing the reviewed source:

```bash
RELEASE_VERSION=$(git rev-parse --short HEAD)
docker build --build-arg VERSION="$RELEASE_VERSION" \
  -t "ghcr.io/gat516/k8s-platform:$RELEASE_VERSION" .
docker push "ghcr.io/gat516/k8s-platform:$RELEASE_VERSION"
sed "s|ghcr.io/gat516/k8s-platform:SHA|ghcr.io/gat516/k8s-platform:$RELEASE_VERSION|" \
  manifests/deployment.yaml > /tmp/status-dashboard-deployment.yaml
```

The registry package must be pullable by the node. If it is private, configure a
narrowly scoped image-pull Secret and reference it in the pod before applying.
Do not put registry credentials or Actions tokens in the image or frontend.

Use the existing SSM operator path to transfer the reviewed deployment file and
`manifests/{rbac,service,ingressroute}.yaml` to the node. On the node, apply just
those four files with `sudo k3s kubectl apply -f FILE`, then run:

```bash
sudo k3s kubectl -n default rollout status deployment/k8s-platform --timeout=120s
```

Do not run the legacy `infrastructure/k3s-setup.sh`, install another cluster, or
apply `infrastructure/traefik-config.yaml` over qireadr’s shared controller. The new
ingress uses its existing HTTPS entrypoint and certificate resolver. It exposes
only the four paths above; `/metrics` and old cluster/trainer routes are excluded.

## 2. DNS and TLS for a new installation

The live record is already configured and verified. For a new installation, use
this record in the `qireadr.com` zone:

| Setting | Value |
| --- | --- |
| Type | A |
| Name | `status` |
| IPv4 address | Existing AWS Elastic IP, recorded as `3.221.223.180`; confirm first |
| Proxy status | DNS only during origin certificate setup |
| TTL | Auto |

The existing Traefik HTTP challenge listener needs port 80 to obtain the new
hostname’s certificate. Preserve qireadr’s current HTTP challenge handling and
zone settings. In particular, do not add an edge redirect that intercepts ACME
challenges. No new Worker route is required for this subdomain.

After DNS resolves and Traefik obtains the certificate, verify origin HTTPS
without `-k` (substitute the confirmed origin address):

```bash
curl --fail --resolve status.qireadr.com:443:3.221.223.180 https://status.qireadr.com/
curl --fail --resolve status.qireadr.com:443:3.221.223.180 https://status.qireadr.com/health
```

Then enable Proxied for the `status` record, retaining Full (strict). Leave the
apex record, qireadr Worker, reader services, and database unchanged.

## 3. Verify the public release

- `https://status.qireadr.com` loads the dashboard and its JS/CSS assets.
- `/api/v1/dashboard` returns JSON with `pipeline` and `service`, without tokens,
  raw logs, actor identities, or commit messages.
- `/health` reports the deployed image’s release version.
- `/metrics`, `/api/v1/cluster`, `/trainer/`, and unknown paths return 404.
- GitHub failures and unknown/stale health remain explicit in the page.

```bash
python3 scripts/release.py verify --health-url https://status.qireadr.com/health \
  --version "$RELEASE_VERSION"
```

Also inspect the page: the version check alone does not prove the frontend renders
or upstream data is available. The verifier identifies itself as
`cicd-dashboard-release/1.0`; the default Python user agent was rejected by
Cloudflare with error 1010 during rollout. The named release client passed without
changing Cloudflare security settings.

## 4. Enable automated deployment only when connected

CI still tests, scans, and publishes the combined image. Its deployment job is
disabled unless repository variable `DASHBOARD_DEPLOY_ENABLED` is `true` and
`DEPLOY_HEALTH_URL` is exactly `https://status.qireadr.com/health`.

The retained job uses kubectl and `KUBECONFIG_B64`; it does not yet implement the
SSM transport used by this AWS host. Keep it disabled until a scoped deployment
path and runner connectivity are implemented and tested. Merely supplying a new
kubeconfig will not make the private API reachable from a public runner. Do not
reuse retired-cluster credentials or expose port 6443 to make the old job work.

The capture/verify/rollback script and simulated tests are available. Initial
installation has no previous revision; rehearse subsequent release/recovery in a
disposable environment before describing rollback as live-verified. qireadr’s
own release workflow remains owned by the `book` repository.

References:

- [Cloudflare error 1010](https://developers.cloudflare.com/support/troubleshooting/http-status-codes/cloudflare-1xxx-errors/error-1010/)
- [Cloudflare Full (strict)](https://developers.cloudflare.com/ssl/origin-configuration/ssl-modes/full-strict/)
- [Traefik IngressRoute TLS](https://doc.traefik.io/traefik/reference/routing-configuration/kubernetes/crd/http/ingressroute/)
