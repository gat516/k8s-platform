# k8s-platform

A personal Kubernetes platform running portfolio services on a single-node k3s cluster on a cheap VPS. The point is to demonstrate real DevOps work — automated deployments, Helm packaging, and live observability — not just a static portfolio site.

## Stack

- **API:** Go
- **Orchestration:** k3s (single-node, ~512MB overhead)
- **Ingress:** Traefik (bundled with k3s)
- **DNS/TLS:** Cloudflare free tier
- **CI/CD:** GitHub Actions (build → push to GHCR → deploy via Helm)
- **Monitoring:** Prometheus + Grafana (kube-prometheus-stack)
- **Databases:** PostgreSQL + Redis (Bitnami Helm charts, PVCs)
- **Infrastructure:** Hetzner or DigitalOcean VPS (~$6/mo, 4GB RAM)

## Status

k3s cluster is live on a DigitalOcean VPS. Currently working on **MVP 1**: first service deployed end-to-end.

## Roadmap

| MVP | Goal | Status |
|-----|------|--------|
| 1 | Single-node k3s cluster + one service deployed | in progress |
| 2 | GitHub Actions CI/CD pipeline | not started |
| 3 | DNS + TLS + Traefik Ingress | not started |
| 4 | Helm charts + Prometheus/Grafana | not started |
| 5 | In-cluster PostgreSQL + Redis with persistent storage | not started |
| 6 | Multi-service platform (all portfolio projects deployed) | not started |

See `docs/mvp.md` for detailed plans per phase and `docs/architecture.md` for the full system design.

## Project Structure

```
k8s-platform/
├── src/                   # Go API service
├── manifests/             # Kubernetes Deployment + Service
├── infrastructure/        # VPS bootstrap and kubeconfig scripts
├── .github/workflows/     # CI pipeline (test + build + push to GHCR)
└── docs/                  # Architecture and MVP planning docs
```
