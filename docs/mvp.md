# Portfolio Project Plan

---

## The Big Picture

Three projects, three repos, one cluster. Each project stands alone as a resume line with its own signal, but all three deploy to the same Kubernetes platform and share infrastructure -- exactly how real microservices work.

| Project | Signal | Repo |
|---|---|---|
| Kubernetes Portfolio Platform | Infrastructure, DevOps, observability | `k8s-platform` |
| LeetCode Pattern Trainer | Full-stack product, real-time systems, adaptive learning | `pattern-trainer` |
| Distributed Task Queue | Distributed systems, message-driven architecture, failure handling | `task-queue` |

## Build Order

1. **K8s Platform** -- get the cluster and pipeline running, build the status dashboard
2. **Pattern Trainer** -- deploy backend to the cluster, frontend TBD
3. **Task Queue** -- round out the resume with the systems project

---

# Project 1: Kubernetes Portfolio Platform

**Repo:** `k8s-platform`

## Overview

A production-style Kubernetes platform that acts as a centralized backend for all portfolio services. The platform has two parts:

1. **A static status dashboard** hosted on Vercel -- a React app that polls the `/health` endpoints of each service and displays real-time uptime and status.
2. **A k3s cluster on a VPS** -- runs backend services, Prometheus + Grafana for observability, and PostgreSQL + Redis as shared stateful services.

The dashboard is intentionally decoupled from the cluster. It is a pure frontend that anyone can open in a browser and see what's running.

## Objective

Demonstrate the ability to design, deploy, and operate containerized applications in a cloud-native environment using modern DevOps practices -- while building something that is actually visible and shareable.

## Architecture Highlights

- k3s single-node cluster on a cheap VPS (no cloud provider needed)
- CI/CD pipeline (GitHub Actions) builds, versions, and deploys images automatically
- Helm charts manage application configuration and release versioning
- Prometheus and Grafana provide real-time metrics inside the cluster
- Traefik Ingress exposes backend service APIs to the public internet
- Vercel hosts the status dashboard -- no server needed, zero cost

## Domain Map (planned -- domains not yet purchased/configured)

| Domain | Hosting | Purpose |
|---|---|---|
| `charlesgatchalian.dev` | Separate repo | Personal portfolio -- independent of this project |
| TBD | Vercel | This project's status dashboard |
| TBD | VPS / Traefik | LeetCode Pattern Trainer (fully in-cluster) |
| TBD | VPS / Traefik | Backend API endpoints (polled by dashboard) |
| TBD | VPS / Traefik | Grafana (internal observability) |

## Demonstration Scenario

Open the status dashboard on a mobile device and see:

- Real-time health status for each service (green/red, last seen timestamp)
- Pod-level metrics from Prometheus (CPU, memory, request rate)
- Uptime history and recent incidents

## Professional Signal

Platform engineering, Kubernetes operations, CI/CD automation, observability design, decoupled frontend/backend architecture, production-readiness mindset.

## Tech Stack

| Category | Technology |
|---|---|
| Orchestration | k3s |
| CI/CD | GitHub Actions |
| Config Management | Helm |
| Monitoring | Prometheus + Grafana |
| Ingress | Traefik (bundled with k3s) |
| DNS/TLS | Cloudflare (free tier) |
| Container Registry | GitHub Container Registry (GHCR) |
| Dashboard Hosting | Vercel (free tier) |
| Stateful Services | PostgreSQL + Redis (Bitnami Helm charts) |

## MVP Phases

### MVP 1: Single-Node Cluster with One Service -- COMPLETE

**Goal:** Get a k3s cluster running on a cheap VPS with one containerized app deployed.

- Provision a $5-10/mo VPS (Hetzner or DigitalOcean, 4GB RAM)
- Install k3s (single-node cluster, ~512MB overhead)
- Containerize a simple health-check API (Go, returns JSON status)
- Write a Dockerfile, push to GHCR
- Create a Kubernetes Deployment and Service manifest, apply manually with kubectl
- **Deliverable:** `curl your-server:port/health` returns `{"status": "ok"}` from a pod

### MVP 2: CI/CD Pipeline -- COMPLETE

**Goal:** Push to main triggers automatic build and deploy.

- GitHub Actions workflow: build image, tag with SHA, push to GHCR
- Add a deploy step that uses a remote kubeconfig to run kubectl apply with the updated image tag
- Set up a kubeconfig secret in GitHub Actions for remote cluster access
- **Deliverable:** Git push triggers new image build, push to GHCR, and rolling update on cluster -- zero manual steps

### MVP 3: Vercel Status Dashboard

**Goal:** A public-facing status page showing real-time health of all services, hosted free on Vercel.

- Expose the k8s-platform API publicly via Traefik Ingress (requires domain + Cloudflare)
- Add CORS support to the Go API (`CORS_ORIGINS` env var, comma-separated allowed origins, defaults to `http://localhost:3000`)
- Build a minimal React dashboard in `dashboard/` that polls each service's `/health` URL
- Configure Vercel deployment (vercel.json, environment variables for API URLs)
- Add a `deploy-dashboard.yaml` GitHub Actions workflow that deploys to Vercel on push to main
- **Deliverable:** Status dashboard shows live health for each service

### MVP 4: Helm + Observability

**Goal:** Replace raw manifests with Helm charts, add Prometheus and Grafana.

- Create a Helm chart for each service (values.yaml for image tag, replicas, env vars)
- Deploy kube-prometheus-stack via Helm (Prometheus + Grafana in one chart)
- Tune resource requests/limits so monitoring fits on a single node alongside services
- Expose Grafana behind Traefik Ingress
- Import a basic dashboard showing CPU, memory, pod status, and request rate
- **Deliverable:** Live Grafana URL showing cluster health metrics

### MVP 5: In-Cluster Stateful Services

**Goal:** PostgreSQL and Redis running inside the cluster with persistent storage.

- Deploy PostgreSQL via Bitnami Helm chart with a PersistentVolumeClaim
- Deploy Redis via Bitnami Helm chart
- Configure resource limits to fit within VPS memory budget (Postgres ~512MB, Redis ~128MB)
- Create a backup CronJob that dumps PostgreSQL to a file and keeps N rolling backups on disk
- **Deliverable:** `psql` into in-cluster Postgres, `redis-cli ping` returns PONG, data survives pod restarts

### MVP 6: Multi-Service Platform

**Goal:** Deploy the pattern trainer and task queue as services in the same cluster, with the dashboard showing all of them.

- Each project gets its own Helm chart and GitHub Actions workflow
- Shared Traefik Ingress routes traffic by domain/subdomain
- All services connect to the shared PostgreSQL (separate databases) and Redis instances
- All services report metrics to the same Prometheus/Grafana stack
- Vercel dashboard updated to show status for all services
- **Deliverable:** Unified platform; LeetCode Pattern Trainer is fully live, dashboard reflects all services
