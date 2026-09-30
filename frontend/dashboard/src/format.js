export function duration(seconds) {
  if (seconds == null || !Number.isFinite(seconds)) return '—'
  return `${Math.floor(seconds / 60)}m ${Math.floor(seconds % 60)}s`
}
export function timestamp(value) {
  if (!value || !Number.isFinite(Date.parse(value))) return '—'
  return new Date(value).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}
export function stageLabel(state) {
  return ({ passed: 'Passed', success: 'Passed', failed: 'Failed', failure: 'Failed', timed_out: 'Timed out', cancelled: 'Cancelled', skipped: 'Not run', not_configured: 'Not configured', unverified: 'Not verified', running: 'Running', in_progress: 'Running', queued: 'Queued', pending: 'Pending', waiting: 'Waiting', action_required: 'Action required', neutral: 'Neutral', unknown: 'Unknown' })[state] ?? 'Unknown'
}
