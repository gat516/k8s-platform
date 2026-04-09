// Static pipeline data derived from recent git history shown in the original
// static dashboard. This section intentionally stays static — wiring it to a
// live GitHub Actions API would require a token and is out of scope.
const RUNS = [
  { num: 82, sha: 'eab0a2a', msg: 'pivoting to hosting vercel for frontend', actor: 'cj', dur: '2m 14s', age: '3h ago',  status: 'ok' },
  { num: 81, sha: 'a545496', msg: 'feat: add CD deploy job to CI workflow',   actor: 'cj', dur: '3m 02s', age: '1d ago',  status: 'ok' },
  { num: 80, sha: 'd930adc', msg: 'added setup to readme',                    actor: 'cj', dur: '1m 48s', age: '5d ago',  status: 'ok' },
  { num: 79, sha: '6cad6e3', msg: 'updated readme',                           actor: 'cj', dur: '1m 55s', age: '6d ago',  status: 'ok' },
  { num: 78, sha: '2cda8ab', msg: 'add kubeconfig helper script',             actor: 'cj', dur: '4m 31s', age: '8d ago',  status: 'warn' },
]

function StatusDot({ status }) {
  const map = {
    ok:      { bg: 'var(--green)', shadow: '0 0 6px rgba(57,217,138,0.6)' },
    running: { bg: 'var(--cyan)',  shadow: undefined, animation: 'pulse-cyan 1.5s ease-in-out infinite' },
    warn:    { bg: 'var(--amber)', shadow: undefined },
    error:   { bg: 'var(--red)',   shadow: '0 0 6px rgba(255,59,92,0.5)' },
  }
  const s = map[status] ?? map.warn
  return (
    <div style={{
      width: '8px', height: '8px', borderRadius: '50%', flexShrink: 0,
      background: s.bg, boxShadow: s.shadow, animation: s.animation,
    }} />
  )
}

const COL = '26px 140px 90px 1fr 90px 90px 60px'

export default function Pipeline() {
  return (
    <div style={{
      gridColumn: '1 / 4',
      background: 'var(--surface)',
      border: '1px solid var(--border)',
      position: 'relative',
      overflow: 'hidden',
      animation: 'fadeUp 0.4s ease forwards 0.30s',
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
          CI/CD Pipeline · GitHub Actions
        </span>
        <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
          <span style={{ fontSize: '9px', color: 'var(--text-dim)' }}>main ·</span>
          <span style={{
            fontSize: '9px', letterSpacing: '0.12em', textTransform: 'uppercase',
            padding: '2px 7px', borderRadius: '1px', fontWeight: 500,
            color: 'var(--green)', background: 'rgba(57,217,138,0.08)', border: '1px solid rgba(57,217,138,0.2)',
          }}>
            Last: success
          </span>
        </div>
      </div>

      {/* Header row */}
      <div style={{
        display: 'grid', gridTemplateColumns: COL,
        alignItems: 'center', padding: '10px 14px',
        borderBottom: '1px solid var(--border)', gap: '12px',
        fontSize: '9px', letterSpacing: '0.15em', textTransform: 'uppercase', color: 'var(--text-dim)',
      }}>
        <div />
        <div>Branch / Run</div>
        <div>Commit</div>
        <div>Message</div>
        <div>Actor</div>
        <div>Duration</div>
        <div style={{ textAlign: 'right' }}>Age</div>
      </div>

      {RUNS.map((run, idx) => (
        <div key={run.num} style={{
          display: 'grid', gridTemplateColumns: COL,
          alignItems: 'center', padding: '9px 14px',
          borderBottom: idx < RUNS.length - 1 ? '1px solid var(--border)' : undefined,
          gap: '12px', fontSize: '10px',
          transition: 'background 0.15s',
          cursor: 'default',
        }}
          onMouseEnter={e => e.currentTarget.style.background = 'var(--surface2)'}
          onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
        >
          <div><StatusDot status={run.status} /></div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px', color: 'var(--text)' }}>
            <span style={{ color: 'var(--cyan-dim)', fontSize: '10px' }}>⎇</span>
            {' '}main #{run.num}
          </div>
          <div style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '9px', color: 'var(--cyan)', letterSpacing: '0.06em' }}>
            {run.sha}
          </div>
          <div style={{ color: 'var(--text)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
            {run.msg}
          </div>
          <div style={{ color: 'var(--text-dim)', fontSize: '9px' }}>{run.actor}</div>
          <div style={{ color: 'var(--text-dim)', fontSize: '9px' }}>{run.dur}</div>
          <div style={{ color: 'var(--text-dim)', fontSize: '9px', textAlign: 'right' }}>{run.age}</div>
        </div>
      ))}
    </div>
  )
}
