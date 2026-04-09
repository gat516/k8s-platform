// Circumference of the SVG ring: 2π * 44 ≈ 276.46
const C = 276.46

function RingGauge({ label, value, color, colorVar, glowColor, history }) {
  const pct     = value ?? 0
  const offset  = C * (1 - pct / 100)
  const numStr  = value != null ? Math.round(pct) : '—'

  const maxH = history.length > 0 ? Math.max(...history, 1) : 1

  return (
    <div style={{
      background: 'var(--surface)', padding: '18px 14px',
      display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '10px',
    }}>
      <span style={{ fontSize: '9px', letterSpacing: '0.2em', textTransform: 'uppercase', color: 'var(--text-dim)' }}>
        {label}
      </span>

      <div style={{ position: 'relative', width: '110px', height: '110px' }}>
        <svg width="110" height="110" viewBox="0 0 110 110">
          <circle cx="55" cy="55" r="44" fill="none" stroke="var(--surface3)" strokeWidth="8" />
          <circle
            cx="55" cy="55" r="44" fill="none"
            stroke={colorVar} strokeWidth="8"
            strokeDasharray={C}
            strokeDashoffset={offset}
            strokeLinecap="round"
            transform="rotate(-90 55 55)"
            style={{
              transition: 'stroke-dashoffset 1.5s cubic-bezier(0.4,0,0.2,1)',
              filter: `drop-shadow(0 0 6px ${glowColor})`,
            }}
          />
        </svg>
        <div style={{
          position: 'absolute', inset: 0,
          display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
        }}>
          <span style={{
            fontFamily: "'Bebas Neue', sans-serif",
            fontSize: '30px', letterSpacing: '0.04em', lineHeight: 1,
            color,
          }}>
            {numStr}
          </span>
          <span style={{ fontSize: '9px', color: 'var(--text-dim)', letterSpacing: '0.1em' }}>percent</span>
        </div>
      </div>

      {/* Sparkline */}
      <div style={{ width: '100%', display: 'flex', alignItems: 'flex-end', gap: '2px', height: '24px' }}>
        {history.length === 0
          ? Array.from({ length: 12 }, (_, i) => (
              <div key={i} style={{ flex: 1, height: '3px', borderRadius: '1px 1px 0 0', background: 'var(--surface3)', minWidth: '3px' }} />
            ))
          : history.map((v, i) => (
              <div key={i} style={{
                flex: 1,
                height: `${Math.max(3, Math.round((v / maxH) * 24))}px`,
                borderRadius: '1px 1px 0 0',
                background: colorVar,
                opacity: 0.5,
                minWidth: '3px',
                transition: 'height 0.5s ease',
              }} />
            ))
        }
      </div>
    </div>
  )
}

export default function ResourceGauges({ resources, cpuHistory, memHistory, diskHistory, cluster }) {
  const netIn    = resources?.net_in_mbps
  const netOut   = resources?.net_out_mbps

  // Derive namespace count from cluster data (not directly available, hard-code
  // label as in the original static design).
  const namespaceCount = 3

  return (
    <div style={{
      gridColumn: '2 / 4',
      background: 'var(--surface)',
      border: '1px solid var(--border)',
      position: 'relative',
      overflow: 'hidden',
      animation: 'fadeUp 0.4s ease forwards 0.12s',
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
          Resource Metrics
        </span>
        <span style={{ fontSize: '9px', color: 'var(--text-dim)', letterSpacing: '.1em' }}>Last 5m avg</span>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '1px', background: 'var(--border)' }}>
        <RingGauge
          label="CPU Usage"
          value={resources?.cpu_percent}
          color="var(--amber)"
          colorVar="var(--amber)"
          glowColor="rgba(232,160,32,0.5)"
          history={cpuHistory}
        />
        <RingGauge
          label="Memory"
          value={resources?.memory_percent}
          color="var(--cyan)"
          colorVar="var(--cyan)"
          glowColor="rgba(0,217,184,0.4)"
          history={memHistory}
        />
        <RingGauge
          label="Disk I/O"
          value={resources?.disk_io_percent}
          color="var(--green)"
          colorVar="var(--green)"
          glowColor="rgba(57,217,138,0.4)"
          history={diskHistory}
        />
      </div>

      {/* Network + namespace row */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '1px', background: 'var(--border)', borderTop: '1px solid var(--border)' }}>
        <NetStat label="Net In"     value={netIn}    unit="MB/s avg" color="var(--cyan)"  glow="rgba(0,217,184,0.3)" />
        <NetStat label="Net Out"    value={netOut}   unit="MB/s avg" color="var(--amber)" glow="rgba(232,160,32,0.3)" />
        <NetStat label="Namespaces" value={namespaceCount} unit="default · monitoring · kube-system" color="var(--text-hi)" />
      </div>
    </div>
  )
}

function NetStat({ label, value, unit, color, glow }) {
  const display = value != null ? (typeof value === 'number' && !Number.isInteger(value) ? value.toFixed(1) : value) : '—'
  return (
    <div style={{ background: 'var(--surface)', padding: '10px 14px' }}>
      <div style={{ fontSize: '9px', letterSpacing: '0.18em', textTransform: 'uppercase', color: 'var(--text-dim)', marginBottom: '2px' }}>
        {label}
      </div>
      <div style={{
        fontFamily: "'Bebas Neue', sans-serif", fontSize: '20px',
        color, textShadow: glow ? `0 0 8px ${glow}` : undefined,
      }}>
        {display}
      </div>
      <div style={{ fontSize: '10px', color: 'var(--text-dim)', marginTop: '2px', letterSpacing: '0.08em' }}>
        {unit}
      </div>
    </div>
  )
}
