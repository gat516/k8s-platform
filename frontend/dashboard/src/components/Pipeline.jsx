import { duration, timestamp, stageLabel } from '../format.js'

export default function Pipeline({ runs, loading, stale, unavailable }) {
  return <section className="history-panel" aria-labelledby="history-heading">
    <div className="panel-heading"><h2 id="history-heading">Recent runs</h2>{stale && <span className="detail">Stale data</span>}</div>
    {runs.length === 0 ? <p className="empty-state" role="status">{loading ? 'Loading runs…' : unavailable ? 'Workflow results unavailable.' : 'No runs yet.'}</p> :
      <div className="table-scroll"><table><thead><tr><th scope="col">Run / commit</th><th scope="col">Started</th><th scope="col">Duration</th><th scope="col">Result</th><th scope="col">Rollback</th><th scope="col"><span className="sr-only">Details</span></th></tr></thead><tbody>{runs.map(run => <tr key={run.id}><td><strong>#{run.number}</strong><code>{run.commit.slice(0, 7)}</code></td><td>{timestamp(run.started_at)}</td><td>{duration(run.duration_seconds)}</td><td><span className={`badge state-${run.status === 'completed' ? run.conclusion : run.status}`}>{stageLabel(run.status === 'completed' ? run.conclusion : run.status)}</span></td><td>{run.rollback === 'passed' ? 'Recovery verified' : stageLabel(run.rollback)}</td><td><a className="text-link" href={run.url} target="_blank" rel="noreferrer" aria-label={`View workflow run ${run.number}`}>View ↗</a></td></tr>)}</tbody></table></div>}
  </section>
}
