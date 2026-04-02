# Architecture: Kubernetes Portfolio Platform

## System Architecture

```
                         ┌─────────────┐
                         │  Cloudflare  │
                         │  DNS + SSL   │
                         └──────┬───────┘
                                │ HTTPS
                                ▼
┌───────────────────────────────────────────────────────────┐
│                    VPS (Hetzner / DO)                      │
│                     4GB RAM / 2 vCPU                       │
│  ┌─────────────────────────────────────────────────────┐  │
│  │                   k3s Cluster                        │  │
│  │                                                      │  │
│  │  ┌──────────────────────────────────────────────┐   │  │
│  │  │            Traefik Ingress Controller          │   │  │
│  │  │  patterns.domain.com → pattern-trainer-svc     │   │  │
│  │  │  tasks.domain.com    → task-queue-svc          │   │  │
│  │  │  grafana.domain.com  → grafana-svc             │   │  │
│  │  └──────────────────────────────────────────────┘   │  │
│  │                                                      │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │  │
│  │  │ Pattern  │  │  Task    │  │    Monitoring     │  │  │
│  │  │ Trainer  │  │  Queue   │  │    Namespace      │  │  │
│  │  │ Service  │  │ Service  │  │  ┌────────────┐   │  │  │
│  │  │          │  │          │  │  │ Prometheus │   │  │  │
│  │  │ frontend │  │ gateway  │  │  │ + Grafana  │   │  │  │
│  │  │ backend  │  │ workers  │  │  │ + Alertmgr │   │  │  │
│  │  └──────────┘  └──────────┘  │  └────────────┘   │  │  │
│  │                               └──────────────────┘  │  │
│  │  ┌──────────────────────────────────────────────┐   │  │
│  │  │            Shared Data Layer                   │   │  │
│  │  │  ┌────────────┐       ┌─────────────┐        │   │  │
│  │  │  │ PostgreSQL │       │    Redis     │        │   │  │
│  │  │  │  (Bitnami) │       │  (Bitnami)  │        │   │  │
│  │  │  │  + PVC     │       │  + PVC      │        │   │  │
│  │  │  └────────────┘       └─────────────┘        │   │  │
│  │  └──────────────────────────────────────────────┘   │  │
│  └─────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────┘
         ▲
         │ SSH deploy / kubectl apply
         │
┌────────┴────────┐
│  GitHub Actions  │
│  CI/CD Pipeline  │
│  Build → Push →  │
│  Deploy          │
└─────────────────┘
```

## Component Breakdown

### Infrastructure Layer

|Component|Purpose|Implementation|
|---|---|---|
|k3s|Lightweight Kubernetes distribution|Single-node install, ~512MB overhead, includes containerd|
|Traefik|Ingress controller, reverse proxy|Bundled with k3s, routes subdomains to services via IngressRoute CRDs|
|Cloudflare|DNS management, SSL termination|Free tier, proxies traffic to VPS IP, handles TLS at the edge|
|local-path-provisioner|Persistent storage|Bundled with k3s, provides PVCs backed by host filesystem|

### CI/CD Layer

|Component|Purpose|Implementation|
|---|---|---|
|GitHub Actions|Build and deploy automation|Triggered on push to main, builds Docker images|
|GHCR|Container image registry|Images tagged with Git SHA, pulled by k3s on deploy|
|Helm|Release management|Each service has a Helm chart, values.yaml stores image tag and config|
|kubectl|Cluster access|Kubeconfig stored as GitHub Actions secret, deploy step runs kubectl/helm|

### Observability Layer

|Component|Purpose|Implementation|
|---|---|---|
|Prometheus|Metrics collection|kube-prometheus-stack Helm chart, scrapes all services via ServiceMonitor CRDs|
|Grafana|Dashboards and visualization|Bundled with kube-prometheus-stack, pre-loaded dashboards + custom ones|
|Alertmanager|Alert routing|Bundled, sends alerts on pod failures, high error rates, resource pressure|
|kube-state-metrics|Kubernetes object metrics|Bundled, exposes pod/deployment/node status as Prometheus metrics|

### Data Layer

|Component|Purpose|Implementation|
|---|---|---|
|PostgreSQL|Relational database|Bitnami Helm chart, PVC for persistence, ~512MB memory limit|
|Redis|Cache, sessions, pub/sub|Bitnami Helm chart, PVC for persistence, ~128MB memory limit|
|CronJob (backup)|Database backup|Runs pg_dump nightly, keeps N rolling backups on disk|

## Networking Flow

```
User request → Cloudflare (DNS + SSL) → VPS:443 → Traefik → Service ClusterIP → Pod
```

1. User hits `patterns.yourdomain.com`
2. Cloudflare resolves DNS to VPS IP, terminates TLS
3. Request forwarded to VPS port 443 (HTTP)
4. Traefik matches IngressRoute rule for subdomain
5. Routes to the correct Kubernetes Service (ClusterIP)
6. Service load-balances across healthy pods

## CI/CD Workflow

```
Developer pushes to main
        │
        ▼
GitHub Actions triggers
        │
        ├── Build Docker image
        ├── Tag with Git SHA (e.g., ghcr.io/gat516/pattern-trainer:abc1234)
        ├── Push to GHCR
        │
        ▼
Deploy step
        │
        ├── SSH into VPS (or use remote kubeconfig)
        ├── helm upgrade --set image.tag=abc1234
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

|Component|Memory Allocation|
|---|---|
|k3s system (kubelet, containerd, Traefik)|~600MB|
|PostgreSQL|~512MB|
|Redis|~128MB|
|Prometheus + Grafana + Alertmanager|~512MB|
|Pattern Trainer (frontend + backend)|~384MB|
|Task Queue (gateway + workers)|~384MB|
|RabbitMQ|~256MB|
|Buffer|~224MB|
|**Total**|**~3000MB**|

This leaves headroom for pod restarts and burst usage. If memory gets tight, first targets for optimization are Prometheus retention (reduce from default 15d to 3d) and Grafana (disable unused dashboards/plugins).

## Directory Structure

```
k8s-platform/
├── README.md
├── infrastructure/
│   ├── k3s-setup.sh              # k3s install + initial config
│   ├── cloudflare-dns.md         # DNS setup notes
│   └── backup-cronjob.yaml       # PostgreSQL backup CronJob
├── helm-charts/
│   ├── pattern-trainer/
│   │   ├── Chart.yaml
│   │   ├── values.yaml
│   │   └── templates/
│   │       ├── deployment.yaml
│   │       ├── service.yaml
│   │       └── ingress.yaml
│   ├── task-queue/
│   │   ├── Chart.yaml
│   │   ├── values.yaml
│   │   └── templates/
│   └── monitoring/
│       ├── prom-values.yaml      # kube-prometheus-stack overrides
│       └── dashboards/           # Custom Grafana dashboard JSON
├── .github/
│   └── workflows/
│       ├── deploy-pattern-trainer.yaml
│       └── deploy-task-queue.yaml
└── docs/
    └── architecture.md
```
