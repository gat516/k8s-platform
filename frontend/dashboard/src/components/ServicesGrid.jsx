function statusBadge(status) {
  switch (status) {
    case 'running': return { label: 'Running', color: 'var(--green)',  bg: 'rgba(57,217,138,0.08)',  border: 'rgba(57,217,138,0.2)' }
    case 'pending': return { label: 'Pending', color: 'var(--amber)',  bg: 'var(--amber-glow)',       border: 'var(--amber-dim)' }
    case 'error':   return { label: 'Error',   color: 'var(--red)',    bg: 'var(--red-glow)',         border: 'rgba(255,59,92,0.3)' }
    default:        return { label: status,    color: 'var(--text-dim)', bg: 'transparent',           border: 'var(--border)' }
  }
}

function PodDots({ ready, desired, status }) {
  const dots = []
  for (let i = 0; i < desired; i++) {
    let dotStatus = 'dead'
    if (i < ready) {
      dotStatus = 'ok'
    } else if (status === 'pending') {
      dotStatus = 'pending'
    }
    const bg = dotStatus === 'ok'      ? 'var(--green)'
             : dotStatus === 'pending' ? 'var(--amber)'
             :                          'var(--surface3)'
    dots.push(
      <div key={i} style={{
        width: '6px', height: '6px', borderRadius: '1px',
        background: bg,
        border: dotStatus === 'dead' ? '1px solid var(--border-hi)' : undefined,
        animation: dotStatus === 'pending' ? 'blink 1s step-end infinite' : undefined,
      }} />
    )
  }
  return <div style={{ display: 'flex', gap: '3px', marginTop: '10px' }}>{dots}</div>
}

function ServiceTile({ svc }) {
  const badge = statusBadge(svc.status)
  const accentColor = svc.status === 'running' ? 'var(--green)'
                    : svc.status === 'pending'  ? 'var(--amber)'
                    :                             'var(--red)'

  // Trim the image to keep it readable
  const imageShort = svc.image.length > 28 ? svc.image.slice(svc.image.lastIndexOf('/') + 1) : svc.image

  return (
    <div style={{
      background: 'var(--surface)',
      padding: '14px',
      position: 'relative',
      borderBottom: `2px solid ${accentColor}`,
      transition: 'background 0.2s',
    }}
      onMouseEnter={e => e.currentTarget.style.background = 'var(--surface2)'}
      onMouseLeave={e => e.currentTarget.style.background = 'var(--surface)'}
    >
      <div style={{ fontSize: '11px', fontWeight: 600, color: 'var(--text-hi)', letterSpacing: '0.04em', marginBottom: '4px' }}>
        {svc.name}
      </div>
      <div style={{ fontSize: '9px', color: 'var(--text-dim)', letterSpacing: '0.12em', marginBottom: '12px' }}>
        ns: {svc.namespace}
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
        {[
          { k: 'Image',    v: imageShort },
          { k: 'Replicas', v: `${svc.replicas_ready}/${svc.replicas_desired}` },
        ].map(({ k, v }) => (
          <div key={k} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '9px', color: 'var(--text-dim)' }}>
            <span>{k}</span>
            <span style={{ color: 'var(--text)', letterSpacing: '0.06em' }}>{v}</span>
          </div>
        ))}
      </div>
      <PodDots ready={svc.replicas_ready} desired={svc.replicas_desired} status={svc.status} />
      <div style={{ marginTop: '8px' }}>
        <span style={{
          fontSize: '9px', letterSpacing: '0.12em', textTransform: 'uppercase',
          padding: '2px 7px', borderRadius: '1px', fontWeight: 500,
          color: badge.color, background: badge.bg, border: `1px solid ${badge.border}`,
        }}>
          {badge.label}
        </span>
      </div>
    </div>
  )
}

function SkeletonTile() {
  return (
    <div style={{ background: 'var(--surface)', padding: '14px' }}>
      {[80, 50, 60, 40].map((w, i) => (
        <div key={i} style={{
          height: i === 0 ? '14px' : '10px',
          width: `${w}%`,
          background: 'var(--surface3)',
          borderRadius: '2px',
          marginBottom: '8px',
        }} />
      ))}
    </div>
  )
}

export default function ServicesGrid({ services }) {
  const running = services?.filter(s => s.status === 'running').length ?? 0
  const pending = services?.filter(s => s.status === 'pending').length ?? 0

  return (
    <div style={{
      gridColumn: '1 / 4',
      background: 'var(--surface)',
      border: '1px solid var(--border)',
      position: 'relative',
      overflow: 'hidden',
      animation: 'fadeUp 0.4s ease forwards 0.20s',
      opacity: 0,
    }}>
      <div style={{
        position: 'absolute', top: 0, left: 0, right: 0, height: '1px',
        background: 'linear-gradient(90deg, transparent, var(--border-hi), transparent)',
      }} />

      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        padding: '10px 14px 9px', borderBottom: '1px solid var(--border)',
      }}>
        <span style={{ fontSize: '9px', letterSpacing: '0.2em', textTransform: 'uppercase', color: 'var(--text-dim)', fontWeight: 500 }}>
          Deployed Services
        </span>
        <span style={{ fontSize: '9px', color: 'var(--text-dim)' }}>
          <span style={{ color: 'var(--green)' }}>{running}</span> running
          {' · '}
          <span style={{ color: 'var(--amber)' }}>{pending}</span> pending
        </span>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '1px', background: 'var(--border)' }}>
        {services == null
          ? Array.from({ length: 8 }, (_, i) => <SkeletonTile key={i} />)
          : services.map(svc => <ServiceTile key={`${svc.namespace}/${svc.name}`} svc={svc} />)
        }
      </div>
    </div>
  )
}
