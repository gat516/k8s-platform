# kube-platform

A personal Kubernetes platform built to demonstrate production-style DevOps practices — automated deployments, Helm packaging, and live observability through Prometheus and Grafana.

## Stack

- **Backend:** Go (API service)
- **Frontend:** Next.js on Vercel (portfolio + interactive platform)
- **Container Orchestration:** Kubernetes (k3s local, EKS production)
- **Infrastructure as Code:** Terraform
- **CI/CD:** GitHub Actions
- **Monitoring:** Prometheus + Grafana
- **Database:** Postgres (RDS)
- **Auth:** GitHub OAuth
- **Cloud:** AWS

## Architecture

**Go API** handles all backend logic — health checks, Prometheus metrics, and Kubernetes operations (deployments, pod status, log streaming). It acts as a secure proxy to the Kubernetes API so no cluster credentials are exposed to the frontend. Deployed as a `ClusterIP` service (internal-only, no public Ingress).

**Next.js frontend** on Vercel serves two modes:
- **Public portfolio** — server-rendered pages showing live cluster health, deployment history, and architecture. No auth required.
- **Interactive platform** — authenticated via GitHub OAuth with explicit allowlist authorization. Trigger deployments, view pod logs, scale services, monitor real-time status via SSE.

**Security model:** Next.js server-side code is the sole public gateway to the Go API (Backend for Frontend pattern). Every protected request carries a JWT validated by the Go API. The Go API runs under a least-privilege Kubernetes `ServiceAccount` bound to a namespace-scoped `Role`. GitHub Actions authenticates to AWS via OIDC — no static credentials stored.

**Infrastructure** runs on AWS (EKS in production, k3s locally). GitHub Actions handles CI/CD (test → build → push to GHCR → deploy via Helm). Prometheus and Grafana provide observability (internal-only, auth-gated).

## Project Structure

```
kube-platform/
├── .devcontainer/         # VS Code dev container (cross-platform)
├── .github/workflows/     # CI/CD pipelines
├── helm/                  # Helm charts
├── k8s/                   # Kubernetes manifests
│   ├── base/             # Base resources
│   └── overlays/         # Environment-specific overlays
├── monitoring/           # Prometheus, Grafana configs
├── scripts/              # Setup and utility scripts
├── src/                  # Go API service
├── frontend/             # Next.js dashboard (portfolio + platform)
└── terraform/            # Infrastructure as Code
```

## Getting Started

### Prerequisites

**Option A (Recommended): Dev Container**
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (Windows/macOS) or Docker Engine (Linux)
- [VS Code](https://code.visualstudio.com/) with [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)

**Option B: Native Installation**
- Docker
- kubectl, helm, terraform (install via scripts or package manager)
- k3s (Linux), k3d (macOS), or Docker Desktop K8s (Windows)

### Local Setup

**Option A: Dev Container (Works on all platforms)**
1. Open project in VS Code
2. When prompted, click "Reopen in Container" 
3. Setup local cluster:
   ```bash
   ./scripts/setup-local.sh
   ```

**Option B: Native Installation**

*Fedora:*
```bash
sudo dnf install -y docker
sudo systemctl enable --now docker
./scripts/install-tools.sh && source ~/.bashrc
```

*Ubuntu:*
```bash
sudo apt install -y docker.io
sudo systemctl enable --now docker
./scripts/install-tools.sh && source ~/.bashrc
```

*macOS:*
```bash
brew install --cask docker
brew install kubectl helm terraform k3d
```

*Windows:*
Use WSL2 + Docker Desktop, then follow Ubuntu instructions.

**Setup Cluster & Verify:**
```bash
./scripts/check-prereqs.sh
./scripts/setup-local.sh
kubectl get nodes
```

### Deploying to AWS

coming soon

## Status

Work in progress.

## Roadmap

- [ ] Local k3s cluster setup
- [ ] Go API service with `/health` and `/metrics` endpoints
- [ ] Unit tests (table-driven, `httptest`)
- [ ] Dockerfile + GHCR image publishing
- [ ] Helm chart with dedicated ServiceAccount + RBAC Role
- [ ] GitHub Actions pipeline with OIDC auth (test → build → push → deploy)
- [ ] Prometheus + Grafana stack (internal-only, auth-gated)
- [ ] Next.js frontend (portfolio + interactive platform)
- [ ] SSE real-time cluster status
- [ ] GitHub OAuth + allowlist authorization middleware
- [ ] JWT validation on all protected Go API endpoints
- [ ] Input validation middleware (resource name sanitization)
- [ ] WebSocket log streaming
- [ ] Live public dashboard
