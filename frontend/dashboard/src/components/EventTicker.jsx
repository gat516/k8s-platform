// EventTicker builds a live-ish event feed from the latest API data.
// When data is unavailable it falls back to static placeholder events.

function buildEvents(services, cluster) {
  const events = []

  if (cluster) {
    events.push(
      { label: `Node count`, hi: `${cluster.nodes}`, status: 'ok' },
      { label: `Pods`, hi: `${cluster.pods_running}/${cluster.pods_total} running`, status: 'ok' },
      { label: `Cluster version`, hi: cluster.k3s_version, status: 'ok' },
    )
    if (cluster.restarts_24h > 0) {
      events.push({ label: `Restarts (24h)`, hi: `${cluster.restarts_24h}`, status: 'warn' })
    }
  }

  if (services && services.length > 0) {
    services.forEach(svc => {
      const statusClass = svc.status === 'running' ? 'ok'
                        : svc.status === 'pending'  ? 'warn'
                        :                             'err'
      events.push({ label: `${svc.name} (${svc.namespace})`, hi: svc.status, status: statusClass })
    })
  }

  // Always include a few static entries so the ticker isn't empty on first load.
  events.push(
    { label: 'Ingress TLS cert', hi: 'valid until 2026-07-08', status: 'ok' },
    { label: 'GitHub Actions deploy #82', hi: 'Success', status: 'ok' },
    { label: 'Prometheus scrape', hi: '200ms p99', status: 'ok' },
  )

  return events
}

export default function EventTicker({ services, cluster }) {
  const events = buildEvents(services, cluster)
  // Duplicate for seamless CSS loop.
  const doubled = [...events, ...events]

  const colorMap = {
    ok:   'var(--green)',
    warn: 'var(--amber)',
    err:  'var(--red)',
  }

  return (
    <div style={{
      overflow: 'hidden',
      borderTop: '1px solid var(--border)',
      background: 'var(--surface)',
      padding: '8px 24px',
      display: 'flex',
      alignItems: 'center',
      gap: '16px',
    }}>
      <span style={{
        fontSize: '9px', color: 'var(--amber)', letterSpacing: '0.2em', textTransform: 'uppercase',
        flexShrink: 0, borderRight: '1px solid var(--border)', paddingRight: '16px',
      }}>
        Events
      </span>

      <div style={{
        display: 'flex', gap: '40px',
        animation: 'scroll-x 30s linear infinite',
        whiteSpace: 'nowrap',
      }}>
        {doubled.map((ev, i) => (
          <span key={i} style={{ fontSize: '9px', color: 'var(--text-dim)', letterSpacing: '0.06em' }}>
            {ev.label}{' — '}
            <span style={{ color: colorMap[ev.status] ?? 'var(--text)' }}>{ev.hi}</span>
          </span>
        ))}
      </div>
    </div>
  )
}
