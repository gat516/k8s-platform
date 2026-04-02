# Portfolio Project Plan

---

## The Big Picture

Three projects, three repos, one cluster. Each project stands alone as a resume line with its own signal, but all three deploy to the same Kubernetes platform and share infrastructure — exactly how real microservices work.

| Project                       | Signal                                                             | Repo              |
| ----------------------------- | ------------------------------------------------------------------ | ----------------- |
| Kubernetes Portfolio Platform | Infrastructure, DevOps, observability                              | `k8s-platform`    |
| LeetCode Pattern Trainer      | Full-stack product, real-time systems, adaptive learning           | `pattern-trainer` |
| Distributed Task Queue        | Distributed systems, message-driven architecture, failure handling | `task-queue`      |

## Build Order

1. **Pattern Trainer MVP 1–3** — get the product working and shareable
2. **K8s Platform** — deploy it live on a VPS
3. **Task Queue** — round out the resume with the systems project

# Project 1: Kubernetes Portfolio Platform

**Repo:** `k8s-platform`

## Overview

A production-style Kubernetes platform that acts as a centralized control plane for all other portfolio services. Instead of a traditional portfolio website, this project is built as a live cloud environment with automated deployments, health monitoring, and observability dashboards.

## Objective

Demonstrate the ability to design, deploy, and operate containerized applications in a cloud-native environment using modern DevOps practices.

## Architecture Highlights

- k3s single-node cluster on a cheap VPS (no cloud provider needed)
- CI/CD pipeline (GitHub Actions) builds, versions, and deploys images automatically
- Helm charts manage application configuration and release versioning
- Prometheus and Grafana provide real-time metrics and health dashboards
- Traefik Ingress with Cloudflare for HTTPS routing to all services
- Liveness/readiness probes and self-healing behavior through Kubernetes

## Core Capabilities

- Automated build and deployment workflow
- Centralized metrics for CPU, memory, request rate, and error rate
- Service health visibility and fault recovery
- Infrastructure configuration managed declaratively
- In-cluster stateful services (PostgreSQL, Redis) with persistent storage

## Demonstration Scenario

Live dashboard accessible from mobile device showing:

- Real-time system health and service metrics
- Pod restarts and self-healing behavior
- Traffic spikes reflected in monitoring dashboards

## Professional Signal

Platform engineering, Kubernetes operations, CI/CD automation, observability design, production-readiness mindset.

## Tech Stack

|Category|Technology|
|---|---|
|Orchestration|k3s|
|CI/CD|GitHub Actions|
|Config Management|Helm|
|Monitoring|Prometheus + Grafana|
|Ingress|Traefik (bundled with k3s)|
|DNS/TLS|Cloudflare (free tier)|
|Container Registry|GitHub Container Registry (GHCR)|
|Stateful Services|PostgreSQL + Redis (Bitnami Helm charts)|

## MVP Phases

### MVP 1: Single-Node Cluster with One Service (1–2 days)

**Goal:** Get a k3s cluster running on a cheap VPS with one containerized app deployed.

- Provision a $5–10/mo VPS (Hetzner or DigitalOcean, 4GB RAM)
- Install k3s (single-node cluster, ~512MB overhead)
- Containerize a simple health-check API (FastAPI, returns JSON status)
- Write a Dockerfile, push to GHCR
- Create a Kubernetes Deployment and Service manifest, apply manually with kubectl
- **Deliverable:** `curl your-server:port/health` returns `{"status": "ok"}` from a pod

### MVP 2: CI/CD Pipeline (1 day)

**Goal:** Push to main triggers automatic build and deploy.

- GitHub Actions workflow: build image, tag with SHA, push to GHCR
- Add a deploy step that SSHs into the VPS and runs kubectl apply with the updated image tag
- Set up a kubeconfig secret in GitHub Actions for remote cluster access
- Trigger a deploy by merging a PR
- **Deliverable:** Git push → new image → rolling update on cluster, zero manual steps

### MVP 3: DNS + TLS + Ingress (1 day)

**Goal:** Real domain with HTTPS routing to services.

- Register a cheap domain (~$10/year) and point DNS to Cloudflare
- Cloudflare proxies to your VPS IP (free SSL termination at the edge)
- Deploy Traefik (bundled with k3s) as the Ingress controller
- Create Ingress rules routing subdomains to services (patterns.yourdomain.com, tasks.yourdomain.com)
- **Deliverable:** Hit patterns.yourdomain.com over HTTPS and reach your service

### MVP 4: Helm + Observability (2–3 days)

**Goal:** Replace raw manifests with Helm charts, add Prometheus and Grafana.

- Create a Helm chart for your service (values.yaml for image tag, replicas, env vars)
- Deploy kube-prometheus-stack via Helm (Prometheus + Grafana in one chart)
- Tune resource requests/limits so monitoring fits on a single node alongside your services
- Expose Grafana behind your Ingress (grafana.yourdomain.com)
- Import a basic dashboard showing CPU, memory, pod status, and request rate
- **Deliverable:** Live Grafana URL showing cluster health metrics

### MVP 5: In-Cluster Stateful Services (1–2 days)

**Goal:** PostgreSQL and Redis running inside the cluster with persistent storage.

- Deploy PostgreSQL via Bitnami Helm chart with a PersistentVolumeClaim
- Deploy Redis via Bitnami Helm chart
- Configure resource limits to fit within VPS memory budget (Postgres ~512MB, Redis ~128MB)
- Create a backup CronJob that dumps PostgreSQL to a file and keeps N rolling backups on disk
- **Deliverable:** `psql` into in-cluster Postgres, `redis-cli ping` returns PONG, data survives pod restarts

### MVP 6: Multi-Service Platform (ongoing)

**Goal:** Deploy the pattern trainer and task queue as services in the same cluster.

- Each project gets its own Helm chart and GitHub Actions workflow
- Shared Ingress routes traffic to each service by subdomain
- All services connect to the shared PostgreSQL (separate databases) and Redis instances
- All services report metrics to the same Prometheus/Grafana stack
- **Deliverable:** Unified platform running all portfolio services with centralized observability and a single deploy pipeline

---

