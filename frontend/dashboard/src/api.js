const BASE = (import.meta.env.VITE_API_URL ?? '').replace(/\/$/, '')

export async function fetchDashboard(signal) {
  const response = await fetch(`${BASE}/api/v1/dashboard`, { signal })
  if (!response.ok) throw new Error(`Dashboard request failed: ${response.status}`)
  const data = await response.json()
  if (!Array.isArray(data.pipeline?.runs) || !data.service) throw new Error('Invalid dashboard response')
  return data
}
