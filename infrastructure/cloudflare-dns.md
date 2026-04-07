# Cloudflare DNS Setup

## Domain
- Domain: charlesgatchalian.dev
- Nameservers: pointed to Cloudflare

## DNS Records

### charlesgatchalian.dev zone
| Type | Name | Content | Proxy |
|------|------|---------|-------|
| A | api | 143.198.54.6 | Proxied (orange cloud) |
| A | grafana | 143.198.54.6 | Proxied (orange cloud) |

### lcpatterns.dev zone (separate domain, separate Cloudflare zone)
| Type | Name | Content | Proxy |
|------|------|---------|-------|
| A | @ | 143.198.54.6 | Proxied (orange cloud) |
| A | www | 143.198.54.6 | Proxied (orange cloud) |

## SSL/TLS Settings (apply to each zone)
- Encryption mode: Flexible (Cloudflare terminates TLS, forwards to VPS on port 80)
- Always Use HTTPS: ON
- Automatic HTTPS Rewrites: ON
- Minimum TLS Version: 1.2

## Notes
- `platform.charlesgatchalian.dev` is hosted on Vercel — Vercel manages its own DNS record, no A record needed here
- Flexible mode means no cert-manager or Let's Encrypt needed inside the cluster
- Traffic between Cloudflare and the VPS travels over HTTP on port 80
- If upgrading to Full (Strict) mode later: install cert-manager and configure Traefik TLS
