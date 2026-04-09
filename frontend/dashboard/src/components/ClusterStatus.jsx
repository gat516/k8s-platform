// Skeleton placeholder used while data is loading.
const Skeleton = () => (
  <span style={{
    display: 'inline-block',
    width: '40px', height: '1em',
    background: 'var(--surface3)',
    borderRadius: '2px',
    verticalAlign: 'middle',
  }} />
)

function Stat({ label, value, unit, colorClass }) {
  const colorMap = {
    amber: { color: 'var(--amber)', textShadow: '0 0 12px rgba(232,160,32,0.5)' },
    cyan:  { color: 'var(--cyan)',  textShadow: '0 0 12px rgba(0,217,184,0.4)' },
    green: { color: 'var(--green)', textShadow: '0 0 12px rgba(57,217,138,0.4)' },
    red:   { color: 'var(--red)',   textShadow: '0 0 12px rgba(255,59,92,0.4)' },
  }
  const style = colorMap[colorClass] ?? {}

  return (
    <div style={{ background: 'var(--surface)', padding: '14px' }}>
      <div style={{ fontSize: '9px', letterSpacing: '0.18em', textTransform: 'uppercase', color: 'var(--text-dim)', marginBottom: '6px' }}>
        {label}
      </div>
      <div style={{
        fontFamily: "'Bebas Neue', sans-serif",
        fontSize: '28px', letterSpacing: '0.04em',
        color: 'var(--text-hi)', lineHeight: 1,
        ...style,
      }}>
        {value ?? <Skeleton />}
      </div>
      {unit && (
        <div style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '10px', color: 'var(--text-dim)', marginTop: '2px', letterSpacing: '0.08em' }}>
          {unit}
        </div>
      )}
    </div>
  )
}

function formatUptime(seconds) {
  if (seconds == null) return null
  const days = Math.floor(seconds / 86400)
  if (days > 0) return `${days}d`
  const hours = Math.floor(seconds / 3600)
  return `${hours}h`
}

export default function ClusterStatus({ cluster }) {
  const nodes    = cluster?.nodes
  const uptime   = formatUptime(cluster?.uptime_seconds)
  const version  = cluster?.k3s_version
  const running  = cluster?.pods_running
  const total    = cluster?.pods_total
  const restarts = cluster?.restarts_24h

  const restartColor = restarts > 0 ? 'red' : undefined

  return (
    <div style={{
      gridColumn: '1 / 2',
      background: 'var(--surface)',
      border: '1px solid var(--border)',
      position: 'relative',
      overflow: 'hidden',
      animation: 'fadeUp 0.4s ease forwards 0.05s',
      opacity: 0,
    }}>
      {/* top highlight line */}
      <div style={{
        position: 'absolute', top: 0, left: 0, right: 0, height: '1px',
        background: 'linear-gradient(90deg, transparent, var(--border-hi), transparent)',
      }} />

      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        padding: '10px 14px 9px', borderBottom: '1px solid var(--border)',
      }}>
        <span style={{ fontSize: '9px', letterSpacing: '0.2em', textTransform: 'uppercase', color: 'var(--text-dim)', fontWeight: 500 }}>
          Cluster Status
        </span>
        <span style={{
          fontSize: '9px', letterSpacing: '0.12em', textTransform: 'uppercase',
          padding: '2px 7px', borderRadius: '1px', fontWeight: 500,
          color: 'var(--green)', background: 'rgba(57,217,138,0.08)',
          border: '1px solid rgba(57,217,138,0.2)',
        }}>
          Healthy
        </span>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1px', background: 'var(--border)' }}>
        <Stat label="Nodes"         value={nodes}    colorClass="amber" unit={version ? `Ready · ${version}` : 'Ready'} />
        <Stat label="Uptime"        value={uptime}   colorClass="cyan"  unit="since cluster start" />
        <Stat label="Pods Running"  value={running != null ? `${running}` : null} colorClass="green" unit={total != null ? `of ${total} scheduled` : null} />
        <Stat label="Restarts (24h)" value={restarts != null ? `${restarts}` : null} colorClass={restartColor} unit={restarts === 0 ? 'No crashloops' : 'check logs'} />

        {/* Node info footer */}
        <div style={{ gridColumn: '1/3', background: 'var(--surface)', padding: '14px' }}>
          <div style={{ fontSize: '9px', color: 'var(--text-dim)', marginBottom: '8px', letterSpacing: '0.18em', textTransform: 'uppercase' }}>
            Node: vps-node-01 · 4 vCPU · 8 GB RAM
          </div>
          <div style={{ display: 'flex', gap: '16px' }}>
            {[
              { label: 'KERNEL',  value: '6.1.0-27-amd64', color: 'var(--text)' },
              { label: 'REGION',  value: 'EU-CENTRAL-1',   color: 'var(--text)' },
              { label: 'TRAEFIK', value: 'v2.11.0',        color: 'var(--green)' },
              { label: 'DNS/TLS', value: 'Cloudflare',     color: 'var(--green)' },
            ].map(({ label, value, color }) => (
              <div key={label}>
                <div style={{ fontSize: '9px', color: 'var(--text-dim)', marginBottom: '3px' }}>{label}</div>
                <div style={{ fontSize: '10px', color, letterSpacing: '.04em' }}>{value}</div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
