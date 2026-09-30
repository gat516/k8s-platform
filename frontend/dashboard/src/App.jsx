import { useEffect, useState } from 'react'
import { fetchDashboard } from './api.js'
import Header from './components/Header.jsx'
import Pipeline from './components/Pipeline.jsx'
import { duration, timestamp, stageLabel } from './format.js'

const stages = [['Checks', 'checks'], ['Build', 'build'], ['Deploy', 'deploy'], ['Verify', 'verify'], ['Rollback', 'rollback']]

export default function App() {
  const [data, setData] = useState(null)
  const [error, setError] = useState(false)
  const [refresh, setRefresh] = useState(0)
  const [loading, setLoading] = useState(true)
  useEffect(() => {
    let stopped = false
    let timer
    let controller
    const load = async () => {
      controller = new AbortController()
      setLoading(true)
      try {
        const result = await fetchDashboard(controller.signal)
        if (!stopped) { setData(result); setError(false) }
      } catch (err) {
        if (!stopped && err.name !== 'AbortError') setError(true)
      } finally {
        if (!stopped) { setLoading(false); timer = setTimeout(load, 30000) }
      }
    }
    load()
    return () => { stopped = true; clearTimeout(timer); controller?.abort() }
  }, [refresh])

  const latest = data?.pipeline.runs[0]
  const stale = error || data?.pipeline.state === 'stale'
  const health = error ? 'unknown' : data?.service.state ?? 'unknown'
  const healthLabel = ({ healthy: 'Responding', unhealthy: 'Check failed', unknown: 'Unknown', not_configured: 'Not configured' })[health] ?? 'Unknown'

  return (
    <div className="app-shell">
      <Header application={data?.application ?? 'qireadr'} />
      <main>
        <div className="page-heading">
          <h1>Status</h1>
          <button className="refresh-button" disabled={loading} onClick={() => setRefresh(v => v + 1)}>{loading ? 'Checking…' : 'Refresh'}</button>
        </div>
        {(error || data?.pipeline.message) && <div className="notice" role="status">{error ? 'Status unavailable. Displayed results may be out of date.' : data.pipeline.message}</div>}
        <section className="service-summary" aria-label="Service health">
          <div className="service-line"><h2>{data?.application ?? 'qireadr'}</h2><span className={`health health-${health}`} role="status">{!data && loading ? 'Checking…' : healthLabel}</span></div>
          <dl className="service-details"><div><dt>Checked</dt><dd>{timestamp(data?.service.checked_at)}</dd></div><div><dt>Version</dt><dd><code>{data?.service.version || 'Not reported'}</code></dd></div></dl>
          {data?.service.message && <p className="detail">{data.service.message}</p>}
        </section>
        {latest && <section className="release-panel" aria-labelledby="release-heading">
          <div className="panel-heading"><h2 id="release-heading">Latest workflow</h2><a className="text-link" href={latest.url} target="_blank" rel="noreferrer">Run #{latest.number} ↗</a></div>
          <div className="run-summary"><code>{latest.commit.slice(0, 7)}</code><span className={`badge state-${latest.status === 'completed' ? latest.conclusion : latest.status}`}>{stageLabel(latest.status === 'completed' ? latest.conclusion : latest.status)}</span><span className="detail">{duration(latest.duration_seconds)}</span>{stale && <span className="detail">Stale data</span>}</div>
          <div className="stage-grid">{stages.map(([name, key]) => <div className="stage" key={key}><h3>{name}</h3><span className={`badge state-${latest[key] ?? 'unknown'}`}>{stageLabel(latest[key])}</span></div>)}</div>
        </section>}
        <Pipeline runs={data?.pipeline.runs ?? []} loading={!data && loading} stale={stale} unavailable={error || (!!data && data.pipeline.state !== 'available')} />
        <footer>GitHub updated {timestamp(data?.pipeline.fetched_at)}</footer>
      </main>
    </div>
  )
}
