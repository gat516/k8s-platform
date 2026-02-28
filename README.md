# kube-platform

A personal Kubernetes platform built to demonstrate production-style DevOps practices — automated deployments, Helm packaging, and live observability through Prometheus and Grafana.

## Stack

- Kubernetes (k3s local, EKS production)
- Terraform
- GitHub Actions
- Prometheus + Grafana
- Postgres (RDS)
- JavaScript (Netlify/Vercel)
- AWS

## Architecture

Frontend is hosted on Netlify/Vercel. Backend services and infrastructure run on AWS via EKS, with GitHub Actions handling CI/CD and Helm managing deployments. Prometheus and Grafana provide observability.

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
├── src/                  # Demo service source code
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
- [ ] Demo service with `/health` and `/metrics` endpoints
- [ ] Dockerfile + GHCR image publishing
- [ ] Helm chart
- [ ] GitHub Actions pipeline
- [ ] Prometheus + Grafana stack
- [ ] Live public dashboard
