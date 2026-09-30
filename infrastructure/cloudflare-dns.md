# Dashboard DNS on qireadr.com

The dashboard’s hostname is **status.qireadr.com**. The retired VPS and
its domains are no longer deployment targets.

Follow [the dashboard deployment guide](../docs/domain.md) for the current setup:

- Reuse qireadr’s existing AWS/k3s host; confirm its Elastic IP before changing DNS.
- Add an **A** record named `status`, pointing to that address.
- Start **DNS only** while Traefik obtains and verifies the origin certificate.
- Then enable **Proxied**, retaining **Full (strict)** and the existing ACME path.
- Preserve the apex record, existing Worker route, and unrelated DNS records.

The initial deployment and HTTPS were verified on September 29, 2026. The `status`
record is Proxied and the origin certificate is valid; no DNS change is needed.
The old `k3s-setup.sh` and `traefik-config.yaml` describe the retired cluster; they
must not be applied to qireadr’s shared host.
