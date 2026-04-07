# Architecture: Kubernetes Portfolio Platform

## System Overview

The k8s-platform is a k3s single-node cluster on a VPS that serves as the backend for all portfolio services. The cluster is the backend -- not the user-facing frontend. A static React dashboard hosted on Vercel polls the cluster's API endpoints and displays real-time service status.

## System Diagram

```
                     ┌────────────────────────┐
                     │      Cloudflare         │
                     │    DNS + SSL (Flexible) │
                     │                         │
                     │  Terminates TLS,        │
                     │  forwards HTTP to VPS   │
                     └───────────┬─────────────┘
                                 │ HTTP :80
                                 ▼
┌──────────────────────────────────────────────────────────────┐
│                     VPS (4GB RAM / 2 vCPU)                    │
│  ┌────────────────────────────────────────────────────────┐  │
│  │                     k3s Cluster                         │  │
│  │                                                         │  │
│  │  ┌───────────────────────────────────────────────────┐  │  │
│  │  │         Traefik Ingress (hostPort 80/443)          │  │  │
│  │  │                                                    │  │  │
│  │  │  api.<domain>      → k8s-platform-svc              │  │  │
│  │  │  <lc-domain>       → lcpatterns-svc                │  │  │
│  │  │  grafana.<domain>  → grafana-svc                   │  │  │
│  │  └───────────────────────────────────────────────────┘  │  │
│  │                                                         │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌───────────────┐  │  │
│  │  │ k8s-platform│  │ LCPatterns  │  │  Task Queue   │  │  │
│  │  │   API       │  │             │  │               │  │  │
│  │  │             │  │  frontend   │  │  gateway      │  │  │
│  │  │  /health    │  │  backend    │  │  workers      │  │  │
│  │  └─────────────┘  └─────────────┘  └───────────────┘  │  │
│  │                                                         │  │
│  │  ┌─────────────────────────────┐  ┌─────────────────┐  │  │
│  │  │     Monitoring Namespace     │  │  Data Layer     │  │  │
│  │  │  ┌───────────┐              │  │  ┌───────────┐  │  │  │
│  │  │  │Prometheus │              │  │  │PostgreSQL │  │  │  │
│  │  │  │+ Grafana  │              │  │  │ (Bitnami) │  │  │  │
│  │  │  │+ Alertmgr │              │  │  │ + PVC     │  │  │  │
│  │  │  └───────────┘              │  │  ├───────────┤  │  │  │
│  │  └─────────────────────────────┘  │  │  Redis    │  │  │  │
│  │                                    │  │ (Bitnami) │  │  │  │
│  │                                    │  │ + PVC     │  │  │  │
│  │                                    │  └───────────┘  │  │  │
│  └────────────────────────────────────┴─────────────────┘  │  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
         ▲                                        ▲
         │ SSH deploy / kubectl apply              │ fetch /health
         │                                         │
┌────────┴────────┐                   ┌────────────┴───────────┐
│  GitHub Actions  │                   │   Vercel Dashboard     │
│  CI/CD Pipeline  │                   │   (static React app)   │
│  Build → Push →  │                   │                        │
│  Deploy          │                   │  polls cluster APIs    │
└─────────────────┘                   └────────────────────────┘
```

## Domain Map (planned -- domains not yet purchased/configured)

| Domain | Hosting | Purpose |
|---|---|---|
| `charlesgatchalian.dev` | Separate repo | Personal portfolio -- independent of this project |
| TBD | Vercel | k8s-platform status dashboard (static React, polls cluster APIs) |
| TBD | VPS / Traefik | LeetCode Pattern Trainer -- fully in-cluster (frontend + backend) |
| TBD | VPS / Traefik | k8s-platform API endpoints, polled by the Vercel dashboard |
| TBD | VPS / Traefik | Grafana (internal observability) |

## Component Breakdown

### Frontend Layer

| Component | Purpose | Implementation |
|---|---|---|
| Vercel Dashboard | Public status page showing service health | Static React app in `dashboard/`, polls `/health` endpoints every 30s |
| LCPatterns Frontend | User-facing UI for LeetCode Pattern Trainer | Served from within the cluster by Traefik |

### Infrastructure Layer

| Component | Purpose | Implementation |
|---|---|---|
| k3s | Lightweight Kubernetes distribution | Single-node install, ~600MB overhead, includes containerd |
| Traefik | Ingress controller, reverse proxy | Bundled with k3s, hostPort 80/443 (servicelb disabled), routes via IngressRoute CRDs |
| Cloudflare | DNS management, SSL termination | Free tier, Flexible SSL mode -- terminates TLS at edge, forwards HTTP to VPS on port 80 |
| local-path-provisioner | Persistent storage | Bundled with k3s, provides PVCs backed by host filesystem |

### CI/CD Layer

| Component | Purpose | Implementation |
|---|---|---|
| GitHub Actions | Build and deploy automation | Triggered on push to main, builds Docker images |
| GHCR | Container image registry | Images tagged with Git SHA, pulled by k3s on deploy |
| Helm | Release management (MVP 4+) | Each service gets a Helm chart, values.yaml stores image tag and config |
| kubectl | Cluster access | Kubeconfig stored as GitHub Actions secret, deploy step runs kubectl/helm |

### Observability Layer

| Component | Purpose | Implementation |
|---|---|---|
| Prometheus | Metrics collection | kube-prometheus-stack Helm chart, scrapes all services via ServiceMonitor CRDs |
| Grafana | Dashboards and visualization | Bundled with kube-prometheus-stack, exposed via Traefik |
| Alertmanager | Alert routing | Bundled, sends alerts on pod failures, high error rates, resource pressure |
| kube-state-metrics | Kubernetes object metrics | Bundled, exposes pod/deployment/node status as Prometheus metrics |

### Data Layer

| Component | Purpose | Implementation |
|---|---|---|
| PostgreSQL | Relational database | Bitnami Helm chart, PVC for persistence, ~512MB memory limit |
| Redis | Cache, sessions, pub/sub | Bitnami Helm chart, PVC for persistence, ~128MB memory limit |
| CronJob (backup) | Database backup | Runs pg_dump nightly, keeps N rolling backups on disk |

## Data Flow: Dashboard to Cluster

```
Vercel dashboard (browser)
        │
        │  fetch("https://<api-domain>/health")
        ▼
   Cloudflare (DNS + TLS termination)
        │
        │  HTTP :80
        ▼
   Traefik (matches Host rule for API subdomain)
        │
        ▼
   k8s-platform pod (/health handler)
        │
        ▼
   JSON response: {"status": "ok", ...}
```

Browser-side polling every 30 seconds. No proxy needed -- the dashboard makes direct fetch calls to the cluster API. CORS is handled by the Go API via the `CORS_ORIGINS` environment variable (comma-separated list of allowed origins, defaults to `http://localhost:3000`).

## CI/CD Workflow

```
Developer pushes to main
        │
        ▼
GitHub Actions triggers
        │
        ├── Run tests
        ├── Build Docker image
        ├── Tag with Git SHA (e.g., ghcr.io/gat516/k8s-platform:abc1234)
        ├── Push to GHCR
        │
        ▼
Deploy step
        │
        ├── Use remote kubeconfig (stored as GitHub Actions secret)
        ├── kubectl set image / helm upgrade --set image.tag=abc1234
        │
        ▼
k3s performs rolling update
        │
        ├── New pod starts with new image
        ├── Readiness probe passes
        ├── Old pod terminated
        │
        ▼
Zero-downtime deployment complete
```

## Resource Budget (4GB RAM VPS)

| Component | Memory Allocation |
|---|---|
| k3s system (kubelet, containerd, Traefik) | ~600MB |
| PostgreSQL | ~512MB |
| Redis | ~128MB |
| Prometheus + Grafana + Alertmanager | ~512MB |
| k8s-platform API service | ~64MB |
| Task Queue (gateway + workers) | ~384MB |
| LCPatterns (frontend + backend) | ~384MB |
| Buffer | ~416MB |
| **Total** | **~3000MB** |

This leaves headroom for pod restarts and burst usage. If memory gets tight, first targets for optimization are Prometheus retention (reduce from default 15d to 3d) and Grafana (disable unused dashboards/plugins).

## Directory Structure

```
k8s-platform/
├── infrastructure/
│   ├── k3s-setup.sh                # k3s install + initial config
│   ├── traefik-config.yaml         # Traefik overrides (hostPort, etc.)
│   ├── cloudflare-dns.md           # DNS setup notes
│   └── backup-cronjob.yaml         # PostgreSQL backup CronJob
├── dashboard/                      # Vercel static dashboard (React)
│   ├── package.json
│   ├── vercel.json
│   └── src/
│       └── App.jsx
├── helm-charts/
│   ├── k8s-platform/
│   ├── pattern-trainer/
│   ├── task-queue/
│   └── monitoring/
│       ├── prom-values.yaml        # kube-prometheus-stack overrides
│       └── dashboards/             # Custom Grafana dashboard JSON
├── manifests/
│   ├── deployment.yaml
│   ├── service.yaml
│   └── ingressroute.yaml
├── .github/
│   └── workflows/
│       ├── ci.yaml                 # Build + deploy backend services
│       └── deploy-dashboard.yaml   # Deploy dashboard to Vercel
└── docs/
    └── architecture.md
```
