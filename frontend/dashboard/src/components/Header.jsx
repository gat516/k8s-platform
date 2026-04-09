import { useState, useEffect } from 'react'

export default function Header({ health }) {
  const [clock, setClock] = useState('')

  useEffect(() => {
    const tick = () => {
      const now = new Date()
      setClock(now.toLocaleTimeString('en-GB', { hour12: false }) + ' UTC')
    }
    tick()
    const id = setInterval(tick, 1000)
    return () => clearInterval(id)
  }, [])

  const version = health?.version ?? '—'

  return (
    <header style={{
      padding: '16px 24px 14px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      borderBottom: '1px solid var(--border)',
      background: 'var(--surface)',
      marginBottom: '16px',
    }}>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: '12px' }}>
        <span style={{
          fontFamily: "'Bebas Neue', sans-serif",
          fontSize: '22px',
          letterSpacing: '0.12em',
          color: 'var(--amber)',
          textShadow: '0 0 20px var(--amber), 0 0 40px rgba(232,160,32,0.4)',
        }}>
          K8S // PLATFORM
        </span>
        <span style={{
          fontSize: '10px',
          color: 'var(--text-dim)',
          letterSpacing: '0.2em',
          textTransform: 'uppercase',
        }}>
          Portfolio Cluster Control
        </span>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '24px' }}>
        <div style={{
          display: 'flex', alignItems: 'center', gap: '6px',
          fontSize: '10px', letterSpacing: '0.15em',
          color: 'var(--green)', textTransform: 'uppercase',
        }}>
          <div style={{
            width: '6px', height: '6px', borderRadius: '50%',
            background: 'var(--green)',
            animation: 'pulse 2s ease-in-out infinite',
          }} />
          Live
        </div>
        <span style={{ color: 'var(--border)', fontSize: '10px' }}>|</span>
        <span style={{ fontSize: '10px', color: 'var(--text-dim)', letterSpacing: '.1em' }}>
          k3s · single-node · hetzner VPS
        </span>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '20px', fontSize: '10px', color: 'var(--text-dim)', letterSpacing: '0.08em' }}>
        <span>
          VER: <span style={{ color: 'var(--cyan)' }}>{version}</span>
        </span>
        <span>
          ENV: <span style={{ color: 'var(--amber)' }}>PRODUCTION</span>
        </span>
        <span style={{ fontSize: '11px', color: 'var(--text)', letterSpacing: '0.1em' }}>
          {clock}
        </span>
      </div>
    </header>
  )
}
